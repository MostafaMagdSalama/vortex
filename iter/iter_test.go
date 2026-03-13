package iter_test

import (
	"context"
	"iter"
	"slices"
	"testing"

	viter "github.com/MostafaMagdSalama/vortex/iter"
)

func TestFilter(t *testing.T) {
	var result []int
	for v := range viter.Filter(context.Background(), slices.Values([]int{1, 2, 3, 4, 5}), func(n int) bool { return n > 2 }) {
		result = append(result, v)
	}

	if !slices.Equal(result, []int{3, 4, 5}) {
		t.Fatalf("got %v", result)
	}
}

func TestFilter_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result []int
	for v := range viter.Filter(ctx, slices.Values([]int{1, 2, 3, 4, 5}), func(n int) bool { return true }) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestMap(t *testing.T) {
	var result []int
	for v := range viter.Map(context.Background(), slices.Values([]int{1, 2, 3}), func(n int) int { return n * 2 }) {
		result = append(result, v)
	}

	if !slices.Equal(result, []int{2, 4, 6}) {
		t.Fatalf("got %v", result)
	}
}

func TestMap_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result []int
	for v := range viter.Map(ctx, slices.Values([]int{1, 2, 3}), func(n int) int { return n * 2 }) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestTake(t *testing.T) {
	var result []int
	for v := range viter.Take(context.Background(), slices.Values([]int{1, 2, 3, 4}), 2) {
		result = append(result, v)
	}

	if !slices.Equal(result, []int{1, 2}) {
		t.Fatalf("got %v", result)
	}
}

func TestTake_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result []int
	for v := range viter.Take(ctx, slices.Values([]int{1, 2, 3, 4}), 2) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestFlatMap(t *testing.T) {
	var result []int
	for v := range viter.FlatMap(context.Background(), slices.Values([]int{1, 2, 3}), func(n int) iter.Seq[int] {
		return slices.Values([]int{n, n * 10})
	}) {
		result = append(result, v)
	}

	if !slices.Equal(result, []int{1, 10, 2, 20, 3, 30}) {
		t.Fatalf("got %v", result)
	}
}

func TestFlatMap_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result []int
	for v := range viter.FlatMap(ctx, slices.Values([]int{1, 2, 3}), func(n int) iter.Seq[int] {
		return slices.Values([]int{n, n * 10})
	}) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestTakeWhile(t *testing.T) {
	var result []int
	for v := range viter.TakeWhile(context.Background(), slices.Values([]int{1, 2, 3, 0, 4}), func(n int) bool { return n > 0 }) {
		result = append(result, v)
	}

	if !slices.Equal(result, []int{1, 2, 3}) {
		t.Fatalf("got %v", result)
	}
}

func TestTakeWhile_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result []int
	for v := range viter.TakeWhile(ctx, slices.Values([]int{1, 2, 3}), func(n int) bool { return true }) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestZip(t *testing.T) {
	var result [][2]any
	for pair := range viter.Zip(context.Background(), slices.Values([]int{1, 2, 3}), slices.Values([]string{"a", "b"})) {
		result = append(result, pair)
	}

	if len(result) != 2 || result[0] != [2]any{1, "a"} || result[1] != [2]any{2, "b"} {
		t.Fatalf("got %v", result)
	}
}

func TestZip_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result [][2]any
	for pair := range viter.Zip(ctx, slices.Values([]int{1, 2, 3}), slices.Values([]string{"a", "b"})) {
		result = append(result, pair)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestChunk_EvenSplit(t *testing.T) {
	var result [][]int
	for batch := range viter.Chunk(context.Background(), slices.Values([]int{1, 2, 3, 4, 5, 6}), 2) {
		result = append(result, batch)
	}

	if len(result) != 3 || !slices.Equal(result[0], []int{1, 2}) || !slices.Equal(result[2], []int{5, 6}) {
		t.Fatalf("got %v", result)
	}
}

func TestChunk_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result [][]int
	for batch := range viter.Chunk(ctx, slices.Values([]int{1, 2, 3, 4}), 2) {
		result = append(result, batch)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestFlatten(t *testing.T) {
	var result []int
	for v := range viter.Flatten(context.Background(), slices.Values([][]int{{1, 2}, {3}, {4, 5}})) {
		result = append(result, v)
	}

	if !slices.Equal(result, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("got %v", result)
	}
}

func TestFlatten_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result []int
	for v := range viter.Flatten(ctx, slices.Values([][]int{{1, 2}, {3}})) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestDistinct(t *testing.T) {
	var result []int
	for v := range viter.Distinct(context.Background(), slices.Values([]int{1, 2, 1, 3, 2, 4})) {
		result = append(result, v)
	}

	if !slices.Equal(result, []int{1, 2, 3, 4}) {
		t.Fatalf("got %v", result)
	}
}

func TestDistinct_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result []int
	for v := range viter.Distinct(ctx, slices.Values([]int{1, 2, 1, 3})) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestContains(t *testing.T) {
	if !viter.Contains(context.Background(), slices.Values([]int{1, 2, 3, 4}), 3) {
		t.Fatal("expected true")
	}
}

func TestContains_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if viter.Contains(ctx, slices.Values([]int{1, 2, 3, 4}), 3) {
		t.Fatal("expected false on cancelled context")
	}
}

func TestForEach(t *testing.T) {
	var result []int
	viter.ForEach(context.Background(), slices.Values([]int{1, 2, 3}), func(v int) {
		result = append(result, v*2)
	})

	if !slices.Equal(result, []int{2, 4, 6}) {
		t.Fatalf("got %v", result)
	}
}

func TestForEach_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	viter.ForEach(ctx, slices.Values([]int{1, 2, 3}), func(v int) {
		calls++
	})

	if calls != 0 {
		t.Fatalf("expected 0 calls on cancelled context, got %d", calls)
	}
}

func TestReverse(t *testing.T) {
	var result []int
	for v := range viter.Reverse(context.Background(), slices.Values([]int{1, 2, 3, 4, 5})) {
		result = append(result, v)
	}

	if !slices.Equal(result, []int{5, 4, 3, 2, 1}) {
		t.Fatalf("got %v", result)
	}
}

func TestReverse_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result []int
	for v := range viter.Reverse(ctx, slices.Values([]int{1, 2, 3, 4, 5})) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}
