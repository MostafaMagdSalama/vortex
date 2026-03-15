package iterx

import (
	"context"
	"iter"

	"github.com/MostafaMagdSalama/vortex"
)

// ForEachSeq calls fn for every element in the sequence.
func ForEachSeq[T any](ctx context.Context, seq iter.Seq[T], fn func(T)) {
	for v := range seq {
		if ctx.Err() != nil {
			return
		}
		fn(v)
	}
}

// ForEach calls fn for every element in the sequence.
// Stops immediately and returns the error if the context is cancelled or the underlying sequence yields an error.
func ForEach[T any](ctx context.Context, seq iter.Seq2[T, error], fn func(T)) error {
	for v, err := range seq {
		if ctx.Err() != nil {
			return vortex.Wrap("iterx.ForEach", ctx.Err())
		}
		if err != nil {
			return vortex.Wrap("iterx.ForEach", err)
		}
		fn(v)
	}
	return nil
}
