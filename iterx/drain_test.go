package iterx_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestDrain(t *testing.T) {
	tests := []struct {
		name          string
		input         []int
		seqErrors     map[int]error
		fn            func(int) error
		expectedItems []int
		expectedErr   error
		cancelCtx     bool
	}{
		{
			name:          "drains all elements successfully",
			input:         []int{1, 2, 3, 4, 5},
			seqErrors:     map[int]error{},
			fn:            func(n int) error { return nil },
			expectedItems: []int{1, 2, 3, 4, 5},
			expectedErr:   nil,
		},
		{
			name:          "empty sequence",
			input:         []int{},
			seqErrors:     map[int]error{},
			fn:            func(n int) error { return nil },
			expectedItems: []int{},
			expectedErr:   nil,
		},
		{
			name:          "single element success",
			input:         []int{42},
			seqErrors:     map[int]error{},
			fn:            func(n int) error { return nil },
			expectedItems: []int{42},
			expectedErr:   nil,
		},

		{
			name:      "fn error stops iteration",
			input:     []int{1, 2, 3, 4, 5},
			seqErrors: map[int]error{},
			fn: func(n int) error {
				if n == 3 {
					return errors.New("fn failed at 3")
				}
				return nil
			},
			expectedItems: []int{1, 2, 3},
			expectedErr:   errors.New("fn failed at 3"),
		},
		{
			name:      "fn error on first element",
			input:     []int{1, 2, 3},
			seqErrors: map[int]error{},
			fn: func(n int) error {
				return errors.New("fn failed immediately")
			},
			expectedItems: []int{1},
			expectedErr:   errors.New("fn failed immediately"),
		},
		{
			name:      "fn error on last element",
			input:     []int{1, 2, 3},
			seqErrors: map[int]error{},
			fn: func(n int) error {
				if n == 3 {
					return errors.New("fn failed at last")
				}
				return nil
			},
			expectedItems: []int{1, 2, 3},
			expectedErr:   errors.New("fn failed at last"),
		},

		{
			name:  "seq error stops iteration",
			input: []int{1, 2, 3, 4, 5},
			seqErrors: map[int]error{
				2: errors.New("seq error at index 2"),
			},
			fn:            func(n int) error { return nil },
			expectedItems: []int{1, 2},
			expectedErr:   errors.New("seq error at index 2"),
		},
		{
			name:  "seq error at first element",
			input: []int{1, 2, 3},
			seqErrors: map[int]error{
				0: errors.New("seq error at start"),
			},
			fn:            func(n int) error { return nil },
			expectedItems: []int{},
			expectedErr:   errors.New("seq error at start"),
		},
		{
			name:  "seq error at last element",
			input: []int{1, 2, 3},
			seqErrors: map[int]error{
				2: errors.New("seq error at end"),
			},
			fn:            func(n int) error { return nil },
			expectedItems: []int{1, 2},
			expectedErr:   errors.New("seq error at end"),
		},

		{
			name:  "seq error wins over fn error at same element",
			input: []int{1, 2, 3},
			seqErrors: map[int]error{
				1: errors.New("seq error"),
			},
			fn: func(n int) error {
				if n == 2 {
					return errors.New("fn error")
				}
				return nil
			},
			expectedItems: []int{1},
			expectedErr:   errors.New("seq error"),
		},

		{
			name:          "cancelled context stops iteration immediately",
			input:         []int{1, 2, 3, 4, 5},
			seqErrors:     map[int]error{},
			fn:            func(n int) error { return nil },
			expectedItems: []int{},
			expectedErr:   context.Canceled,
			cancelCtx:     true,
		},

		{
			name:          "fn receives elements in order",
			input:         []int{10, 20, 30},
			seqErrors:     map[int]error{},
			fn:            func(n int) error { return nil },
			expectedItems: []int{10, 20, 30},
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			seq := func(yield func(int, error) bool) {
				for i, v := range tt.input {
					err := tt.seqErrors[i]
					if !yield(v, err) {
						return
					}
				}
			}

			collected := []int{}
			wrappedFn := func(n int) error {
				collected = append(collected, n)
				return tt.fn(n)
			}

			err := iterx.Drain(ctx, seq, wrappedFn)

			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.expectedErr)
				}
				if !strings.Contains(err.Error(), tt.expectedErr.Error()) {
					t.Fatalf("expected error containing %q, got %q", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %q", err)
				}
			}

			if !slices.Equal(collected, tt.expectedItems) {
				t.Fatalf("items: expected %v, got %v", tt.expectedItems, collected)
			}
		})
	}
}
