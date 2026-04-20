package iterx_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		input    []rune
		target   rune
		expected bool
	}{
		{
			name:     "target exists in middle",
			input:    []rune{'a', 'b', 'c', 'd'},
			target:   'b',
			expected: true,
		},
		{
			name:     "target is first element",
			input:    []rune{'a', 'b', 'c', 'd'},
			target:   'a',
			expected: true,
		},
		{
			name:     "target is last element",
			input:    []rune{'a', 'b', 'c', 'd'},
			target:   'd',
			expected: true,
		},
		{
			name:     "target not in sequence",
			input:    []rune{'a', 'b', 'c', 'd'},
			target:   'z',
			expected: false,
		},
		{
			name:     "empty input",
			input:    []rune{},
			target:   'a',
			expected: false,
		},
		{
			name:     "single element match",
			input:    []rune{'x'},
			target:   'x',
			expected: true,
		},
		{
			name:     "single element no match",
			input:    []rune{'x'},
			target:   'y',
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputIter := seqToSeq2(slices.Values(tt.input))

			got, err := iterx.Contains(context.Background(), inputIter, tt.target)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestContains_PropagatesError(t *testing.T) {
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

	got, err := iterx.Contains(context.Background(), inputIter, 'c')

	if !errors.Is(err, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	if got {
		t.Fatalf("expected false when error occurs before finding target")
	}
}

func TestContains_StopsEarlyOnFound(t *testing.T) {
	callCount := 0

	inputIter := func(yield func(rune, error) bool) {
		for _, v := range []rune{'a', 'b', 'c', 'd', 'e'} {
			callCount++
			if !yield(v, nil) {
				return
			}
		}
	}

	got, err := iterx.Contains(context.Background(), inputIter, 'b')

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !got {
		t.Fatalf("expected true, got false")
	}

	if callCount != 2 {
		t.Fatalf("expected iterator to stop early, but got %d calls", callCount)
	}
}

func TestContains_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	inputIter := seqToSeq2(slices.Values([]rune{'a', 'b', 'c'}))

	got, err := iterx.Contains(ctx, inputIter, 'a')

	if err == nil {
		t.Fatalf("expected context error, got nil")
	}

	if got {
		t.Fatalf("expected false with cancelled context")
	}
}
