package iterx

import (
	"context"
	"iter"

	"github.com/MostafaMagdSalama/vortex"
)

// ContainsSeq returns true if the sequence contains the target value.
func ContainsSeq[T comparable](ctx context.Context, seq iter.Seq[T], target T) bool {
	for v := range seq {
		if ctx.Err() != nil {
			return false
		}
		if v == target {
			return true
		}
	}
	return false
}

// Contains returns true if the sequence contains the target value.
// It returns an error if the context is canceled or the underlying sequence yields an error.
func Contains[T comparable](ctx context.Context, seq iter.Seq2[T, error], target T) (bool, error) {
	for v, err := range seq {
		if ctx.Err() != nil {
			return false, vortex.WrapCancelled("iterx.Contains")
		}
		if err != nil {
			return false, vortex.Wrap("iterx.Contains", err)
		}
		if v == target {
			return true, nil
		}
	}
	return false, nil
}
