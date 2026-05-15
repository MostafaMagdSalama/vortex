package iterx_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex"
	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestZipSeq(t *testing.T) {
	tests := []struct {
		name     string
		aInput   []string
		bInput   []int
		expected []iterx.Pair[string, int]
	}{
		{
			name:   "equal length sequences",
			aInput: []string{"a", "b", "c"},
			bInput: []int{1, 2, 3},
			expected: []iterx.Pair[string, int]{
				{First: "a", Second: 1},
				{First: "b", Second: 2},
				{First: "c", Second: 3},
			},
		},
		{
			name:   "first sequence shorter",
			aInput: []string{"a", "b"},
			bInput: []int{1, 2, 3},
			expected: []iterx.Pair[string, int]{
				{First: "a", Second: 1},
				{First: "b", Second: 2},
			},
		},
		{
			name:   "second sequence shorter",
			aInput: []string{"a", "b", "c"},
			bInput: []int{1, 2},
			expected: []iterx.Pair[string, int]{
				{First: "a", Second: 1},
				{First: "b", Second: 2},
			},
		},
		{
			name:     "both empty",
			aInput:   []string{},
			bInput:   []int{},
			expected: []iterx.Pair[string, int]{},
		},
		{
			name:   "single element",
			aInput: []string{"x"},
			bInput: []int{99},
			expected: []iterx.Pair[string, int]{
				{First: "x", Second: 99},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []iterx.Pair[string, int]
			for pair := range iterx.ZipSeq(context.Background(), slices.Values(tt.aInput), slices.Values(tt.bInput)) {
				got = append(got, pair)
			}
			if got == nil {
				got = []iterx.Pair[string, int]{}
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d pairs, got %d: %v", len(tt.expected), len(got), got)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Fatalf("pair[%d]: expected %v, got %v", i, tt.expected[i], got[i])
				}
			}
		})
	}
}

func TestZipSeq_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var got []iterx.Pair[string, int]
	for pair := range iterx.ZipSeq(ctx, slices.Values([]string{"a", "b", "c"}), slices.Values([]int{1, 2, 3})) {
		got = append(got, pair)
	}
	if len(got) != 0 {
		t.Fatalf("expected no output with cancelled context, got %v", got)
	}
}

func TestZip(t *testing.T) {
	tests := []struct {
		name     string
		aInput   []string
		bInput   []int
		expected []iterx.Pair[string, int]
	}{
		{
			name:   "equal length no errors",
			aInput: []string{"a", "b", "c"},
			bInput: []int{1, 2, 3},
			expected: []iterx.Pair[string, int]{
				{First: "a", Second: 1},
				{First: "b", Second: 2},
				{First: "c", Second: 3},
			},
		},
		{
			name:     "both empty",
			aInput:   []string{},
			bInput:   []int{},
			expected: []iterx.Pair[string, int]{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []iterx.Pair[string, int]
			err := iterx.Drain(context.Background(),
				iterx.Zip(context.Background(), seqToSeq2(slices.Values(tt.aInput)), seqToSeq2(slices.Values(tt.bInput))),
				func(p iterx.Pair[string, int]) error {
					got = append(got, p)
					return nil
				})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				got = []iterx.Pair[string, int]{}
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d pairs, got %d", len(tt.expected), len(got))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Fatalf("pair[%d]: expected %v, got %v", i, tt.expected[i], got[i])
				}
			}
		})
	}
}

func TestZip_ErrorOnFirstSequence(t *testing.T) {
	sentinelErr := errors.New("error in first sequence")
	a := func(yield func(string, error) bool) {
		if !yield("a", nil) {
			return
		}
		if !yield("", sentinelErr) {
			return
		}
		yield("c", nil)
	}
	b := seqToSeq2(slices.Values([]int{1, 2, 3}))

	err := iterx.Drain(context.Background(),
		iterx.Zip(context.Background(), a, b),
		func(p iterx.Pair[string, int]) error { return nil })

	if !errors.Is(err, sentinelErr) {
		t.Fatalf("expected sentinel error from first sequence, got %v", err)
	}
}

func TestZip_ErrorOnSecondSequence(t *testing.T) {
	sentinelErr := errors.New("error in second sequence")
	a := seqToSeq2(slices.Values([]string{"a", "b", "c"}))
	b := func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}
		if !yield(0, sentinelErr) {
			return
		}
		yield(3, nil)
	}

	err := iterx.Drain(context.Background(),
		iterx.Zip(context.Background(), a, b),
		func(p iterx.Pair[string, int]) error { return nil })

	if !errors.Is(err, sentinelErr) {
		t.Fatalf("expected sentinel error from second sequence, got %v", err)
	}
}

func TestZip_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := iterx.Drain(ctx,
		iterx.Zip(ctx, seqToSeq2(slices.Values([]string{"a", "b", "c"})), seqToSeq2(slices.Values([]int{1, 2, 3}))),
		func(p iterx.Pair[string, int]) error { return nil })

	if !errors.Is(err, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", err)
	}
}

func TestZip_CancelledInErrorLoopA(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errA := errors.New("error from a")
	a := func(yield func(string, error) bool) {
		if !yield("", errA) {
			return
		}
		yield("real", nil)
	}
	b := seqToSeq2(slices.Values([]int{1, 2}))
	var errCount int
	for _, err := range iterx.Zip(ctx, a, b) {
		if err != nil {
			errCount++
			cancel() // cancel and continue — next inner-loop iteration detects cancelled ctx
		}
	}
	if errCount < 1 {
		t.Fatalf("expected at least 1 error, got %d", errCount)
	}
}

func TestZip_CancelledInErrorLoopB(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errB := errors.New("error from b")
	a := seqToSeq2(slices.Values([]string{"a1", "a2"}))
	b := func(yield func(int, error) bool) {
		if !yield(0, errB) {
			return
		}
		yield(99, nil)
	}
	var errCount int
	for _, err := range iterx.Zip(ctx, a, b) {
		if err != nil {
			errCount++
			cancel() // cancel and continue — next inner-loop iteration for b detects it
		}
	}
	if errCount < 1 {
		t.Fatalf("expected at least 1 error, got %d", errCount)
	}
}

func TestZip_StopsEarly(t *testing.T) {
	aCallCount := 0
	a := func(yield func(string, error) bool) {
		for _, v := range []string{"a", "b", "c", "d", "e"} {
			aCallCount++
			if !yield(v, nil) {
				return
			}
		}
	}

	count := 0
	for _, err := range iterx.Zip(context.Background(), a, seqToSeq2(slices.Values([]int{1, 2, 3, 4, 5}))) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		count++
		if count == 2 {
			break
		}
	}

	if count != 2 {
		t.Fatalf("expected 2 pairs, got %d", count)
	}
	if aCallCount > 3 {
		t.Fatalf("expected upstream to stop early, got %d calls", aCallCount)
	}
}
