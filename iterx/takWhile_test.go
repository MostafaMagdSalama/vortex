package iterx_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex"
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

func TestTakeWhileSeq(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		cond      func(int) bool
		expected  []int
	}{
		{"takes while under 4", []int{1, 2, 3, 4, 5}, func(n int) bool { return n < 4 }, []int{1, 2, 3}},
		{"always true", []int{1, 2, 3}, func(n int) bool { return true }, []int{1, 2, 3}},
		{"always false", []int{1, 2, 3}, func(n int) bool { return false }, []int{}},
		{"empty input", []int{}, func(n int) bool { return true }, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []int
			for v := range iterx.TakeWhileSeq(context.Background(), slices.Values(tt.input), tt.cond) {
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

func TestTakeWhileSeq_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var got []int
	for v := range iterx.TakeWhileSeq(ctx, slices.Values([]int{1, 2, 3}), func(n int) bool { return true }) {
		got = append(got, v)
	}
	if len(got) != 0 {
		t.Fatalf("expected no output with cancelled context, got %v", got)
	}
}

func TestTakeWhileSeq_StopsEarly(t *testing.T) {
	callCount := 0
	seq := func(yield func(int) bool) {
		for _, v := range []int{1, 2, 3, 4, 5} {
			callCount++
			if !yield(v) {
				return
			}
		}
	}
	var got []int
	for v := range iterx.TakeWhileSeq(context.Background(), seq, func(n int) bool { return true }) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("expected [1 2], got %v", got)
	}
	if callCount > 3 {
		t.Fatalf("expected upstream to stop early, got %d calls", callCount)
	}
}

func TestTakeWhile_PropagatesError(t *testing.T) {
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
	for v, err := range iterx.TakeWhile(context.Background(), seq, func(n int) bool { return true }) {
		if err != nil {
			gotErr = err
			break
		}
		got = append(got, v)
	}
	if !errors.Is(gotErr, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", gotErr)
	}
	if !slices.Equal(got, []int{1}) {
		t.Fatalf("expected [1] before error, got %v", got)
	}
}

func TestTakeWhile_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var gotErr error
	for _, err := range iterx.TakeWhile(ctx, seqToSeq2(slices.Values([]int{1, 2, 3})), func(n int) bool { return true }) {
		if err != nil {
			gotErr = err
			break
		}
	}
	if !errors.Is(gotErr, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", gotErr)
	}
}

func TestTakeWhile_StopsEarly(t *testing.T) {
	var got []int
	for v, err := range iterx.TakeWhile(context.Background(), seqToSeq2(slices.Values([]int{1, 2, 3, 4, 5})), func(n int) bool { return true }) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("expected [1 2], got %v", got)
	}
}

func TestTakeWhile_ErrorsDoNotStopPredicate(t *testing.T) {
	sentinelErr := errors.New("stream error")
	seq := func(yield func(int, error) bool) {
		if !yield(1, nil) {
			return
		}
		if !yield(0, sentinelErr) {
			return
		}
		if !yield(2, nil) {
			return
		}
	}
	var got []int
	for v, err := range iterx.TakeWhile(context.Background(), seq, func(n int) bool { return n < 5 }) {
		if err != nil {
			continue // errors pass through without stopping predicate evaluation
		}
		got = append(got, v)
	}
	if !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("expected [1 2] (error skipped, predicate continues), got %v", got)
	}
}
