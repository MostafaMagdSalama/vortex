package parallel

import (
	"iter"
	"sync"

	"github.com/MostafaMagdSalama/vortex/internal/sched"
)

// ParallelMap processes each element concurrently with n workers.
// Output order is NOT guaranteed — results arrive as workers finish.
func ParallelMap[T, U any](seq iter.Seq[T], fn func(T) U, workers int) iter.Seq[U] {
	return func(yield func(U) bool) {
		s := sched.New(workers)
		results := make(chan U, workers*2)
		var mu sync.Mutex
		stopped := false

		// producer: submit all items as tasks
		go func() {
			for v := range seq {
				v := v
				s.Submit(func() {
					result := fn(v)
					mu.Lock()
					if !stopped {
						results <- result
					}
					mu.Unlock()
				})
			}
			s.Stop()       // wait for all tasks
			close(results) // signal consumer we're done
		}()

		// consumer: yield results to caller
		for result := range results {
			if !yield(result) {
				mu.Lock()
				stopped = true
				mu.Unlock()
				return
			}
		}
	}
}

// BatchMap groups items into batches of size n, processes each batch concurrently.
// Useful when fn has a fixed overhead — better to amortize it across a batch.
func BatchMap[T, U any](seq iter.Seq[T], fn func([]T) []U, batchSize int) iter.Seq[U] {
	return func(yield func(U) bool) {
		var batch []T

		flush := func() bool {
			if len(batch) == 0 {
				return true
			}
			results := fn(batch)
			batch = batch[:0] // reset but keep memory
			for _, r := range results {
				if !yield(r) {
					return false
				}
			}
			return true
		}

		for v := range seq {
			batch = append(batch, v)
			if len(batch) >= batchSize {
				if !flush() {
					return
				}
			}
		}
		flush() // process remaining items
	}
}

// WorkerPoolMap uses a fixed pool of workers to process items.
// Unlike ParallelMap, the pool is reused across calls — better for long-running streams.
func WorkerPoolMap[T, U any](seq iter.Seq[T], fn func(T) U, workers int) iter.Seq[U] {
	return func(yield func(U) bool) {
		tasks := make(chan T, workers*2)
		results := make(chan U, workers*2)
		var wg sync.WaitGroup

		// start fixed worker goroutines
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for v := range tasks {
					results <- fn(v)
				}
			}()
		}

		// close results when all workers are done
		go func() {
			wg.Wait()
			close(results)
		}()

		// feed items into tasks channel
		go func() {
			for v := range seq {
				tasks <- v
			}
			close(tasks) // signal workers no more work
		}()

		// yield results to caller
		for result := range results {
			if !yield(result) {
				return
			}
		}
	}
}