package iter_test

import (
	"context"
	"fmt"
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

func ExampleFilter() {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	for v := range viter.Filter(context.Background(), numbers, func(n int) bool {
		return n > 2
	}) {
		fmt.Println(v)
	}
	// Output:
	// 3
	// 4
	// 5
}

func ExampleMap() {
	numbers := slices.Values([]int{1, 2, 3})

	for v := range viter.Map(context.Background(), numbers, func(n int) int {
		return n * 10
	}) {
		fmt.Println(v)
	}
	// Output:
	// 10
	// 20
	// 30
}

func ExampleTake() {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	for v := range viter.Take(context.Background(), numbers, 3) {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// 3
}

func ExampleFlatMap() {
	numbers := slices.Values([]int{1, 2, 3})

	for v := range viter.FlatMap(context.Background(), numbers, func(n int) iter.Seq[int] {
		return slices.Values([]int{n, n * 10})
	}) {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 10
	// 2
	// 20
	// 3
	// 30
}

func ExampleTakeWhile() {
	numbers := slices.Values([]int{1, 2, 3, 0, 4})

	for v := range viter.TakeWhile(context.Background(), numbers, func(n int) bool {
		return n > 0
	}) {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// 3
}

func ExampleZip() {
	left := slices.Values([]int{1, 2, 3})
	right := slices.Values([]string{"a", "b", "c"})

	for pair := range viter.Zip(context.Background(), left, right) {
		fmt.Println(pair[0], pair[1])
	}
	// Output:
	// 1 a
	// 2 b
	// 3 c
}

func ExampleChunk() {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	for batch := range viter.Chunk(context.Background(), numbers, 2) {
		fmt.Println(batch)
	}
	// Output:
	// [1 2]
	// [3 4]
	// [5]
}

func ExampleFlatten() {
	groups := slices.Values([][]int{{1, 2}, {3}, {4, 5}})

	for v := range viter.Flatten(context.Background(), groups) {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
}

func ExampleDistinct() {
	numbers := slices.Values([]int{1, 2, 1, 3, 2, 4})

	for v := range viter.Distinct(context.Background(), numbers) {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// 3
	// 4
}

func ExampleContains() {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	fmt.Println(viter.Contains(context.Background(), numbers, 3))
	// Output:
	// true
}

func ExampleForEach() {
	numbers := slices.Values([]int{1, 2, 3})

	viter.ForEach(context.Background(), numbers, func(v int) {
		fmt.Println(v * 2)
	})
	// Output:
	// 2
	// 4
	// 6
}

func ExampleReverse() {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	for v := range viter.Reverse(context.Background(), numbers) {
		fmt.Println(v)
	}
	// Output:
	// 5
	// 4
	// 3
	// 2
	// 1
}
