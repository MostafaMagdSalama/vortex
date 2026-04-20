package iterx_test

import (
	"context"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestReverse(t *testing.T) {
    tests := []struct {
        name     string
        input    []int
        expected []int
    }{
        
        {
            name:     "normal case",
            input:    []int{1, 2, 3, 4, 5},
            expected: []int{5, 4, 3, 2, 1},
        },
        {
            name:     "two elements",
            input:    []int{1, 2},
            expected: []int{2, 1},
        },

        
        {
            name:     "empty input",
            input:    []int{},
            expected: []int{},
        },
        {
            name:     "single element",
            input:    []int{42},
            expected: []int{42},
        },
        {
            name:     "already reversed",
            input:    []int{5, 4, 3, 2, 1},
            expected: []int{1, 2, 3, 4, 5},
        },
        {
            name:     "already sorted",
            input:    []int{1, 2, 3, 4, 5},
            expected: []int{5, 4, 3, 2, 1},
        },

        
        {
            name:     "all same elements",
            input:    []int{7, 7, 7, 7},
            expected: []int{7, 7, 7, 7},
        },
        {
            name:     "palindrome sequence",
            input:    []int{1, 2, 3, 2, 1},
            expected: []int{1, 2, 3, 2, 1},
        },
        {
            name:     "negative numbers",
            input:    []int{-3, -2, -1},
            expected: []int{-1, -2, -3},
        },
        {
            name:     "mixed negative and positive",
            input:    []int{-1, 0, 1},
            expected: []int{1, 0, -1},
        },
        {
            name:     "duplicate values",
            input:    []int{1, 2, 2, 3, 1},
            expected: []int{1, 3, 2, 2, 1},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            original := slices.Clone(tt.input) 

            var result []int
            reverseIter := iterx.Reverse(context.Background(), seqToSeq2(slices.Values(tt.input)))
            iterx.Drain(context.Background(), reverseIter, func(number int) error {
                result = append(result, number)
                return nil
            })

            
            if result == nil {
                result = []int{}
            }
            if !slices.Equal(result, tt.expected) {
                t.Fatalf("expected %v, got %v", tt.expected, result)
            }

            
            if !slices.Equal(tt.input, original) {
                t.Fatalf("original input was mutated: expected %v, got %v", original, tt.input)
            }
        })
    }
}