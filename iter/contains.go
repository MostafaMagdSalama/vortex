package iter

import "iter"

// Contains returns true if the sequence contains the target value.
// Stops as soon as the value is found — does not consume the rest.
//
// example:
//
//	numbers := slices.Values([]int{1, 2, 3, 4, 5})
//	fmt.Println(iter.Contains(numbers, 3)) // true
//	fmt.Println(iter.Contains(numbers, 9)) // false
func Contains[T comparable](seq iter.Seq[T], target T) bool {
	for v := range seq {
		if v == target {
			return true
		}
	}
	return false
}
