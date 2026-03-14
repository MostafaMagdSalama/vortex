package iterx

import (
	"context"
	"iter"
)

// Reverse collects the sequence into memory and yields it in reverse order.
func Reverse[T any](ctx context.Context, seq iter.Seq[T]) iter.Seq[T] {
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
