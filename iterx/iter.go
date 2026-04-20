package iterx

import (
	"context"
	"fmt"
	"iter"

	"github.com/MostafaMagdSalama/vortex"
)

// ValidationError represents an item that failed validation.
type ValidationError[T any] struct {
	Item   T
	Reason string
}

// Error implements the error interface.
func (e ValidationError[T]) Error() string {
	return fmt.Sprintf("vortex: validation failed for item %v: %s", e.Item, e.Reason)
}

// FilterSeq returns a new sequence containing only elements where fn returns true.
func FilterSeq[T any](ctx context.Context, seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
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

// Filter returns a new sequence containing only elements where fn returns true.
// Errors from the underlying sequence are passed through untouched.
func Filter[T any](ctx context.Context, seq iter.Seq2[T, error], fn func(T) bool) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for v, err := range seq {
			if ctx.Err() != nil {
				var zero T
				yield(zero, vortex.Wrap("iterx.Filter", ctx.Err()))
				return
			}
			if err != nil {
				var zero T
				if !yield(zero, vortex.Wrap("iterx.Filter", err)) {
					return
				}
				continue
			}
			if fn(v) && !yield(v, nil) {
				return
			}
		}
	}
}

// MapSeq transforms each element using fn.
func MapSeq[T, U any](ctx context.Context, seq iter.Seq[T], fn func(T) U) iter.Seq[U] {
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

// Map transforms each element using fn.
// Errors from the underlying sequence are passed through untouched.
func Map[T, U any](ctx context.Context, seq iter.Seq2[T, error], fn func(T) U) iter.Seq2[U, error] {
	return func(yield func(U, error) bool) {
		for v, err := range seq {
			if ctx.Err() != nil {
				var zero U
				yield(zero, vortex.Wrap("iterx.Map", ctx.Err()))
				return
			}
			if err != nil {
				var zero U
				if !yield(zero, vortex.Wrap("iterx.Map", err)) {
					return
				}
				continue
			}
			if !yield(fn(v), nil) {
				return
			}
		}
	}
}

// TakeSeq returns the first n elements.
func TakeSeq[T any](ctx context.Context, seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		i := 0
		for v := range seq {
			if ctx.Err() != nil {
				return
			}

			if !yield(v) {
				return
			}
			i++
			if i >= n {
				return
			}

		}
	}
}

// Take returns the first n elements.
// Errors from the underlying sequence are passed through untouched, and do not count towards n.
func Take[T any](ctx context.Context, seq iter.Seq2[T, error], n int) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		i := 0
		for v, err := range seq {
			if i >= n {
				return
			}
			if ctx.Err() != nil {
				var zero T
				yield(zero, vortex.Wrap("iterx.Take", ctx.Err()))
				return
			}
			if err != nil {
				var zero T
				if !yield(zero, vortex.Wrap("iterx.Take", err)) {
					return
				}
				continue
			}

			if !yield(v, nil) {
				return
			}

			i++

		}
	}
}

// FlatMapSeq transforms each element into a sequence, then flattens all sequences.
func FlatMapSeq[T, U any](ctx context.Context, seq iter.Seq[T], fn func(T) iter.Seq[U]) iter.Seq[U] {
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

// FlatMap transforms each element into a sequence, then flattens all sequences.
// Errors from the underlying sequence or inner sequences are passed through.
func FlatMap[T, U any](ctx context.Context, seq iter.Seq2[T, error], fn func(T) iter.Seq2[U, error]) iter.Seq2[U, error] {
	return func(yield func(U, error) bool) {
		for v, err := range seq {
			if ctx.Err() != nil {
				var zero U
				yield(zero, vortex.Wrap("iterx.FlatMap", ctx.Err()))
				return
			}
			if err != nil {
				var zero U
				if !yield(zero, vortex.Wrap("iterx.FlatMap", err)) {
					return
				}
				continue
			}
			for inner, errInner := range fn(v) {
				if ctx.Err() != nil {
					var zero U
					yield(zero, vortex.Wrap("iterx.FlatMap", ctx.Err()))
					return
				}
				if errInner != nil {
					var zero U
					if !yield(zero, vortex.Wrap("iterx.FlatMap", errInner)) {
						return
					}
					continue
				}
				if !yield(inner, nil) {
					return
				}
			}
		}
	}
}

// TakeWhileSeq yields values from the sequence as long as fn returns true.
// Iteration stops as soon as fn evaluates to false for the first time.
func TakeWhileSeq[T any](ctx context.Context, seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
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

// TakeWhile yields values from the sequence as long as fn returns true.
// Iteration stops as soon as fn evaluates to false for the first time.
// Errors from the underlying sequence are passed through untouched and do not cause stopping.
func TakeWhile[T any](ctx context.Context, seq iter.Seq2[T, error], fn func(T) bool) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for v, err := range seq {
			if ctx.Err() != nil {
				var zero T
				yield(zero, vortex.Wrap("iterx.TakeWhile", ctx.Err()))
				return
			}
			if err != nil {
				var zero T
				if !yield(zero, vortex.Wrap("iterx.TakeWhile", err)) {
					return
				}
				continue
			}
			if !fn(v) {
				return
			}
			if !yield(v, nil) {
				return
			}
		}
	}
}

