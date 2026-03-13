package iter

import "iter"

// Flatten converts a sequence of slices into a flat sequence of elements.
//
// example:
//
//	groups := slices.Values([][]int{{1, 2}, {3, 4}, {5}})
//	for v := range iter.Flatten(groups) {
//	    fmt.Println(v) // 1, 2, 3, 4, 5
//	}
func Flatten[T any](seq iter.Seq[[]T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for slice := range seq {
			for _, v := range slice {
				if !yield(v) {
					return
				}
			}
		}
	}
}
