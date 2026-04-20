package iterx_test

import "iter"

func seqToSeq2[T any](seq iter.Seq[T]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for v := range seq {
			if !yield(v, nil) {
				return
			}
		}
	}
}
