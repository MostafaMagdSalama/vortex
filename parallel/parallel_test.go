package parallel_test

import (
	"slices"
	"sort"
	"testing"

	"github.com/MostafaMagdSalama/vortex/parallel"
)

func TestParallelMap(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})

	var results []int
	for v := range parallel.ParallelMap(input, func(n int) int { return n * 2 }, 3) {
		results = append(results, v)
	}

	// sort because order is not guaranteed
	sort.Ints(results)

	if len(results) != 5 || results[0] != 2 || results[4] != 10 {
		t.Fatalf("got %v", results)
	}
}

func TestBatchMap(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})

	var results []int
	for v := range parallel.BatchMap(input, func(batch []int) []int {
		out := make([]int, len(batch))
		for i, v := range batch {
			out[i] = v * 2
		}
		return out
	}, 2) {
		results = append(results, v)
	}

	if len(results) != 5 {
		t.Fatalf("got %v", results)
	}
}

func TestWorkerPoolMap(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})

	var results []int
	for v := range parallel.WorkerPoolMap(input, func(n int) int { return n * 2 }, 3) {
		results = append(results, v)
	}

	sort.Ints(results)

	if len(results) != 5 || results[0] != 2 || results[4] != 10 {
		t.Fatalf("got %v", results)
	}
}

func TestParallelMap_EarlyStop(t *testing.T) {
	// large input so workers are definitely still running when we stop
	input := slices.Values(make([]int, 10000))

	count := 0
	for range parallel.ParallelMap(
		input,
		func(n int) int { return n * 2 },
		8,
	) {
		count++
		if count == 10 {
			break // stop early
		}
	}

	if count != 10 {
		t.Fatalf("expected 10, got %d", count)
	}
	// if deadlock occurs, test will hang forever and timeout
}


func TestWorkerPoolMap_EarlyStop(t *testing.T) {
	input := make([]int, 10000)

	count := 0
	for range parallel.WorkerPoolMap(
		slices.Values(input),
		func(n int) int { return n * 2 },
		8,
	) {
		count++
		if count == 10 {
			break
		}
	}

	if count != 10 {
		t.Fatalf("expected 10, got %d", count)
	}
}
