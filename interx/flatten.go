package interx

import (
	"context"
	"iter"
)

// Flatten converts a sequence of slices into a flat sequence of elements.
func Flatten[T any](ctx context.Context, seq iter.Seq[[]T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for slice := range seq {
			if ctx.Err() != nil {
				return
			}
			for _, v := range slice {
				if ctx.Err() != nil {
					return
				}
				if !yield(v) {
					return
				}
			}
		}
	}
}
