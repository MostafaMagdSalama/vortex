package iterx_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestForEach(t *testing.T) {
	tests := []struct {
		name     string
		input    []rune
		expected []rune
	}{
		{
			name:     "normal case",
			input:    []rune{'a', 'b', 'c', 'd'},
			expected: []rune{'a', 'b', 'c', 'd'},
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

			err := iterx.ForEach(context.Background(), inputIter, func(r rune) {
				result = append(result, r)
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !slices.Equal(result, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestForEach_PropagatesError(t *testing.T) {
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

	var result []rune

	err := iterx.ForEach(context.Background(), inputIter, func(r rune) {
		result = append(result, r)
	})

	if !errors.Is(err, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	if !slices.Equal(result, []rune{'a', 'b'}) {
		t.Fatalf("expected ['a', 'b'] before error, got %v", result)
	}
}

func TestForEach_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	inputIter := seqToSeq2(slices.Values([]rune{'a', 'b', 'c'}))

	var result []rune

	err := iterx.ForEach(ctx, inputIter, func(r rune) {
		result = append(result, r)
	})

	if err == nil {
		t.Fatalf("expected context error, got nil")
	}

	if len(result) > 0 {
		t.Fatalf("expected no results with cancelled context, got %v", result)
	}
}

func TestForEach_FnCalledInOrder(t *testing.T) {
	input := []rune{'a', 'b', 'c', 'd', 'e'}
	inputIter := seqToSeq2(slices.Values(input))

	var result []rune

	err := iterx.ForEach(context.Background(), inputIter, func(r rune) {
		result = append(result, r)
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !slices.Equal(result, input) {
		t.Fatalf("expected %v, got %v", input, result)
	}
}

func TestForEach_FnCalledForEachElement(t *testing.T) {
	input := []rune{'a', 'b', 'c', 'd'}
	inputIter := seqToSeq2(slices.Values(input))

	callCount := 0

	err := iterx.ForEach(context.Background(), inputIter, func(r rune) {
		callCount++
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != len(input) {
		t.Fatalf("expected fn to be called %d times, got %d", len(input), callCount)
	}
}

func TestForEach_FnNotCalledOnError(t *testing.T) {
	sentinelErr := errors.New("stream error")

	inputIter := func(yield func(rune, error) bool) {
	   if !yield(0, sentinelErr) { return }  
    if !yield('a', nil) { return }
	}

	callCount := 0

	err := iterx.ForEach(context.Background(), inputIter, func(r rune) {
		callCount++
	})

	if !errors.Is(err, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	if callCount != 0 {
		t.Fatalf("expected fn not to be called, got %d calls", callCount)
	}
}
