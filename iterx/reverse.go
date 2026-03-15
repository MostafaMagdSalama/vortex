package iterx

import (
	"context"
	"iter"

	"github.com/MostafaMagdSalama/vortex"
)

// ReverseSeq collects the sequence into memory and yields it in reverse order.
func ReverseSeq[T any](ctx context.Context, seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		var all []T
		for v := range seq {
			if ctx.Err() != nil {
				return
			}
			all = append(all, v)
		}

		for i := len(all) - 1; i >= 0; i-- {
			if ctx.Err() != nil {
				return
			}
			if !yield(all[i]) {
				return
			}
		}
	}
}

// Reverse collects the sequence into memory and yields it in reverse order.
// Any errors encountered during collection or yielding are passed through inline.
func Reverse[T any](ctx context.Context, seq iter.Seq2[T, error]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		type itemErr struct {
			v   T
			err error
		}
		var items []itemErr

		for v, err := range seq {
			if ctx.Err() != nil {
				var zero T
				yield(zero, vortex.Wrap("iterx.Reverse", ctx.Err()))
				return
			}
			items = append(items, itemErr{v, err})
		}

		for i := len(items) - 1; i >= 0; i-- {
			if ctx.Err() != nil {
				var zero T
				yield(zero, vortex.Wrap("iterx.Reverse", ctx.Err()))
				return
			}
			if items[i].err != nil {
				var zero T
				if !yield(zero, vortex.Wrap("iterx.Reverse", items[i].err)) {
					return
				}
				continue
			}
			if !yield(items[i].v, nil) {
				return
			}
		}
	}
}
