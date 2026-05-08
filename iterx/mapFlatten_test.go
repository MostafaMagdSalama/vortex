package iterx_test

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex"
	iterx "github.com/MostafaMagdSalama/vortex/iterx"
)

func TestFlatMap(t *testing.T) {
	type Order struct{ name string }
	type User struct{ orders []Order }

	tests := []struct {
		name    string
		users   []User
		usersIter iter.Seq2[User , error]
		buildFn func(User) iter.Seq2[Order, error]
		want    []string
		wantErr bool
	}{
		{
			name:  "empty input",
			users: []User{},
			buildFn: func(u User) iter.Seq2[Order, error] {
				return seqToSeq2(slices.Values(u.orders))
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "normal case",
			users: []User{
				{orders: []Order{{"iphone"}, {"macbook"}}},
			},
			buildFn: func(u User) iter.Seq2[Order, error] {
				return seqToSeq2(slices.Values(u.orders))
			},
			want:    []string{"iphone", "macbook"},
			wantErr: false,
		},
		{
			name: "inner error",
			users: []User{
				{orders: []Order{{"iphone"}}},
			},
			buildFn: func(u User) iter.Seq2[Order, error] {
				return func(yield func(Order, error) bool) {
					yield(Order{}, fmt.Errorf("inner error"))
				}
			},
			wantErr: true,
		},
		{
			name: "outer error",
			usersIter:func(yield func(User, error)bool)  {
				 yield(User{}, fmt.Errorf("outer Error"))
			},
			buildFn: func(u User) iter.Seq2[Order, error] {
				return func(yield func(Order, error) bool) {
					yield(Order{}, fmt.Errorf("inner error"))
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []string

			usersIter := tt.usersIter
			if usersIter == nil {
				usersIter = seqToSeq2(slices.Values(tt.users))
			}

			orderIter := iterx.FlatMap(context.Background(), usersIter, tt.buildFn)

			err := iterx.Drain(context.Background(), orderIter, func(o Order) error {
				result = append(result, o.name)
				return nil
			})

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !slices.Equal(result, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, result)
			}
		})
	}
}

func TestFlatMapSeq(t *testing.T) {
	seq := slices.Values([]int{1, 2, 3})
	fn := func(n int) iter.Seq[int] {
		return slices.Values([]int{n * 10, n * 10 + 1})
	}
	var got []int
	for v := range iterx.FlatMapSeq(context.Background(), seq, fn) {
		got = append(got, v)
	}
	if !slices.Equal(got, []int{10, 11, 20, 21, 30, 31}) {
		t.Fatalf("expected [10 11 20 21 30 31], got %v", got)
	}
}

func TestFlatMapSeq_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var got []int
	for v := range iterx.FlatMapSeq(ctx, slices.Values([]int{1, 2}), func(n int) iter.Seq[int] {
		return slices.Values([]int{n, n + 1})
	}) {
		got = append(got, v)
	}
	if len(got) != 0 {
		t.Fatalf("expected no output with cancelled context, got %v", got)
	}
}

func TestFlatMapSeq_CancelledDuringInner(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	outer := slices.Values([]int{1})
	fn := func(n int) iter.Seq[int] {
		return func(yield func(int) bool) {
			cancel()       // cancel before first inner yield
			yield(n * 10) // inner loop body checks ctx and returns before yielding to consumer
		}
	}
	var got []int
	for v := range iterx.FlatMapSeq(ctx, outer, fn) {
		got = append(got, v)
	}
	if len(got) != 0 {
		t.Fatalf("expected no output after inner cancel, got %v", got)
	}
}

func TestFlatMapSeq_StopsEarly(t *testing.T) {
	outerCalls := 0
	outer := func(yield func(int) bool) {
		for _, v := range []int{1, 2, 3} {
			outerCalls++
			if !yield(v) {
				return
			}
		}
	}
	var got []int
	for v := range iterx.FlatMapSeq(context.Background(), outer, func(n int) iter.Seq[int] {
		return slices.Values([]int{n * 10, n*10 + 1})
	}) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{10, 11}) {
		t.Fatalf("expected [10 11], got %v", got)
	}
	if outerCalls > 2 {
		t.Fatalf("expected upstream to stop, got %d outer calls", outerCalls)
	}
}

func TestFlatMap_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	outer := seqToSeq2(slices.Values([]int{1, 2}))
	var gotErr error
	for _, err := range iterx.FlatMap(ctx, outer, func(n int) iter.Seq2[int, error] {
		return seqToSeq2(slices.Values([]int{n}))
	}) {
		if err != nil {
			gotErr = err
			break
		}
	}
	if !errors.Is(gotErr, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", gotErr)
	}
}

func TestFlatMap_CancelledDuringInner(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	innerCallCount := 0
	outer := seqToSeq2(slices.Values([][]int{{1, 2}, {3, 4}}))
	fn := func(nums []int) iter.Seq2[int, error] {
		return func(yield func(int, error) bool) {
			for _, n := range nums {
				innerCallCount++
				if innerCallCount == 2 {
					cancel() // cancel before yielding second inner item
				}
				if !yield(n, nil) {
					return
				}
			}
		}
	}
	var got []int
	var gotErr error
	for v, err := range iterx.FlatMap(ctx, outer, fn) {
		if err != nil {
			gotErr = err
			break
		}
		got = append(got, v)
	}
	if !errors.Is(gotErr, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled during inner, got %v", gotErr)
	}
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("expected [1] before cancellation, got %v", got)
	}
}

func TestFlatMap_StopsEarlyOnValue(t *testing.T) {
	outer := seqToSeq2(slices.Values([]int{1, 2, 3}))
	var got []int
	for v, err := range iterx.FlatMap(context.Background(), outer, func(n int) iter.Seq2[int, error] {
		return seqToSeq2(slices.Values([]int{n * 10, n*10 + 1}))
	}) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{10, 11}) {
		t.Fatalf("expected [10 11], got %v", got)
	}
}

func TestFlatMap_StopsEarlyOnOuterError(t *testing.T) {
	sentinelErr := errors.New("outer error")
	outer := func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}
		if !yield(0, sentinelErr) {
			return
		}
		yield(2, nil)
	}
	var gotErr error
	var got []int
	for v, err := range iterx.FlatMap(context.Background(), outer, func(n int) iter.Seq2[int, error] {
		return seqToSeq2(slices.Values([]int{n * 10}))
	}) {
		if err != nil {
			gotErr = err
			break // consumer stops on error
		}
		got = append(got, v)
	}
	if !errors.Is(gotErr, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", gotErr)
	}
	if !slices.Equal(got, []int{10}) {
		t.Fatalf("expected [10] before error, got %v", got)
	}
}
