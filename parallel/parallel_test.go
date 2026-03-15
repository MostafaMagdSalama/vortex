package parallel_test

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/MostafaMagdSalama/vortex/parallel"
)

func seq2FromSlice[T any](items []T) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for _, item := range items {
			if !yield(item, nil) {
				return
			}
		}
	}
}

func seq2WithError[T any](items []T, errAt int, err error) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for i, item := range items {
			if i == errAt {
				var zero T
				if !yield(zero, err) {
					return
				}
				continue
			}
			if !yield(item, nil) {
				return
			}
		}
	}
}

func TestParallelMapSeq(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})

	var results []int
	for v := range parallel.ParallelMapSeq(context.Background(), input, func(n int) int { return n * 2 }, 3) {
		results = append(results, v)
	}

	sort.Ints(results)
	if !slices.Equal(results, []int{2, 4, 6, 8, 10}) {
		t.Fatalf("got %v", results)
	}
}

func TestParallelMapSeq_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var results []int
	for v := range parallel.ParallelMapSeq(ctx, slices.Values([]int{1, 2, 3, 4, 5}), func(n int) int { return n * 2 }, 3) {
		results = append(results, v)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(results))
	}
}

func TestParallelMap(t *testing.T) {
	input := seq2FromSlice([]int{1, 2, 3, 4, 5})

	var results []int
	for v, err := range parallel.ParallelMap(context.Background(), input, func(n int) int { return n * 2 }, 3) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		results = append(results, v)
	}

	sort.Ints(results)
	if !slices.Equal(results, []int{2, 4, 6, 8, 10}) {
		t.Fatalf("got %v", results)
	}
}

func TestParallelMap_PassesThroughErrors(t *testing.T) {
	input := seq2WithError([]int{1, 2, 3}, 1, errors.New("boom"))

	var results []int
	var errs []error
	for v, err := range parallel.ParallelMap(context.Background(), input, func(n int) int { return n * 10 }, 2) {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		results = append(results, v)
	}

	sort.Ints(results)
	if !slices.Equal(results, []int{10, 30}) {
		t.Fatalf("got results %v", results)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestBatchMapSeq(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})

	var results []int
	for v := range parallel.BatchMapSeq(context.Background(), input, func(batch []int) []int {
		out := make([]int, len(batch))
		for i, value := range batch {
			out[i] = value * 2
		}
		return out
	}, 2) {
		results = append(results, v)
	}

	if !slices.Equal(results, []int{2, 4, 6, 8, 10}) {
		t.Fatalf("got %v", results)
	}
}

func TestBatchMapSeq_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var results []int
	for v := range parallel.BatchMapSeq(ctx, slices.Values([]int{1, 2, 3, 4, 5}), func(batch []int) []int {
		return batch
	}, 2) {
		results = append(results, v)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(results))
	}
}

func TestBatchMap(t *testing.T) {
	input := seq2FromSlice([]int{1, 2, 3, 4, 5})

	var results []int
	for v, err := range parallel.BatchMap(context.Background(), input, func(batch []int) []int {
		out := make([]int, len(batch))
		for i, value := range batch {
			out[i] = value * 2
		}
		return out
	}, 2) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		results = append(results, v)
	}

	if !slices.Equal(results, []int{2, 4, 6, 8, 10}) {
		t.Fatalf("got %v", results)
	}
}

func TestBatchMap_PassesThroughErrors(t *testing.T) {
	input := seq2WithError([]int{1, 2, 3, 4}, 2, errors.New("bad row"))

	type output struct {
		value int
		err   error
	}

	var got []output
	for v, err := range parallel.BatchMap(context.Background(), input, func(batch []int) []int {
		out := make([]int, len(batch))
		for i, n := range batch {
			out[i] = n * 10
		}
		return out
	}, 2) {
		got = append(got, output{value: v, err: err})
	}

	if len(got) != 4 {
		t.Fatalf("expected 4 outputs, got %d", len(got))
	}
	if got[0].value != 10 || got[1].value != 20 || got[2].err == nil || got[3].value != 40 {
		t.Fatalf("unexpected outputs: %+v", got)
	}
}

func TestParallelMapSeq_EarlyStop(t *testing.T) {
	count := 0
	for range parallel.ParallelMapSeq(context.Background(), slices.Values(make([]int, 10000)), func(n int) int { return n * 2 }, 8) {
		count++
		if count == 10 {
			break
		}
	}

	if count != 10 {
		t.Fatalf("expected 10, got %d", count)
	}
}

func ExampleParallelMapSeq() {
	numbers := slices.Values([]int{1, 2, 3})

	for v := range parallel.ParallelMapSeq(context.Background(), numbers, func(n int) int {
		return n * 2
	}, 1) {
		fmt.Println(v)
	}
	// Output:
	// 2
	// 4
	// 6
}

func ExampleParallelMap() {
	numbers := seq2FromSlice([]int{1, 2, 3})

	for v, err := range parallel.ParallelMap(context.Background(), numbers, func(n int) int {
		return n * 2
	}, 1) {
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println(v)
	}
	// Output:
	// 2
	// 4
	// 6
}

