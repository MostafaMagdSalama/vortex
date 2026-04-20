package iterx_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestDistinct(t *testing.T) {
	tests := []struct {
		name     string
		input    []rune
		expected []rune
	}{
		{
			name:     "no duplicates",
			input:    []rune{'a', 'b', 'c', 'd'},
			expected: []rune{'a', 'b', 'c', 'd'},
		},
		{
			name:     "all duplicates",
			input:    []rune{'a', 'a', 'a', 'a'},
			expected: []rune{'a'},
		},
		{
			name:     "duplicates at the beginning",
			input:    []rune{'a', 'a', 'b', 'c'},
			expected: []rune{'a', 'b', 'c'},
		},
		{
			name:     "duplicates at the end",
			input:    []rune{'a', 'b', 'c', 'c'},
			expected: []rune{'a', 'b', 'c'},
		},
		{
			name:     "duplicates in the middle",
			input:    []rune{'a', 'b', 'b', 'c'},
			expected: []rune{'a', 'b', 'c'},
		},
		{
			name:     "non-contiguous duplicates",
			input:    []rune{'a', 'b', 'a', 'c', 'b'},
			expected: []rune{'a', 'b', 'c'},
		},
		{
			name:     "single element",
			input:    []rune{'a'},
			expected: []rune{'a'},
		},
		{
			name:     "empty input",
			input:    []rune{},
			expected: []rune{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []rune

			inputIter := seqToSeq2(slices.Values(tt.input))

			distinctIter := iterx.Distinct(context.Background(), inputIter)

			iterx.Drain(context.Background(), distinctIter, func(r rune) error {
				result = append(result, r)
				return nil
			})

			if !slices.Equal(result, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDistinct_PropagatesError(t *testing.T) {
	sentinelErr := errors.New("stream error")

	inputIter := func(yield func(rune, error) bool) {
		if !yield('a', nil) {
			return
		}
		if !yield('b', nil) {
			return
		}
		if !yield(0, sentinelErr) {
			return
		}
		if !yield('c', nil) {
			return
		}
	}

	var gotErr error
	var result []rune

	distinctIter := iterx.Distinct(context.Background(), inputIter)

	distinctIter(func(r rune, err error) bool {
		if err != nil {
			gotErr = err
			return false
		}
		result = append(result, r)
		return true
	})

	if !errors.Is(gotErr, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", gotErr)
	}

	if !slices.Equal(result, []rune{'a', 'b'}) {
		t.Fatalf("expected ['a', 'b'] before error, got %v", result)
	}
}

func TestDistinct_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	inputIter := seqToSeq2(slices.Values([]rune{'a', 'b', 'c'}))

	distinctIter := iterx.Distinct(ctx, inputIter)

	var result []rune
	distinctIter(func(r rune, err error) bool {
		if err != nil {
			return false
		}
		result = append(result, r)
		return true
	})

	if len(result) > 0 {
		t.Fatalf("expected no results with cancelled context, got %v", result)
	}
}

func TestDistinct_StopsEarly(t *testing.T) {
	input := []rune{'a', 'a', 'b', 'b', 'c', 'c'}
	inputIter := seqToSeq2(slices.Values(input))

	distinctIter := iterx.Distinct(context.Background(), inputIter)

	var result []rune
	distinctIter(func(r rune, err error) bool {
		if err != nil {
			return false
		}
		result = append(result, r)
		return len(result) < 2
	})

	if !slices.Equal(result, []rune{'a', 'b'}) {
		t.Fatalf("expected ['a', 'b'] after early stop, got %v", result)
	}
}

func TestDistinct_PreservesOrder(t *testing.T) {

	input := []rune{'c', 'a', 'b', 'a', 'c', 'b'}
	inputIter := seqToSeq2(slices.Values(input))

	var result []rune

	distinctIter := iterx.Distinct(context.Background(), inputIter)

	iterx.Drain(context.Background(), distinctIter, func(r rune) error {
		result = append(result, r)
		return nil
	})

	expected := []rune{'c', 'a', 'b'}
	if !slices.Equal(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}
