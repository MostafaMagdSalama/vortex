package iterx_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestValidateSeq(t *testing.T) {
	tests := []struct {
		name           string
		input          []int
		validator      func(int) (bool, string)
		expectedResult []int
		expectedErrors []iterx.ValidationError[int]
	}{
		{
			name:  "all elements valid",
			input: []int{2, 4, 6, 8},
			validator: func(n int) (bool, string) {
				return n%2 == 0, "must be even"
			},
			expectedResult: []int{2, 4, 6, 8},
			expectedErrors: []iterx.ValidationError[int]{},
		},
		{
			name:  "all elements invalid",
			input: []int{1, 3, 5, 7},
			validator: func(n int) (bool, string) {
				return n%2 == 0, "must be even"
			},
			expectedResult: []int{},
			expectedErrors: []iterx.ValidationError[int]{
				{Item: 1, Reason: "must be even"},
				{Item: 3, Reason: "must be even"},
				{Item: 5, Reason: "must be even"},
				{Item: 7, Reason: "must be even"},
			},
		},
		{
			name:  "mixed valid and invalid",
			input: []int{1, 2, 3, 4, 5},
			validator: func(n int) (bool, string) {
				return n%2 == 0, "must be even"
			},
			expectedResult: []int{2, 4},
			expectedErrors: []iterx.ValidationError[int]{
				{Item: 1, Reason: "must be even"},
				{Item: 3, Reason: "must be even"},
				{Item: 5, Reason: "must be even"},
			},
		},

		{
			name:  "empty input",
			input: []int{},
			validator: func(n int) (bool, string) {
				return n > 0, "must be positive"
			},
			expectedResult: []int{},
			expectedErrors: []iterx.ValidationError[int]{},
		},
		{
			name:  "single element valid",
			input: []int{5},
			validator: func(n int) (bool, string) {
				return n > 0, "must be positive"
			},
			expectedResult: []int{5},
			expectedErrors: []iterx.ValidationError[int]{},
		},
		{
			name:  "single element invalid",
			input: []int{-1},
			validator: func(n int) (bool, string) {
				return n > 0, "must be positive"
			},
			expectedResult: []int{},
			expectedErrors: []iterx.ValidationError[int]{
				{Item: -1, Reason: "must be positive"},
			},
		},

		{
			name:  "validator always true",
			input: []int{1, 2, 3},
			validator: func(n int) (bool, string) {
				return true, ""
			},
			expectedResult: []int{1, 2, 3},
			expectedErrors: []iterx.ValidationError[int]{},
		},
		{
			name:  "validator always false",
			input: []int{1, 2, 3},
			validator: func(n int) (bool, string) {
				return false, "always invalid"
			},
			expectedResult: []int{},
			expectedErrors: []iterx.ValidationError[int]{
				{Item: 1, Reason: "always invalid"},
				{Item: 2, Reason: "always invalid"},
				{Item: 3, Reason: "always invalid"},
			},
		},
		{
			name:  "error Reason is preserved correctly",
			input: []int{-1, 2, -3},
			validator: func(n int) (bool, string) {
				if n < 0 {
					return false, fmt.Sprintf("%d is negative", n)
				}
				return true, ""
			},
			expectedResult: []int{2},
			expectedErrors: []iterx.ValidationError[int]{
				{Item: -1, Reason: "-1 is negative"},
				{Item: -3, Reason: "-3 is negative"},
			},
		},

		{
			name:  "valid elements preserve order",
			input: []int{5, 1, 4, 2, 3},
			validator: func(n int) (bool, string) {
				return n > 2, "must be greater than 2"
			},
			expectedResult: []int{5, 4, 3},
			expectedErrors: []iterx.ValidationError[int]{
				{Item: 1, Reason: "must be greater than 2"},
				{Item: 2, Reason: "must be greater than 2"},
			},
		},
		{
			name:  "errors preserve order of occurrence",
			input: []int{1, 2, 3, 4, 5},
			validator: func(n int) (bool, string) {
				return n%2 == 0, "must be even"
			},
			expectedResult: []int{2, 4},
			expectedErrors: []iterx.ValidationError[int]{
				{Item: 1, Reason: "must be even"},
				{Item: 3, Reason: "must be even"},
				{Item: 5, Reason: "must be even"},
			},
		},

		{
			name:  "context cancelled stops iteration",
			input: []int{1, 2, 3, 4, 5},
			validator: func(n int) (bool, string) {
				return true, ""
			},
			expectedResult: []int{},
			expectedErrors: []iterx.ValidationError[int]{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectedErrors := []iterx.ValidationError[int]{}

			onError := func(e iterx.ValidationError[int]) {
				collectedErrors = append(collectedErrors, e)
			}

			ctx := context.Background()

			if tt.name == "context cancelled stops iteration" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			result := []int{}
			validSeq := iterx.ValidateSeq(ctx, slices.Values(tt.input), tt.validator, onError)
			iterx.Drain(ctx, seqToSeq2(validSeq), func(n int) error {
				result = append(result, n)
				return nil
			})

			if !slices.Equal(result, tt.expectedResult) {
				t.Fatalf("result: expected %v, got %v", tt.expectedResult, result)
			}

			if len(collectedErrors) != len(tt.expectedErrors) {
				t.Fatalf("errors count: expected %d, got %d", len(tt.expectedErrors), len(collectedErrors))
			}
			for i := range tt.expectedErrors {
				if collectedErrors[i] != tt.expectedErrors[i] {
					t.Fatalf("error[%d]: expected %+v, got %+v", i, tt.expectedErrors[i], collectedErrors[i])
				}
			}
		})
	}
}
