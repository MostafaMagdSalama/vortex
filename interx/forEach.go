package interx

import (
	"context"
	"iter"
)

// ForEach calls fn for every element in the sequence.
func ForEach[T any](ctx context.Context, seq iter.Seq[T], fn func(T)) {
	for v := range seq {
		if ctx.Err() != nil {
			return
		}
		fn(v)
	}
}
