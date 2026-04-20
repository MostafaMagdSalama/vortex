package iterx_test

import (
	"context"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestTakeWile(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		expected  []int
		condition func(int) bool
	}{
		{
			name: "normal test case", input: []int{1, 2, 3, 4, 5, 6, 7, 8}, expected: []int{1, 2, 3, 4, 5},
			condition: func(number int) bool {
				return number <= 5
			},
		},
		{name: "condition always true", input: []int{1, 2, 3, 4, 5, 6, 7, 8}, expected: []int{1, 2, 3, 4, 5, 6, 7, 8},
			condition: func(number int) bool {
				return true
			},
		},
		{name: "condition always false", input: []int{1, 2, 3, 4, 5, 6, 7, 8}, expected: []int{},
			condition: func(number int) bool {
				return false
			},
		},
		{name: "stop condition at first element", input: []int{5, 1, 2, 3, 4}, expected: []int{5},
			condition: func(number int) bool {
				return number > 3
			},
		},
		{
			name:      "empty input",
			input:     []int{},
			expected:  []int{},
			condition: func(number int) bool { return number > 0 },
		},
		{
			name:      "single element condition true",
			input:     []int{5},
			expected:  []int{5},
			condition: func(number int) bool { return number > 0 },
		},
		{
			name:      "single element condition false",
			input:     []int{5},
			expected:  []int{},
			condition: func(number int) bool { return number > 10 },
		},
		{
			name:      "does not resume after condition fails",
			input:     []int{1, 3, 2, 4, 5},
			expected:  []int{1, 3},
			condition: func(number int) bool { return number%2 != 0 },
		},
		{
			name:      "all pass except last",
			input:     []int{1, 2, 3, 4, 10},
			expected:  []int{1, 2, 3, 4},
			condition: func(number int) bool { return number < 10 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []int
			takeWileIter := iterx.TakeWhile(context.Background(), seqToSeq2(slices.Values(tt.input)), tt.condition)
			iterx.Drain(context.Background(), takeWileIter, func(number int) error {
				result = append(result, number)
				return nil
			})
			if !slices.Equal(result, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})

	}

}
