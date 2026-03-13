package iter

import (
	"context"
	"iter"
)

type ValidationError[T any] struct {
	Item   T
	Reason string
}

// ValidateConfig controls what happens with invalid items.
type ValidateConfig[T any] struct {
	OnError func(ValidationError[T])
}

// Filter returns a new sequence containing only elements where fn returns true.
func Filter[T any](ctx context.Context, seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if ctx.Err() != nil {
				return
			}
			if fn(v) && !yield(v) {
				return
			}
		}
	}
}

// Map transforms each element using fn.
func Map[T, U any](ctx context.Context, seq iter.Seq[T], fn func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if ctx.Err() != nil {
				return
			}
			if !yield(fn(v)) {
				return
			}
		}
	}
}

// Take returns the first n elements.
func Take[T any](ctx context.Context, seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		i := 0
		for v := range seq {
			if ctx.Err() != nil {
				return
			}
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

// FlatMap transforms each element into a sequence, then flattens all sequences.
func FlatMap[T, U any](ctx context.Context, seq iter.Seq[T], fn func(T) iter.Seq[U]) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if ctx.Err() != nil {
				return
			}
			for inner := range fn(v) {
				if ctx.Err() != nil {
					return
				}
				if !yield(inner) {
					return
				}
			}
		}
	}
}

func TakeWhile[T any](ctx context.Context, seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if ctx.Err() != nil {
				return
			}
			if !fn(v) {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

func Zip[A, B any](ctx context.Context, a iter.Seq[A], b iter.Seq[B]) iter.Seq[[2]any] {
	return func(yield func([2]any) bool) {
		nextA, stopA := iter.Pull(a)
		defer stopA()

		nextB, stopB := iter.Pull(b)
		defer stopB()

		for {
			if ctx.Err() != nil {
				return
			}

			va, okA := nextA()
			if !okA {
				return
			}

			vb, okB := nextB()
			if !okB {
				return
			}

			if !yield([2]any{va, vb}) {
				return
			}
		}
	}
}

func Validate[T any](ctx context.Context, seq iter.Seq[T], fn func(T) (bool, string), onError func(ValidationError[T])) iter.Seq[T] {
	return func(yield func(T) bool) {
		for item := range seq {
			if ctx.Err() != nil {
				return
			}
			ok, reason := fn(item)
			if !ok {
				if onError != nil {
					onError(ValidationError[T]{Item: item, Reason: reason})
				}
				continue
			}
			if !yield(item) {
				return
			}
		}
	}
}
