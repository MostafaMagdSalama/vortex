package iter

import "iter"

// Chunk splits a sequence into slices of size n.
// The last slice may be smaller if the sequence length is not divisible by n.
//
// example:
//
//	for batch := range iter.Chunk(users, 3) {
//	    fmt.Println(batch) // [1 2 3], [4 5 6], [7]
//	}
func Chunk[T any](seq iter.Seq[T], n int) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		batch := make([]T, 0, n)

		for v := range seq {
			batch = append(batch, v)
			if len(batch) == n {
				if !yield(batch) {
					return
				}
				batch = make([]T, 0, n) // new slice for next batch
			}
		}

		// yield remaining items
		if len(batch) > 0 {
			yield(batch)
		}
	}
}
