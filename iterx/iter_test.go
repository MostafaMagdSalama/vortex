package iterx_test

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"slices"
	"testing"

	vinterx "github.com/MostafaMagdSalama/vortex/interx"
)

func TestFilter(t *testing.T) {
	var result []int
	for v := range vinterx.Filter(context.Background(), slices.Values([]int{1, 2, 3, 4, 5}), func(n int) bool { return n > 2 }) {
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
	for v := range vinterx.Filter(ctx, slices.Values([]int{1, 2, 3, 4, 5}), func(n int) bool { return true }) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestMap(t *testing.T) {
	var result []int
	for v := range vinterx.Map(context.Background(), slices.Values([]int{1, 2, 3}), func(n int) int { return n * 2 }) {
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
	for v := range vinterx.Map(ctx, slices.Values([]int{1, 2, 3}), func(n int) int { return n * 2 }) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestTake(t *testing.T) {
	var result []int
	for v := range vinterx.Take(context.Background(), slices.Values([]int{1, 2, 3, 4}), 2) {
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
	for v := range vinterx.Take(ctx, slices.Values([]int{1, 2, 3, 4}), 2) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestFlatMap(t *testing.T) {
	var result []int
	for v := range vinterx.FlatMap(context.Background(), slices.Values([]int{1, 2, 3}), func(n int) iter.Seq[int] {
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
	for v := range vinterx.FlatMap(ctx, slices.Values([]int{1, 2, 3}), func(n int) iter.Seq[int] {
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
	for v := range vinterx.TakeWhile(context.Background(), slices.Values([]int{1, 2, 3, 0, 4}), func(n int) bool { return n > 0 }) {
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
	for v := range vinterx.TakeWhile(ctx, slices.Values([]int{1, 2, 3}), func(n int) bool { return true }) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestZip(t *testing.T) {
	var result [][2]any
	for pair := range vinterx.Zip(context.Background(), slices.Values([]int{1, 2, 3}), slices.Values([]string{"a", "b"})) {
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
	for pair := range vinterx.Zip(ctx, slices.Values([]int{1, 2, 3}), slices.Values([]string{"a", "b"})) {
		result = append(result, pair)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestChunk_EvenSplit(t *testing.T) {
	var result [][]int
	for batch := range vinterx.Chunk(context.Background(), slices.Values([]int{1, 2, 3, 4, 5, 6}), 2) {
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
	for batch := range vinterx.Chunk(ctx, slices.Values([]int{1, 2, 3, 4}), 2) {
		result = append(result, batch)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestFlatten(t *testing.T) {
	var result []int
	for v := range vinterx.Flatten(context.Background(), slices.Values([][]int{{1, 2}, {3}, {4, 5}})) {
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
	for v := range vinterx.Flatten(ctx, slices.Values([][]int{{1, 2}, {3}})) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestDistinct(t *testing.T) {
	var result []int
	for v := range vinterx.Distinct(context.Background(), slices.Values([]int{1, 2, 1, 3, 2, 4})) {
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
	for v := range vinterx.Distinct(ctx, slices.Values([]int{1, 2, 1, 3})) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestContains(t *testing.T) {
	if !vinterx.Contains(context.Background(), slices.Values([]int{1, 2, 3, 4}), 3) {
		t.Fatal("expected true")
	}
}

func TestContains_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if vinterx.Contains(ctx, slices.Values([]int{1, 2, 3, 4}), 3) {
		t.Fatal("expected false on cancelled context")
	}
}

func TestForEach(t *testing.T) {
	var result []int
	vinterx.ForEach(context.Background(), slices.Values([]int{1, 2, 3}), func(v int) {
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
	vinterx.ForEach(ctx, slices.Values([]int{1, 2, 3}), func(v int) {
		calls++
	})

	if calls != 0 {
		t.Fatalf("expected 0 calls on cancelled context, got %d", calls)
	}
}

func TestReverse(t *testing.T) {
	var result []int
	for v := range vinterx.Reverse(context.Background(), slices.Values([]int{1, 2, 3, 4, 5})) {
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
	for v := range vinterx.Reverse(ctx, slices.Values([]int{1, 2, 3, 4, 5})) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func ExampleFilter() {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	for v := range vinterx.Filter(context.Background(), numbers, func(n int) bool {
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

	for v := range vinterx.Map(context.Background(), numbers, func(n int) int {
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

	for v := range vinterx.Take(context.Background(), numbers, 3) {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// 3
}

func ExampleFlatMap() {
	numbers := slices.Values([]int{1, 2, 3})

	for v := range vinterx.FlatMap(context.Background(), numbers, func(n int) iter.Seq[int] {
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

	for v := range vinterx.TakeWhile(context.Background(), numbers, func(n int) bool {
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

	for pair := range vinterx.Zip(context.Background(), left, right) {
		fmt.Println(pair[0], pair[1])
	}
	// Output:
	// 1 a
	// 2 b
	// 3 c
}

func ExampleChunk() {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	for batch := range vinterx.Chunk(context.Background(), numbers, 2) {
		fmt.Println(batch)
	}
	// Output:
	// [1 2]
	// [3 4]
	// [5]
}

func ExampleFlatten() {
	groups := slices.Values([][]int{{1, 2}, {3}, {4, 5}})

	for v := range vinterx.Flatten(context.Background(), groups) {
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

	for v := range vinterx.Distinct(context.Background(), numbers) {
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

	fmt.Println(vinterx.Contains(context.Background(), numbers, 3))
	// Output:
	// true
}

func ExampleForEach() {
	numbers := slices.Values([]int{1, 2, 3})

	vinterx.ForEach(context.Background(), numbers, func(v int) {
		fmt.Println(v * 2)
	})
	// Output:
	// 2
	// 4
	// 6
}

func ExampleReverse() {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	for v := range vinterx.Reverse(context.Background(), numbers) {
		fmt.Println(v)
	}
	// Output:
	// 5
	// 4
	// 3
	// 2
	// 1
}

func TestDrain_Basic(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3})
	var result []int

	err := vinterx.Drain(context.Background(), numbers, func(n int) error {
		result = append(result, n)
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
}

func TestDrain_StopsOnError(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})
	count := 0

	err := vinterx.Drain(context.Background(), numbers, func(n int) error {
		count++
		if n == 3 {
			return errors.New("stop here")
		}
		return nil
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if count != 3 {
		t.Fatalf("expected 3 calls, got %d", count)
	}
}

func TestDrain_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	numbers := slices.Values([]int{1, 2, 3})
	count := 0

	err := vinterx.Drain(ctx, numbers, func(n int) error {
		count++
		return nil
	})

	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	if count != 0 {
		t.Fatalf("expected 0 calls, got %d", count)
	}
}

func ExampleDrain() {
	numbers := slices.Values([]int{1, 2, 3})

	_ = vinterx.Drain(context.Background(), numbers, func(n int) error {
		fmt.Println(n)
		return nil
	})
	// Output:
	// 1
	// 2
	// 3
}
