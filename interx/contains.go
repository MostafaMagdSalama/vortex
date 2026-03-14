package interx

import (
	"context"
	"iter"
)

// Contains returns true if the sequence contains the target value.
func Contains[T comparable](ctx context.Context, seq iter.Seq[T], target T) bool {
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
