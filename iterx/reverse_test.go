package iterx_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex"
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
func TestReverseSeq(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{"normal", []int{1, 2, 3}, []int{3, 2, 1}},
		{"single", []int{42}, []int{42}},
		{"empty", []int{}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []int
			for v := range iterx.ReverseSeq(context.Background(), slices.Values(tt.input)) {
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

func TestReverseSeq_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var got []int
	for v := range iterx.ReverseSeq(ctx, slices.Values([]int{1, 2, 3})) {
		got = append(got, v)
	}
	if len(got) != 0 {
		t.Fatalf("expected no output with cancelled context, got %v", got)
	}
}

func TestReverseSeq_StopsEarly(t *testing.T) {
	var got []int
	for v := range iterx.ReverseSeq(context.Background(), slices.Values([]int{1, 2, 3, 4, 5})) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{5, 4}) {
		t.Fatalf("expected [5 4], got %v", got)
	}
}

func TestReverseSeq_CancelledDuringYield(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var got []int
	for v := range iterx.ReverseSeq(ctx, slices.Values([]int{1, 2, 3, 4, 5})) {
		got = append(got, v)
		if len(got) == 2 {
			cancel() // cancel without break — ctx check stops next iteration
		}
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 items before ctx cancel, got %v", got)
	}
}

func TestReverse_PropagatesError(t *testing.T) {
	sentinelErr := errors.New("stream error")
	seq := func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}
		if !yield(0, sentinelErr) {
			return
		}
		yield(2, nil)
	}
	var got []int
	var gotErr error
	for v, err := range iterx.Reverse(context.Background(), seq) {
		if err != nil {
			gotErr = err
			continue
		}
		got = append(got, v)
	}
	if !errors.Is(gotErr, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", gotErr)
	}
	// Reverse collects all items including the error, then yields in reverse order
	// so we get: 2, error, 1 (reversed)
	if !slices.Equal(got, []int{2, 1}) {
		t.Fatalf("expected [2 1] (reversed, error skipped), got %v", got)
	}
}

func TestReverse_CancelledDuringCollection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	count := 0
	seq := func(yield func(int, error) bool) {
		for _, v := range []int{1, 2, 3, 4, 5} {
			count++
			if count == 2 {
				cancel() // cancel before yielding item 2
			}
			if !yield(v, nil) {
				return
			}
		}
	}
	var gotErr error
	for _, err := range iterx.Reverse(ctx, seq) {
		if err != nil {
			gotErr = err
			break
		}
	}
	if !errors.Is(gotErr, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled during collection, got %v", gotErr)
	}
}

func TestReverse_CancelledDuringYield(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var got []int
	var gotErr error
	for v, err := range iterx.Reverse(ctx, seqToSeq2(slices.Values([]int{1, 2, 3}))) {
		if err != nil {
			gotErr = err
			break
		}
		got = append(got, v)
		if len(got) == 1 {
			cancel() // cancel after receiving first reversed item (3)
		}
	}
	if !errors.Is(gotErr, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled during yield, got %v", gotErr)
	}
	if len(got) != 1 || got[0] != 3 {
		t.Fatalf("expected [3] before cancellation, got %v", got)
	}
}

func TestReverse_StopsEarly(t *testing.T) {
	var got []int
	for v, err := range iterx.Reverse(context.Background(), seqToSeq2(slices.Values([]int{1, 2, 3, 4, 5}))) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{5, 4}) {
		t.Fatalf("expected [5 4], got %v", got)
	}
}
