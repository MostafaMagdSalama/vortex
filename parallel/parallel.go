package parallel

import (
	"context"
	"iter"
	"sync"
)

type task[T any] struct {
    index int
    value T
}

type result[U any] struct {
    index int
    value U
}

// ParallelMap processes each element concurrently with n workers.
func ParallelMap[T, U any](ctx context.Context, seq iter.Seq[T], fn func(T) U, workers int) iter.Seq[U] {
	return func(yield func(U) bool) {
		if ctx.Err() != nil {
			return
		}

		jobs := make(chan T, workers)
		results := make(chan U, workers*2)
		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					if ctx.Err() != nil {
						return
					}
					v, ok := <-jobs
					if !ok {
						return
					}
					select {
					case results <- fn(v):
					case <-ctx.Done():
						return
					}
				}
			}()
		}

		go func() {
			defer close(jobs)
			for v := range seq {
				if ctx.Err() != nil {
					return
				}
				select {
				case jobs <- v:
				case <-ctx.Done():
					return
				}
			}
		}()

		go func() {
			wg.Wait()
			close(results)
		}()

		for {
			if ctx.Err() != nil {
				return
			}
			result, ok := <-results
			if !ok {
				return
			}
			if !yield(result) {
				cancel()
				return
			}
		}
	}
}

// BatchMap groups items into batches of size n and processes each batch.
func BatchMap[T, U any](ctx context.Context, seq iter.Seq[T], fn func([]T) []U, batchSize int) iter.Seq[U] {
	return func(yield func(U) bool) {
		var batch []T

		flush := func() bool {
			if len(batch) == 0 {
				return true
			}
			if ctx.Err() != nil {
				return false
			}
			results := fn(batch)
			batch = batch[:0]
			for _, r := range results {
				if ctx.Err() != nil {
					return false
				}
				if !yield(r) {
					return false
				}
			}
			return true
		}

		for v := range seq {
			if ctx.Err() != nil {
				return
			}
			batch = append(batch, v)
			if len(batch) >= batchSize {
				if !flush() {
					return
				}
			}
		}

		flush()
	}
}

// OrderedParallelMap applies fn to each element of seq concurrently using
// workers goroutines and yields results in the original input order.
func OrderedParallelMap[T, U any](
	ctx context.Context,
	seq iter.Seq[T],
	fn func(T) U,
	workers int,
) iter.Seq[U] {

	return func(yield func(U) bool) {

		if ctx.Err() != nil {
			return
		}

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		tasks := make(chan task[T], workers*2)
		results := make(chan result[U], workers*2)

		var wg sync.WaitGroup

		// workers
		for i := 0; i < workers; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for {
					select {
					case <-ctx.Done():
						return

					case t, ok := <-tasks:
						if !ok {
							return
						}

						r := fn(t.value)

						select {
						case results <- result[U]{t.index, r}:
						case <-ctx.Done():
							return
						}
					}
				}
			}()
		}

		// producer
		go func() {
			defer close(tasks)

			i := 0
			for v := range seq {

				select {
				case tasks <- task[T]{i, v}:
					i++

				case <-ctx.Done():
					return
				}
			}
		}()

		// close results when workers done
		go func() {
			wg.Wait()
			close(results)
		}()

		// ordered buffer
		next := 0
		buffer := map[int]U{}

		for {

			select {
			case <-ctx.Done():
				return

			case r, ok := <-results:

				if !ok {
					return
				}

				buffer[r.index] = r.value

				for {
					v, exists := buffer[next]
					if !exists {
						break
					}

					delete(buffer, next)

					if !yield(v) {
						cancel()
						return
					}

					next++
				}
			}
		}
	}
}