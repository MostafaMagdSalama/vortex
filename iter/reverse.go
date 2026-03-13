package iter

import "iter"

// Reverse collects the sequence into memory and yields it in reverse order.
// Note: requires loading the full sequence — not suitable for infinite sequences.
//
// example:
//
//	numbers := slices.Values([]int{1, 2, 3, 4, 5})
//	for v := range iter.Reverse(numbers) {
//	    fmt.Println(v) // 5, 4, 3, 2, 1
//	}
func Reverse[T any](seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		// collect into slice first — needed to reverse
		var all []T
		for v := range seq {
			all = append(all, v)
		}

		// yield in reverse
		for i := len(all) - 1; i >= 0; i-- {
			if !yield(all[i]) {
				return
			}
		}
	}
}