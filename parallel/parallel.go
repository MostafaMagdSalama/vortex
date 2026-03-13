package parallel

import (
	"context"
	"iter"
	"sync"

)

// ParallelMap processes each element concurrently with n workers.
// Output order is NOT guaranteed — results arrive as workers finish.
func ParallelMap[T, U any](seq iter.Seq[T], fn func(T) U, workers int) iter.Seq[U] {
	return func(yield func(U) bool) {
		jobs := make(chan T, workers)
		results := make(chan U, workers*2)
		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// start fixed workers
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for v := range jobs {
					select {
					case results <- fn(v):  // send result
					case <-ctx.Done():       // caller stopped — exit cleanly
						return
					}
				}
			}()
		}

		// producer
		go func() {
			defer close(jobs)
			for v := range seq {
				select {
				case jobs <- v:       // send job
				case <-ctx.Done():    // caller stopped — stop feeding jobs
					return
				}
			}
		}()

		// close results when all workers finish
		go func() {
			wg.Wait()
			close(results)
		}()

		// consumer
		for result := range results {
			if !yield(result) {
				cancel() // signal all goroutines to stop
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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// start fixed worker goroutines
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for v := range tasks {
					select {
					case results <- fn(v): // send result
					case <-ctx.Done():      // caller stopped — exit cleanly
						return
					}
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
			defer close(tasks)
			for v := range seq {
				select {
				case tasks <- v:    // send task
				case <-ctx.Done(): // caller stopped — stop feeding
					return
				}
			}
		}()

		// yield results to caller
		for result := range results {
			if !yield(result) {
				cancel() // signal all goroutines to stop
				return
			}
		}
	}
}