// ZipSeq combines two sequences into pairs, yielding [2]any{a, b} for each
// corresponding element. It stops as soon as the shortest sequence runs out.
func ZipSeq[A, B any](ctx context.Context, a iter.Seq[A], b iter.Seq[B]) iter.Seq[[2]any] {
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

// Zip combines two sequences into pairs, yielding [2]any{a, b} for each
// corresponding valid element. It skips errors from both sequences independently,
// yielding the errors directly in the stream. It stops as soon as either sequence runs out of valid elements safely.
func Zip[A, B any](ctx context.Context, a iter.Seq2[A, error], b iter.Seq2[B, error]) iter.Seq2[[2]any, error] {
	return func(yield func([2]any, error) bool) {
		nextA, stopA := iter.Pull2(a)
		defer stopA()

		nextB, stopB := iter.Pull2(b)
		defer stopB()

		for {
			if ctx.Err() != nil {
				var zero [2]any
				yield(zero, vortex.Wrap("iterx.Zip", ctx.Err()))
				return
			}

			var va A
			var errA error
			var okA bool

			// Get next valid item from a, yielding any errors encountered along the way
			for {
				if ctx.Err() != nil {
					var zero [2]any
					yield(zero, vortex.Wrap("iterx.Zip", ctx.Err()))
					return
				}
				va, errA, okA = nextA()
				if !okA {
					break // 'a' ran out
				}
				if errA != nil {
					var zero [2]any
					if !yield(zero, vortex.Wrap("iterx.Zip", errA)) {
						return
					}
					continue
				}
				break
			}
			if !okA {
				return
			}

			var vb B
			var errB error
			var okB bool

			// Get next valid item from b, yielding any errors encountered along the way
			for {
				if ctx.Err() != nil {
					var zero [2]any
					yield(zero, vortex.Wrap("iterx.Zip", ctx.Err()))
					return
				}
				vb, errB, okB = nextB()
				if !okB {
					break // 'b' ran out
				}
				if errB != nil {
					var zero [2]any
					if !yield(zero, vortex.Wrap("iterx.Zip", errB)) {
						return
					}
					continue
				}
				break
			}
			if !okB {
				return
			}

			if !yield([2]any{va, vb}, nil) {
				return
			}
		}
	}
}

// ValidateSeq conditionally streams elements by evaluating fn(item). If fn yields
// {false, reason}, the onError callback is triggered with a ValidationError,
// and the element is discarded from the resulting sequence.
func ValidateSeq[T any](ctx context.Context, seq iter.Seq[T], fn func(T) (bool, string), onError func(ValidationError[T])) iter.Seq[T] {
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

// Validate conditionally streams elements by evaluating fn(item). If fn yields
// {false, reason}, the onError callback is triggered with a ValidationError,
// and the element is discarded from the resulting sequence.
func Validate[T any](ctx context.Context, seq iter.Seq2[T, error], fn func(T) (bool, string), onError func(ValidationError[T])) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for item, err := range seq {
			if ctx.Err() != nil {
				var zero T
				yield(zero, vortex.Wrap("iterx.Validate", ctx.Err()))
				return
			}
			if err != nil {
				var zero T
				if !yield(zero, vortex.Wrap("iterx.Validate", err)) {
					return
				}
				continue
			}
			ok, reason := fn(item)
			if !ok {
				if onError != nil {
					onError(ValidationError[T]{Item: item, Reason: reason})
				}
				continue
			}
			if !yield(item, nil) {
				return
			}
		}
	}
}
