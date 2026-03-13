package iter

import "iter"

// ForEach calls fn for every element in the sequence.
// Use for side effects — logging, sending notifications, updating cache.
//
// example:
//
//	iter.ForEach(users, func(u User) {
//	    fmt.Println(u.Name)
//	})
func ForEach[T any](seq iter.Seq[T], fn func(T)) {
	for v := range seq {
		fn(v)
	}
}
