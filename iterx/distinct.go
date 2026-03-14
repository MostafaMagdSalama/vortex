package iterx

import (
	"context"
	"iter"
)

// Distinct filters out duplicate values keeping only the first occurrence.
func Distinct[T comparable](ctx context.Context, seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[T]bool)

		for v := range seq {
			if ctx.Err() != nil {
				return
			}
			if !seen[v] {
				seen[v] = true
				if !yield(v) {
					return
				}
			}
		}
	}
}
