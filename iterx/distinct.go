package iterx

import (
	"context"
	"iter"

	"github.com/MostafaMagdSalama/vortex"
)

// DistinctSeq filters out duplicate values keeping only the first occurrence.
//
// Memory: O(unique items). The internal seen-set grows for the lifetime of the
// iteration and is not bounded. For high-cardinality streams, consider chunking
// with iterx.Chunk and deduplicating per chunk, or use a probabilistic structure
// outside the pipeline.
func DistinctSeq[T comparable](ctx context.Context, seq iter.Seq[T]) iter.Seq[T] {
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

// Distinct filters out duplicate values keeping only the first occurrence.
// Errors from the underlying sequence are passed through untouched.
//
// Memory: O(unique items). The internal seen-set grows for the lifetime of the
// iteration and is not bounded. For high-cardinality streams, consider chunking
// with iterx.Chunk and deduplicating per chunk, or use a probabilistic structure
// outside the pipeline.
func Distinct[T comparable](ctx context.Context, seq iter.Seq2[T, error]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		seen := make(map[T]bool)

		for v, err := range seq {
			if ctx.Err() != nil {
				var zero T
				yield(zero, vortex.WrapCancelled("iterx.Distinct"))
				return
			}
			if err != nil {
				var zero T
				if !yield(zero, vortex.Wrap("iterx.Distinct", err)) {
					return
				}
				continue
			}
			if !seen[v] {
				seen[v] = true
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}
