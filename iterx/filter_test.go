package iterx_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex"
	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestFilterSeq(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expected  []int
	}{
		{
			name:      "filters even numbers",
			input:     []int{1, 2, 3, 4, 5, 6},
			predicate: func(n int) bool { return n%2 == 0 },
			expected:  []int{2, 4, 6},
		},
		{
			name:      "empty input",
			input:     []int{},
			predicate: func(n int) bool { return true },
			expected:  []int{},
		},
		{
			name:      "predicate always false",
			input:     []int{1, 2, 3},
			predicate: func(n int) bool { return false },
			expected:  []int{},
		},
		{
			name:      "predicate always true",
			input:     []int{1, 2, 3},
			predicate: func(n int) bool { return true },
			expected:  []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []int
			for v := range iterx.FilterSeq(context.Background(), slices.Values(tt.input), tt.predicate) {
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

func TestFilterSeq_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var got []int
	for v := range iterx.FilterSeq(ctx, slices.Values([]int{1, 2, 3}), func(n int) bool { return true }) {
		got = append(got, v)
	}
	if len(got) != 0 {
		t.Fatalf("expected no output with cancelled context, got %v", got)
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expected  []int
	}{
		{
			name:      "filters even numbers",
			input:     []int{1, 2, 3, 4, 5, 6},
			predicate: func(n int) bool { return n%2 == 0 },
			expected:  []int{2, 4, 6},
		},
		{
			name:      "empty input",
			input:     []int{},
			predicate: func(n int) bool { return true },
			expected:  []int{},
		},
		{
			name:      "predicate always false",
			input:     []int{1, 2, 3},
			predicate: func(n int) bool { return false },
			expected:  []int{},
		},
		{
			name:      "predicate always true",
			input:     []int{1, 2, 3},
			predicate: func(n int) bool { return true },
			expected:  []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []int
			err := iterx.Drain(context.Background(),
				iterx.Filter(context.Background(), seqToSeq2(slices.Values(tt.input)), tt.predicate),
				func(v int) error {
					got = append(got, v)
					return nil
				})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
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

func TestFilter_PropagatesError(t *testing.T) {
	sentinelErr := errors.New("stream error")
	seq := func(yield func(int, error) bool) {
		if !yield(2, nil) {
			return
		}
		if !yield(4, nil) {
			return
		}
		if !yield(0, sentinelErr) {
			return
		}
		yield(6, nil)
	}

	var got []int
	err := iterx.Drain(context.Background(),
		iterx.Filter(context.Background(), seq, func(n int) bool { return true }),
		func(v int) error {
			got = append(got, v)
			return nil
		})

	if !errors.Is(err, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if !slices.Equal(got, []int{2, 4}) {
		t.Fatalf("expected [2 4] before error, got %v", got)
	}
}

func TestFilter_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := iterx.Drain(ctx,
		iterx.Filter(ctx, seqToSeq2(slices.Values([]int{1, 2, 3})), func(n int) bool { return true }),
		func(v int) error { return nil })

	if !errors.Is(err, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", err)
	}
}

func TestFilter_ContinuesAfterError(t *testing.T) {
	sentinelErr := errors.New("filter stream error")
	seq := func(yield func(int, error) bool) {
		if !yield(2, nil) {
			return
		}
		if !yield(0, sentinelErr) {
			return
		}
		if !yield(4, nil) {
			return
		}
	}
	var got []int
	var errCount int
	for v, err := range iterx.Filter(context.Background(), seq, func(n int) bool { return n%2 == 0 }) {
		if err != nil {
			errCount++
			continue // consumer continues — exercises the continue in Filter
		}
		got = append(got, v)
	}
	if !slices.Equal(got, []int{2, 4}) {
		t.Fatalf("expected [2 4], got %v", got)
	}
	if errCount != 1 {
		t.Fatalf("expected 1 error, got %d", errCount)
	}
}

func TestFilter_StopsEarly(t *testing.T) {
	callCount := 0
	seq := func(yield func(int, error) bool) {
		for _, v := range []int{1, 2, 3, 4, 5} {
			callCount++
			if !yield(v, nil) {
				return
			}
		}
	}

	count := 0
	for v, err := range iterx.Filter(context.Background(), seq, func(n int) bool { return true }) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		count++
		if v == 2 {
			break
		}
	}

	if count != 2 {
		t.Fatalf("expected 2 items consumed, got %d", count)
	}
	if callCount > 3 {
		t.Fatalf("expected upstream to stop early, got %d calls", callCount)
	}
}
