package iter

import "iter"

// Filter returns a new sequence containing only elements where fn returns true.
// Nothing runs until the caller does `for range`.
func Filter[T any](seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
    return func(yield func(T) bool) {
        for v := range seq {
            if fn(v) {
                if !yield(v) {
                    return // caller broke out of the loop — stop early
                }
            }
        }
    }
}

// Map transforms each element using fn.
func Map[T, U any](seq iter.Seq[T], fn func(T) U) iter.Seq[U] {
    return func(yield func(U) bool) {
        for v := range seq {
            if !yield(fn(v)) {
                return
            }
        }
    }
}

// Take returns the first n elements.
func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
    return func(yield func(T) bool) {
        i := 0
        for v := range seq {
            if i >= n {
                return
            }
            if !yield(v) {
                return
            }
            i++
        }
    }
}

// FlatMap transforms each element into a sequence, then flattens
// all sequences into one.
//
// example:
// input:  [1, 2, 3]
// fn:     n -> [n, n*10]
// output: [1, 10, 2, 20, 3, 30]
func FlatMap[T, U any](seq iter.Seq[T], fn func(T) iter.Seq[U]) iter.Seq[U] {
    return func(yield func(U) bool) {
        for v := range seq {
            // fn returns a sequence — range over it
            for inner := range fn(v) {
                if !yield(inner) {
                    return // caller stopped, exit everything
                }
            }
        }
    }
}

func TakeWhile[T any](seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
    return func(yield func(T) bool) {
        for v := range seq {
            if !fn(v) {
                return // condition failed — stop
            }
            if !yield(v) {
                return // caller stopped — stop
            }
        }
    }
}