package iterx

import (
	"context"
	"iter"

	"github.com/MostafaMagdSalama/vortex"
)

// FlattenSeq converts a sequence of slices into a flat sequence of elements.
func FlattenSeq[T any](ctx context.Context, seq iter.Seq[[]T]) iter.Seq[T] {
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

// Flatten converts a sequence of slices into a flat sequence of elements.
// Errors from the underlying sequence are passed through untouched.
func Flatten[T any](ctx context.Context, seq iter.Seq2[[]T, error]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for slice, err := range seq {
			if ctx.Err() != nil {
				var zero T
				yield(zero, vortex.Wrap("iterx.Flatten", ctx.Err()))
				return
			}
			if err != nil {
				var zero T
				if !yield(zero, vortex.Wrap("iterx.Flatten", err)) {
					return
				}
				continue
			}
			for _, v := range slice {
				if ctx.Err() != nil {
					var zero T
					yield(zero, vortex.Wrap("iterx.Flatten", ctx.Err()))
					return
				}
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}
