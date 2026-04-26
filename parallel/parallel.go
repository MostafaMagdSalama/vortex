package parallel

import (
	"context"
	"iter"
	"sync"

	"github.com/MostafaMagdSalama/vortex"
)

type task[T any] struct {
	index int
	value T
}

type result[U any] struct {
	index int
	value U
	err   error
}

// ParallelMapSeq processes each element concurrently with n workers.
func ParallelMapSeq[T, U any](ctx context.Context, seq iter.Seq[T], fn func(T) U, workers int) iter.Seq[U] {
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
					select {
					case <-ctx.Done():
						return
					case v, ok := <-jobs:
						if !ok {
							return
						}
						select {
						case results <- fn(v):
						case <-ctx.Done():
							return
						}
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

// ParallelMap processes each element concurrently with n workers.
// Errors from the underlying sequence are passed through untouched.
func ParallelMap[T, U any](ctx context.Context, seq iter.Seq2[T, error], fn func(T) U, workers int) iter.Seq2[U, error] {
	return func(yield func(U, error) bool) {
		if ctx.Err() != nil {
			var zero U
			yield(zero, vortex.Wrap("parallel.ParallelMap", ctx.Err()))
			return
		}

		jobs := make(chan T, workers)
		results := make(chan result[U], workers*2)
		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case v, ok := <-jobs:
						if !ok {
							return
						}
						select {
						case results <- result[U]{value: fn(v)}:
						case <-ctx.Done():
							return
						}
					}
				}
			}()
		}

		go func() {
			defer close(jobs)
			for v, err := range seq {
				if ctx.Err() != nil {
					return
				}
				if err != nil {
					select {
					case results <- result[U]{err: vortex.Wrap("parallel.ParallelMap", err)}:
					case <-ctx.Done():
					}
					continue
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
				var zero U
				yield(zero, vortex.WrapCancelled("parallel.ParallelMap"))
				return
			}
			r, ok := <-results
			if !ok {
				return
			}
			if r.err != nil {
				var zero U
				if !yield(zero, r.err) {
					cancel()
					return
				}
				continue
			}
			if !yield(r.value, nil) {
				cancel()
				return
			}
		}
	}
}

// BatchMapSeq groups items into batches of size n and processes each batch.
func BatchMapSeq[T, U any](ctx context.Context, seq iter.Seq[T], fn func([]T) []U, batchSize int) iter.Seq[U] {
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

// BatchMap groups items into batches of size n and processes each batch.
// Errors from the underlying sequence are yielded inline and do not enter batches.
func BatchMap[T, U any](ctx context.Context, seq iter.Seq2[T, error], fn func([]T) []U, batchSize int) iter.Seq2[U, error] {
	return func(yield func(U, error) bool) {
		var batch []T

		flush := func() bool {
			if len(batch) == 0 {
				return true
			}
			if ctx.Err() != nil {
				var zero U
				yield(zero, vortex.WrapCancelled("parallel.BatchMap"))
				return false
			}
			results := fn(batch)
			batch = batch[:0]
			for _, r := range results {
				if ctx.Err() != nil {
					var zero U
					yield(zero, vortex.WrapCancelled("parallel.BatchMap"))
					return false
				}
				if !yield(r, nil) {
					return false
				}
			}
			return true
		}

		for v, err := range seq {
			if ctx.Err() != nil {
				var zero U
				yield(zero, vortex.WrapCancelled("parallel.BatchMap"))
				return
			}
			if err != nil {
				if !flush() {
					return
				}
				var zero U
				if !yield(zero, vortex.Wrap("parallel.BatchMap", err)) {
					return
				}
				continue
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

// OrderedParallelMapSeq applies fn to each element of seq concurrently using
// workers goroutines and yields results in the original input order.
func OrderedParallelMapSeq[T, U any](
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
						case results <- result[U]{index: t.index, value: r}:
						case <-ctx.Done():
							return
						}
					}
				}
			}()
		}

		go func() {
			defer close(tasks)

			i := 0
			for v := range seq {
				select {
				case tasks <- task[T]{index: i, value: v}:
					i++
				case <-ctx.Done():
					return
				}
			}
		}()

		go func() {
			wg.Wait()
			close(results)
		}()

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

// OrderedParallelMap applies fn to each element of seq concurrently using
// worker goroutines and yields results and errors in the original input order.
func OrderedParallelMap[T, U any](
	ctx context.Context,
	seq iter.Seq2[T, error],
	fn func(T) U,
	workers int,
) iter.Seq2[U, error] {
	return func(yield func(U, error) bool) {
		if ctx.Err() != nil {
			var zero U
			yield(zero, vortex.Wrap("parallel.OrderedParallelMap", ctx.Err()))
			return
		}

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		tasks := make(chan task[T], workers*2)
		results := make(chan result[U], workers*2)

		var wg sync.WaitGroup

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
						case results <- result[U]{index: t.index, value: r}:
						case <-ctx.Done():
							return
						}
					}
				}
			}()
		}

		go func() {
			defer close(tasks)

			i := 0
			for v, err := range seq {
				if ctx.Err() != nil {
					return
				}
				if err != nil {
					select {
					case results <- result[U]{index: i, err: vortex.Wrap("parallel.OrderedParallelMap", err)}:
						i++
					case <-ctx.Done():
						return
					}
					continue
				}

				select {
				case tasks <- task[T]{index: i, value: v}:
					i++
				case <-ctx.Done():
					return
				}
			}
		}()

		go func() {
			wg.Wait()
			close(results)
		}()

		next := 0
		buffer := map[int]result[U]{}

		for {
			select {
			case <-ctx.Done():
				var zero U
				yield(zero, vortex.WrapCancelled("parallel.OrderedParallelMap"))
				return
			case r, ok := <-results:
				if !ok {
					return
				}

				buffer[r.index] = r

				for {
					buffered, exists := buffer[next]
					if !exists {
						break
					}

					delete(buffer, next)

					if buffered.err != nil {
						var zero U
						if !yield(zero, buffered.err) {
							cancel()
							return
						}
					} else if !yield(buffered.value, nil) {
						cancel()
						return
					}

					next++
				}
			}
		}
	}
}
