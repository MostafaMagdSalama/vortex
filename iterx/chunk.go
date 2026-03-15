package iterx

import (
	"context"
	"iter"

	"github.com/MostafaMagdSalama/vortex"
)

// ChunkSeq splits a sequence into slices of size n.
func ChunkSeq[T any](ctx context.Context, seq iter.Seq[T], n int) iter.Seq[[]T] {
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

// Chunk splits a sequence into slices of size n.
// Errors from the underlying sequence are passed through untouched, and they yield an empty batch with the error.
func Chunk[T any](ctx context.Context, seq iter.Seq2[T, error], n int) iter.Seq2[[]T, error] {
	return func(yield func([]T, error) bool) {
		batch := make([]T, 0, n)

		for v, err := range seq {
			if ctx.Err() != nil {
				yield(nil, vortex.Wrap("iterx.Chunk", ctx.Err()))
				return
			}
			if err != nil {
				if !yield(nil, vortex.Wrap("iterx.Chunk", err)) {
					return
				}
				continue
			}
			batch = append(batch, v)
			if len(batch) == n {
				if !yield(batch, nil) {
					return
				}
				batch = make([]T, 0, n)
			}
		}

		if len(batch) > 0 {
			if ctx.Err() != nil {
				yield(nil, vortex.Wrap("iterx.Chunk", ctx.Err()))
				return
			}
			yield(batch, nil)
		}
	}
}
