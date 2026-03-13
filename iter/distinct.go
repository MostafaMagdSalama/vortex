package iter

import "iter"

// Distinct filters out duplicate values keeping only the first occurrence.
// Uses a map internally so T must be comparable.
//
// example:
//
//	numbers := slices.Values([]int{1, 2, 1, 3, 2, 4})
//	for v := range iter.Distinct(numbers) {
//	    fmt.Println(v) // 1, 2, 3, 4
//	}
func Distinct[T comparable](seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[T]bool)

		for v := range seq {
			if !seen[v] {
				seen[v] = true
				if !yield(v) {
					return
				}
			}
		}
	}
}
