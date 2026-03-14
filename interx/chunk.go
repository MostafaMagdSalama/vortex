package interx

import (
	"context"
	"iter"
)

// Chunk splits a sequence into slices of size n.
func Chunk[T any](ctx context.Context, seq iter.Seq[T], n int) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		batch := make([]T, 0, n)

		for v := range seq {
			if ctx.Err() != nil {
				return
			}
			batch = append(batch, v)
			if len(batch) == n {
				if !yield(batch) {
					return
				}
				batch = make([]T, 0, n)
			}
		}

		if len(batch) > 0 {
			if ctx.Err() != nil {
				return
			}
			yield(batch)
		}
	}
}
