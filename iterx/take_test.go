package iterx_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex"
	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestTake(t *testing.T) {
	tests := []struct {
		name     string
		input    []rune
		n        int
		expected []rune
	}{
		{
			name:     "normal case",
			input:    []rune{'a', 'b', 'c', 'd'},
			n:        3,
			expected: []rune{'a', 'b', 'c'},
		},
		{
			name:     "take zero",
			input:    []rune{'a', 'b', 'c', 'd'},
			n:        0,
			expected: []rune{},
		},
		{
			name:     "empty input",
			input:    []rune{},
			n:        3,
			expected: []rune{},
		},
		{
			name:     "negative n",
			input:    []rune{'a', 'b', 'c', 'd'},
			n:        -1,
			expected: []rune{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []rune

			inputIter := seqToSeq2(slices.Values(tt.input))

			takeIter := iterx.Take(context.Background(), inputIter, tt.n)

			iterx.Drain(context.Background(), takeIter, func(r rune) error {
				result = append(result, r)
				return nil
			})

			if !slices.Equal(result, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
func TestTakeSeq(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		n        int
		expected []int
	}{
		{"takes first n", []int{1, 2, 3, 4, 5}, 3, []int{1, 2, 3}},
		{"take zero", []int{1, 2, 3}, 0, []int{}},
		{"take more than length", []int{1, 2}, 10, []int{1, 2}},
		{"empty input", []int{}, 3, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []int
			for v := range iterx.TakeSeq(context.Background(), slices.Values(tt.input), tt.n) {
				got = append(got, v)
			}
			if got == nil {
				got = []int{}
			}
			if !slices.Equal(got, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestTakeSeq_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var got []int
	for v := range iterx.TakeSeq(ctx, slices.Values([]int{1, 2, 3}), 10) {
		got = append(got, v)
	}
	if len(got) != 0 {
		t.Fatalf("expected no output with cancelled context, got %v", got)
	}
}

func TestTakeSeq_StopsEarly(t *testing.T) {
	var got []int
	for v := range iterx.TakeSeq(context.Background(), slices.Values([]int{1, 2, 3, 4, 5}), 10) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("expected [1 2], got %v", got)
	}
}

func TestTake_PropagatesError(t *testing.T) {
	sentinelErr := errors.New("stream error")
	seq := func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}
		if !yield(0, sentinelErr) {
			return
		}
		yield(2, nil)
	}
	var got []int
	var gotErr error
	for v, err := range iterx.Take(context.Background(), seq, 10) {
		if err != nil {
			gotErr = err
			break
		}
		got = append(got, v)
	}
	if !errors.Is(gotErr, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", gotErr)
	}
	if !slices.Equal(got, []int{1}) {
		t.Fatalf("expected [1] before error, got %v", got)
	}
}

func TestTake_ErrorsDoNotCountTowardN(t *testing.T) {
	sentinelErr := errors.New("stream error")
	seq := func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}
		if !yield(0, sentinelErr) {
			return
		}
		if !yield(2, nil) {
			return
		}
		yield(3, nil)
	}
	var got []int
	for v, err := range iterx.Take(context.Background(), seq, 2) {
		if err != nil {
			continue // errors pass through but don't count
		}
		got = append(got, v)
	}
	if !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("expected [1 2] (error skipped, n=2 values taken), got %v", got)
	}
}

func TestTake_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var gotErr error
	for _, err := range iterx.Take(ctx, seqToSeq2(slices.Values([]int{1, 2, 3})), 10) {
		if err != nil {
			gotErr = err
			break
		}
	}
	if !errors.Is(gotErr, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", gotErr)
	}
}

func TestTake_StopsEarly(t *testing.T) {
	var got []int
	for v, err := range iterx.Take(context.Background(), seqToSeq2(slices.Values([]int{1, 2, 3, 4, 5})), 10) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("expected [1 2], got %v", got)
	}
}