func ExampleBatchMapSeq() {
	numbers := slices.Values([]int{1, 2, 3, 4})

	for v := range parallel.BatchMapSeq(context.Background(), numbers, func(batch []int) []int {
		out := make([]int, len(batch))
		for i, n := range batch {
			out[i] = n * 10
		}
		return out
	}, 2) {
		fmt.Println(v)
	}
	// Output:
	// 10
	// 20
	// 30
	// 40
}

func ExampleOrderedParallelMapSeq() {
	ctx := context.Background()
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	for v := range parallel.OrderedParallelMapSeq(ctx, numbers, func(n int) int {
		return n * 2
	}, 3) {
		fmt.Println(v)
	}
	// Output:
	// 2
	// 4
	// 6
	// 8
	// 10
}

func TestOrderedParallelMapSeq_Order(t *testing.T) {
	ctx := context.Background()
	input := make([]int, 100)
	for i := range input {
		input[i] = i
	}

	var result []int
	for v := range parallel.OrderedParallelMapSeq(ctx,
		slices.Values(input),
		func(n int) int {
			if n%2 == 0 {
				time.Sleep(time.Millisecond)
			}
			return n * 2
		},
		8,
	) {
		result = append(result, v)
	}

	if len(result) != 100 {
		t.Fatalf("expected 100 results, got %d", len(result))
	}
	for i, v := range result {
		if v != i*2 {
			t.Fatalf("index %d: expected %d, got %d", i, i*2, v)
		}
	}
}

func TestOrderedParallelMap(t *testing.T) {
	ctx := context.Background()
	numbers := seq2FromSlice([]int{5, 3, 1, 4, 2})

	var result []int
	for v, err := range parallel.OrderedParallelMap(ctx, numbers, func(n int) int {
		return n * 10
	}, 3) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, v)
	}

	expected := []int{50, 30, 10, 40, 20}
	for i, v := range result {
		if v != expected[i] {
			t.Fatalf("index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestOrderedParallelMap_PreservesErrorOrder(t *testing.T) {
	ctx := context.Background()
	input := seq2WithError([]int{1, 2, 3, 4}, 1, errors.New("bad row"))

	type output struct {
		value int
		err   error
	}

	var got []output
	for v, err := range parallel.OrderedParallelMap(ctx, input, func(n int) int {
		if n%2 == 0 {
			time.Sleep(time.Millisecond)
		}
		return n * 10
	}, 3) {
		got = append(got, output{value: v, err: err})
	}

	if len(got) != 4 {
		t.Fatalf("expected 4 outputs, got %d", len(got))
	}
	if got[0].value != 10 || got[1].err == nil || got[2].value != 30 || got[3].value != 40 {
		t.Fatalf("unexpected outputs: %+v", got)
	}
}

func TestOrderedParallelMapSeq_SingleWorker(t *testing.T) {
	ctx := context.Background()
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	var result []int
	for v := range parallel.OrderedParallelMapSeq(ctx, numbers, func(n int) int {
		return n * 2
	}, 1) {
		result = append(result, v)
	}

	expected := []int{2, 4, 6, 8, 10}
	for i, v := range result {
		if v != expected[i] {
			t.Fatalf("index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestOrderedParallelMapSeq_Empty(t *testing.T) {
	ctx := context.Background()

	var result []int
	for v := range parallel.OrderedParallelMapSeq(ctx,
		slices.Values([]int{}),
		func(n int) int { return n },
		4,
	) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0, got %d", len(result))
	}
}

func TestOrderedParallelMapSeq_EarlyStop(t *testing.T) {
	ctx := context.Background()
	numbers := slices.Values([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	count := 0
	for range parallel.OrderedParallelMapSeq(ctx, numbers, func(n int) int {
		return n
	}, 4) {
		count++
		if count == 3 {
			break
		}
	}

	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}
}

func TestOrderedParallelMapSeq_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result []int
	for v := range parallel.OrderedParallelMapSeq(ctx,
		slices.Values([]int{1, 2, 3, 4, 5}),
		func(n int) int { return n },
		4,
	) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 on cancelled context, got %d", len(result))
	}
}

func TestOrderedParallelMapSeq_Race(t *testing.T) {
	ctx := context.Background()
	input := make([]int, 1000)
	for i := range input {
		input[i] = i
	}

	var result []int
	for v := range parallel.OrderedParallelMapSeq(ctx,
		slices.Values(input),
		func(n int) int { return n * 2 },
		16,
	) {
		result = append(result, v)
	}

	if len(result) != 1000 {
		t.Fatalf("expected 1000, got %d", len(result))
	}
}

func ExampleOrderedParallelMapSeq_strings() {
	ctx := context.Background()
	words := slices.Values([]string{"hello", "world", "foo"})

	for v := range parallel.OrderedParallelMapSeq(ctx, words, strings.ToUpper, 2) {
		fmt.Println(v)
	}
	// Output:
	// HELLO
	// WORLD
	// FOO
}
