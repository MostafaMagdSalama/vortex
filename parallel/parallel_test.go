package parallel_test

import (
	"context"
	"slices"
	"sort"
	"testing"

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

func TestWorkerPoolMap(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})

	var results []int
	for v := range parallel.WorkerPoolMap(context.Background(), input, func(n int) int { return n * 2 }, 3) {
		results = append(results, v)
	}

	sort.Ints(results)
	if !slices.Equal(results, []int{2, 4, 6, 8, 10}) {
		t.Fatalf("got %v", results)
	}
}

func TestWorkerPoolMap_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var results []int
	for v := range parallel.WorkerPoolMap(ctx, slices.Values([]int{1, 2, 3, 4, 5}), func(n int) int { return n * 2 }, 3) {
		results = append(results, v)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(results))
	}
}

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

func TestWorkerPoolMap_EarlyStop(t *testing.T) {
	count := 0
	for range parallel.WorkerPoolMap(context.Background(), slices.Values(make([]int, 10000)), func(n int) int { return n * 2 }, 8) {
		count++
		if count == 10 {
			break
		}
	}

	if count != 10 {
		t.Fatalf("expected 10, got %d", count)
	}
}
