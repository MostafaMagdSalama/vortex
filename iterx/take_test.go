package iterx_test

import (
	"context"
	"slices"
	"testing"

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