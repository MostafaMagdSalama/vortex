package parallel

import (
	"context"
	"iter"
	"sync"
)

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

// WorkerPoolMap uses a fixed pool of workers to process items.
func WorkerPoolMap[T, U any](ctx context.Context, seq iter.Seq[T], fn func(T) U, workers int) iter.Seq[U] {
	return func(yield func(U) bool) {
		if ctx.Err() != nil {
			return
		}

		tasks := make(chan T, workers*2)
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
					v, ok := <-tasks
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
			defer close(results)
			wg.Wait()
		}()

		go func() {
			defer close(tasks)
			for v := range seq {
				if ctx.Err() != nil {
					return
				}
				select {
				case tasks <- v:
				case <-ctx.Done():
					return
				}
			}
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
