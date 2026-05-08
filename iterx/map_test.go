package iterx_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex"
	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestMapSeq(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []string
	}{
		{
			name:     "transforms int to string",
			input:    []int{1, 2, 3},
			expected: []string{"item-1", "item-2", "item-3"},
		},
		{
			name:     "empty input",
			input:    []int{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []int{42},
			expected: []string{"item-42"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for v := range iterx.MapSeq(context.Background(), slices.Values(tt.input), func(n int) string {
				return fmt.Sprintf("item-%d", n)
			}) {
				got = append(got, v)
			}
			if got == nil {
				got = []string{}
			}
			if !slices.Equal(got, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestMapSeq_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var got []string
	for v := range iterx.MapSeq(ctx, slices.Values([]int{1, 2, 3}), func(n int) string {
		return fmt.Sprintf("%d", n)
	}) {
		got = append(got, v)
	}
	if len(got) != 0 {
		t.Fatalf("expected no output with cancelled context, got %v", got)
	}
}

func TestMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []string
	}{
		{
			name:     "transforms int to string",
			input:    []int{1, 2, 3},
			expected: []string{"x1", "x2", "x3"},
		},
		{
			name:     "empty input",
			input:    []int{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []int{7},
			expected: []string{"x7"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			err := iterx.Drain(context.Background(),
				iterx.Map(context.Background(), seqToSeq2(slices.Values(tt.input)), func(n int) string {
					return fmt.Sprintf("x%d", n)
				}),
				func(v string) error {
					got = append(got, v)
					return nil
				})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				got = []string{}
			}
			if !slices.Equal(got, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestMap_PropagatesError(t *testing.T) {
	sentinelErr := errors.New("map stream error")
	seq := func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}
		if !yield(2, nil) {
			return
		}
		if !yield(0, sentinelErr) {
			return
		}
		yield(3, nil)
	}

	var got []string
	err := iterx.Drain(context.Background(),
		iterx.Map(context.Background(), seq, func(n int) string { return fmt.Sprintf("%d", n) }),
		func(v string) error {
			got = append(got, v)
			return nil
		})

	if !errors.Is(err, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if !slices.Equal(got, []string{"1", "2"}) {
		t.Fatalf("expected [1 2] before error, got %v", got)
	}
}

func TestMap_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := iterx.Drain(ctx,
		iterx.Map(ctx, seqToSeq2(slices.Values([]int{1, 2, 3})), func(n int) string {
			return fmt.Sprintf("%d", n)
		}),
		func(v string) error { return nil })

	if !errors.Is(err, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", err)
	}
}

func TestMap_ContinuesAfterError(t *testing.T) {
	sentinelErr := errors.New("map stream error")
	seq := func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}
		if !yield(0, sentinelErr) {
			return
		}
		if !yield(3, nil) {
			return
		}
	}
	var got []string
	var errCount int
	for v, err := range iterx.Map(context.Background(), seq, func(n int) string { return fmt.Sprintf("x%d", n) }) {
		if err != nil {
			errCount++
			continue // consumer continues — exercises the continue in Map
		}
		got = append(got, v)
	}
	if !slices.Equal(got, []string{"x1", "x3"}) {
		t.Fatalf("expected [x1 x3], got %v", got)
	}
	if errCount != 1 {
		t.Fatalf("expected 1 error, got %d", errCount)
	}
}

func TestMap_StopsEarly(t *testing.T) {
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
	for _, err := range iterx.Map(context.Background(), seq, func(n int) string { return fmt.Sprintf("%d", n) }) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		count++
		if count == 2 {
			break
		}
	}

	if count != 2 {
		t.Fatalf("expected 2 items, got %d", count)
	}
	if callCount > 3 {
		t.Fatalf("expected upstream to stop early, got %d calls", callCount)
	}
}
