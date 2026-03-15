package iterx

import (
	"context"
	"iter"

	"github.com/MostafaMagdSalama/vortex"
)

// DrainSeq consumes a sequence and calls fn for each item.
// Stops immediately if ctx is cancelled or fn returns an error.
// Use DrainSeq when your terminal operation can fail — writing to CSV, DB, files.
// Use ForEachSeq when your terminal operation cannot fail — logging, printing.
func DrainSeq[T any](ctx context.Context, seq iter.Seq[T], fn func(T) error) error {
	for v := range seq {
		if ctx.Err() != nil {
			return vortex.Wrap("iterx.DrainSeq", ctx.Err())
		}
		if err := fn(v); err != nil {
			return vortex.Wrap("iterx.DrainSeq", err)
		}
	}
	return nil
}

// Drain consumes a sequence and calls fn for each item.
// Stops immediately if ctx is cancelled, fn returns an error, or the underlying sequence yields an error.
// Use Drain when your terminal operation can fail — writing to CSV, DB, files.
// Use ForEach when your terminal operation cannot fail — logging, printing.
func Drain[T any](ctx context.Context, seq iter.Seq2[T, error], fn func(T) error) error {
	for v, err := range seq {
		if ctx.Err() != nil {
			return vortex.Wrap("iterx.Drain", ctx.Err())
		}
		if err != nil {
			return vortex.Wrap("iterx.Drain", err)
		}
		if err := fn(v); err != nil {
			return vortex.Wrap("iterx.Drain", err)
		}
	}
	return nil
}
