package parallel_test

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"testing"
	"time"
	"strings"
	"github.com/MostafaMagdSalama/vortex/parallel"
)

func TestParallelMap(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})

	var results []int
	for v := range parallel.ParallelMap(context.Background(), input, func(n int) int { return n * 2 }, 3) {
		results = append(results, v)
	}

	sort.Ints(results)
	if !slices.Equal(results, []int{2, 4, 6, 8, 10}) {
		t.Fatalf("got %v", results)
	}
}

func TestParallelMap_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var results []int
	for v := range parallel.ParallelMap(ctx, slices.Values([]int{1, 2, 3, 4, 5}), func(n int) int { return n * 2 }, 3) {
		results = append(results, v)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(results))
	}
}

func TestBatchMap(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})

	var results []int
	for v := range parallel.BatchMap(context.Background(), input, func(batch []int) []int {
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

func TestBatchMap_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var results []int
	for v := range parallel.BatchMap(ctx, slices.Values([]int{1, 2, 3, 4, 5}), func(batch []int) []int {
		return batch
	}, 2) {
		results = append(results, v)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(results))
	}
}

// func TestWorkerPoolMap(t *testing.T) {
// 	input := slices.Values([]int{1, 2, 3, 4, 5})

// 	var results []int
// 	for v := range parallel.WorkerPoolMap(context.Background(), input, func(n int) int { return n * 2 }, 3) {
// 		results = append(results, v)
// 	}

// 	sort.Ints(results)
// 	if !slices.Equal(results, []int{2, 4, 6, 8, 10}) {
// 		t.Fatalf("got %v", results)
// 	}
// }

// func TestWorkerPoolMap_Cancelled(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	cancel()

// 	var results []int
// 	for v := range parallel.WorkerPoolMap(ctx, slices.Values([]int{1, 2, 3, 4, 5}), func(n int) int { return n * 2 }, 3) {
// 		results = append(results, v)
// 	}

// 	if len(results) != 0 {
// 		t.Fatalf("expected 0 results on cancelled context, got %d", len(results))
// 	}
// }

func TestParallelMap_EarlyStop(t *testing.T) {
	count := 0
	for range parallel.ParallelMap(context.Background(), slices.Values(make([]int, 10000)), func(n int) int { return n * 2 }, 8) {
		count++
		if count == 10 {
			break
		}
	}

	if count != 10 {
		t.Fatalf("expected 10, got %d", count)
	}
}

// func TestWorkerPoolMap_EarlyStop(t *testing.T) {
// 	count := 0
// 	for range parallel.WorkerPoolMap(context.Background(), slices.Values(make([]int, 10000)), func(n int) int { return n * 2 }, 8) {
// 		count++
// 		if count == 10 {
// 			break
// 		}
// 	}

// 	if count != 10 {
// 		t.Fatalf("expected 10, got %d", count)
// 	}
// }

func ExampleParallelMap() {
	numbers := slices.Values([]int{1, 2, 3})

	for v := range parallel.ParallelMap(context.Background(), numbers, func(n int) int {
		return n * 2
	}, 1) {
		fmt.Println(v)
	}
	// Output:
	// 2
	// 4
	// 6
}

func ExampleBatchMap() {
	numbers := slices.Values([]int{1, 2, 3, 4})

	for v := range parallel.BatchMap(context.Background(), numbers, func(batch []int) []int {
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

// func ExampleWorkerPoolMap() {
// 	numbers := slices.Values([]int{1, 2, 3})

// 	for v := range parallel.WorkerPoolMap(context.Background(), numbers, func(n int) int {
// 		return n * 3
// 	}, 1) {
// 		fmt.Println(v)
// 	}
// 	// Output:
// 	// 3
// 	// 6
// 	// 9
// }

func ExampleOrderedParallelMap() {
	ctx := context.Background()
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	// results always come back in original order
	for v := range parallel.OrderedParallelMap(ctx, numbers, func(n int) int {
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

// results come back in original order regardless of worker count
func TestOrderedParallelMap_Order(t *testing.T) {
	ctx := context.Background()
	input := make([]int, 100)
	for i := range input {
		input[i] = i
	}

	var result []int
	for v := range parallel.OrderedParallelMap(ctx,
		slices.Values(input),
		func(n int) int {
			// simulate variable processing time
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

// results match ParallelMap results but in order
func TestOrderedParallelMap_MatchesInput(t *testing.T) {
	ctx := context.Background()
	numbers := slices.Values([]int{5, 3, 1, 4, 2})

	var result []int
	for v := range parallel.OrderedParallelMap(ctx, numbers, func(n int) int {
		return n * 10
	}, 3) {
		result = append(result, v)
	}

	expected := []int{50, 30, 10, 40, 20}
	for i, v := range result {
		if v != expected[i] {
			t.Fatalf("index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

// single worker — still works correctly
func TestOrderedParallelMap_SingleWorker(t *testing.T) {
	ctx := context.Background()
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	var result []int
	for v := range parallel.OrderedParallelMap(ctx, numbers, func(n int) int {
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

// empty sequence — no results no panic
func TestOrderedParallelMap_Empty(t *testing.T) {
	ctx := context.Background()

	var result []int
	for v := range parallel.OrderedParallelMap(ctx,
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

// early stop — cancels cleanly
func TestOrderedParallelMap_EarlyStop(t *testing.T) {
	ctx := context.Background()
	numbers := slices.Values([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	count := 0
	for range parallel.OrderedParallelMap(ctx, numbers, func(n int) int {
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

// cancelled context — stops immediately
func TestOrderedParallelMap_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result []int
	for v := range parallel.OrderedParallelMap(ctx,
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

// race detector — no data races
func TestOrderedParallelMap_Race(t *testing.T) {
	ctx := context.Background()
	input := make([]int, 1000)
	for i := range input {
		input[i] = i
	}

	var result []int
	for v := range parallel.OrderedParallelMap(ctx,
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

func ExampleOrderedParallelMap_strings() {
	ctx := context.Background()
	words := slices.Values([]string{"hello", "world", "foo"})

	for v := range parallel.OrderedParallelMap(ctx, words, strings.ToUpper, 2) {
		fmt.Println(v)
	}
	// Output:
	// HELLO
	// WORLD
	// FOO
}
