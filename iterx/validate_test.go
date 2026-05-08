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

func TestValidationError_Error(t *testing.T) {
	e := iterx.ValidationError[int]{Item: 42, Reason: "must be even"}
	got := e.Error()
	if got == "" {
		t.Fatal("expected non-empty error string")
	}
	if !contains(got, "42") || !contains(got, "must be even") {
		t.Fatalf("expected error to mention item and reason, got: %s", got)
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	e := iterx.ValidationError[string]{Item: "foo", Reason: "invalid"}
	if !errors.Is(e, vortex.ErrValidation) {
		t.Fatalf("expected errors.Is(e, ErrValidation) to be true")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name           string
		input          []int
		validator      func(int) (bool, string)
		expectedResult []int
		expectedErrors int
	}{
		{
			name:           "all valid",
			input:          []int{2, 4, 6},
			validator:      func(n int) (bool, string) { return n%2 == 0, "must be even" },
			expectedResult: []int{2, 4, 6},
			expectedErrors: 0,
		},
		{
			name:           "mixed valid and invalid",
			input:          []int{1, 2, 3, 4},
			validator:      func(n int) (bool, string) { return n%2 == 0, "must be even" },
			expectedResult: []int{2, 4},
			expectedErrors: 2,
		},
		{
			name:           "all invalid",
			input:          []int{1, 3, 5},
			validator:      func(n int) (bool, string) { return n%2 == 0, "must be even" },
			expectedResult: []int{},
			expectedErrors: 3,
		},
		{
			name:           "empty input",
			input:          []int{},
			validator:      func(n int) (bool, string) { return true, "" },
			expectedResult: []int{},
			expectedErrors: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errs []iterx.ValidationError[int]
			var got []int
			seq := seqToSeq2(slices.Values(tt.input))
			validated := iterx.Validate(context.Background(), seq, tt.validator, func(e iterx.ValidationError[int]) {
				errs = append(errs, e)
			})
			for v, err := range validated {
				if err != nil {
					t.Fatalf("unexpected seq error: %v", err)
				}
				got = append(got, v)
			}
			if got == nil {
				got = []int{}
			}
			if !slices.Equal(got, tt.expectedResult) {
				t.Fatalf("expected %v, got %v", tt.expectedResult, got)
			}
			if len(errs) != tt.expectedErrors {
				t.Fatalf("expected %d validation errors, got %d", tt.expectedErrors, len(errs))
			}
		})
	}
}

func TestValidate_PropagatesError(t *testing.T) {
	sentinelErr := errors.New("stream error")
	seq := func(yield func(int, error) bool) {
		if !yield(2, nil) {
			return
		}
		if !yield(0, sentinelErr) {
			return
		}
		yield(4, nil)
	}
	var got []int
	var gotErr error
	for v, err := range iterx.Validate(context.Background(), seq, func(n int) (bool, string) { return n%2 == 0, "odd" }, nil) {
		if err != nil {
			gotErr = err
			break
		}
		got = append(got, v)
	}
	if !errors.Is(gotErr, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", gotErr)
	}
	if !slices.Equal(got, []int{2}) {
		t.Fatalf("expected [2] before error, got %v", got)
	}
}

func TestValidate_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var gotErr error
	for _, err := range iterx.Validate(ctx, seqToSeq2(slices.Values([]int{1, 2, 3})),
		func(n int) (bool, string) { return true, "" }, nil) {
		if err != nil {
			gotErr = err
			break
		}
	}
	if !errors.Is(gotErr, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", gotErr)
	}
}

func TestValidate_NilOnError(t *testing.T) {
	// nil onError callback must not panic when an item fails validation
	var got []int
	seq := seqToSeq2(slices.Values([]int{1, 2, 3}))
	for v, err := range iterx.Validate(context.Background(), seq,
		func(n int) (bool, string) { return n%2 == 0, "odd" }, nil) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	if !slices.Equal(got, []int{2}) {
		t.Fatalf("expected [2], got %v", got)
	}
}

func TestValidate_StopsEarly(t *testing.T) {
	var got []int
	seq := seqToSeq2(slices.Values([]int{2, 4, 6, 8}))
	for v, err := range iterx.Validate(context.Background(), seq,
		func(n int) (bool, string) { return true, "" }, nil) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{2, 4}) {
		t.Fatalf("expected [2 4], got %v", got)
	}
}
