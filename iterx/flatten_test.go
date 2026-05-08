package iterx_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex"
	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestFlattenSeq(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]int
		expected []int
	}{
		{
			name:     "multiple batches",
			input:    [][]int{{1, 2}, {3}, {4, 5, 6}},
			expected: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "empty outer sequence",
			input:    [][]int{},
			expected: []int{},
		},
		{
			name:     "some empty inner slices",
			input:    [][]int{{1}, {}, {2, 3}, {}},
			expected: []int{1, 2, 3},
		},
		{
			name:     "single inner slice",
			input:    [][]int{{10, 20, 30}},
			expected: []int{10, 20, 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []int
			for v := range iterx.FlattenSeq(context.Background(), slices.Values(tt.input)) {
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

func TestFlattenSeq_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var got []int
	for v := range iterx.FlattenSeq(ctx, slices.Values([][]int{{1, 2}, {3, 4}})) {
		got = append(got, v)
	}
	if len(got) != 0 {
		t.Fatalf("expected no output with cancelled context, got %v", got)
	}
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]int
		expected []int
	}{
		{
			name:     "multiple batches",
			input:    [][]int{{1, 2}, {3}, {4, 5}},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "empty outer sequence",
			input:    [][]int{},
			expected: []int{},
		},
		{
			name:     "single batch",
			input:    [][]int{{7, 8, 9}},
			expected: []int{7, 8, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []int
			err := iterx.Drain(context.Background(),
				iterx.Flatten(context.Background(), seqToSeq2(slices.Values(tt.input))),
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

func TestFlatten_PropagatesError(t *testing.T) {
	sentinelErr := errors.New("flatten stream error")
	seq := func(yield func([]int, error) bool) {
		if !yield([]int{1, 2}, nil) {
			return
		}
		if !yield(nil, sentinelErr) {
			return
		}
		yield([]int{3, 4}, nil)
	}

	var got []int
	err := iterx.Drain(context.Background(),
		iterx.Flatten(context.Background(), seq),
		func(v int) error {
			got = append(got, v)
			return nil
		})

	if !errors.Is(err, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("expected [1 2] before error, got %v", got)
	}
}

func TestFlatten_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := iterx.Drain(ctx,
		iterx.Flatten(ctx, seqToSeq2(slices.Values([][]int{{1, 2}, {3, 4}}))),
		func(v int) error { return nil })

	if !errors.Is(err, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", err)
	}
}

func TestFlatten_StopsEarly(t *testing.T) {
	callCount := 0
	seq := func(yield func([]int, error) bool) {
		for _, v := range [][]int{{1, 2}, {3, 4}, {5, 6}} {
			callCount++
			if !yield(v, nil) {
				return
			}
		}
	}

	count := 0
	for _, err := range iterx.Flatten(context.Background(), seq) {
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
	if callCount > 2 {
		t.Fatalf("expected upstream to stop after first batch, got %d calls", callCount)
	}
}
