package iterx_test

import (
	"context"
	"errors"
	"iter"
	"slices"
	"testing"

	viterx "github.com/MostafaMagdSalama/vortex/iterx"
)

func TestFilter(t *testing.T) {
	var result []int
	for v := range viterx.Filter(context.Background(), slices.Values([]int{1, 2, 3, 4, 5}), func(n int) bool { return n > 2 }) {
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
	for v := range viterx.Filter(ctx, slices.Values([]int{1, 2, 3, 4, 5}), func(n int) bool { return true }) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestMap(t *testing.T) {
	var result []int
	for v := range viterx.Map(context.Background(), slices.Values([]int{1, 2, 3}), func(n int) int { return n * 2 }) {
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
	for v := range viterx.Map(ctx, slices.Values([]int{1, 2, 3}), func(n int) int { return n * 2 }) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestTake(t *testing.T) {
	var result []int
	for v := range viterx.Take(context.Background(), slices.Values([]int{1, 2, 3, 4}), 2) {
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
	for v := range viterx.Take(ctx, slices.Values([]int{1, 2, 3, 4}), 2) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestFlatMap(t *testing.T) {
	var result []int
	for v := range viterx.FlatMap(context.Background(), slices.Values([]int{1, 2, 3}), func(n int) iter.Seq[int] {
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
	for v := range viterx.FlatMap(ctx, slices.Values([]int{1, 2, 3}), func(n int) iter.Seq[int] {
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
	for v := range viterx.TakeWhile(context.Background(), slices.Values([]int{1, 2, 3, 0, 4}), func(n int) bool { return n > 0 }) {
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
	for v := range viterx.TakeWhile(ctx, slices.Values([]int{1, 2, 3}), func(n int) bool { return true }) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestZip(t *testing.T) {
	var result [][2]any
	for pair := range viterx.Zip(context.Background(), slices.Values([]int{1, 2, 3}), slices.Values([]string{"a", "b"})) {
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
	for pair := range viterx.Zip(ctx, slices.Values([]int{1, 2, 3}), slices.Values([]string{"a", "b"})) {
		result = append(result, pair)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestChunk_EvenSplit(t *testing.T) {
	var result [][]int
	for batch := range viterx.Chunk(context.Background(), slices.Values([]int{1, 2, 3, 4, 5, 6}), 2) {
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
	for batch := range viterx.Chunk(ctx, slices.Values([]int{1, 2, 3, 4}), 2) {
		result = append(result, batch)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestFlatten(t *testing.T) {
	var result []int
	for v := range viterx.Flatten(context.Background(), slices.Values([][]int{{1, 2}, {3}, {4, 5}})) {
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
	for v := range viterx.Flatten(ctx, slices.Values([][]int{{1, 2}, {3}})) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestDistinct(t *testing.T) {
	var result []int
	for v := range viterx.Distinct(context.Background(), slices.Values([]int{1, 2, 1, 3, 2, 4})) {
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
	for v := range viterx.Distinct(ctx, slices.Values([]int{1, 2, 1, 3})) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestContains(t *testing.T) {
	if !viterx.Contains(context.Background(), slices.Values([]int{1, 2, 3, 4}), 3) {
		t.Fatal("expected true")
	}
}

func TestContains_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if viterx.Contains(ctx, slices.Values([]int{1, 2, 3, 4}), 3) {
		t.Fatal("expected false on cancelled context")
	}
}

func TestForEach(t *testing.T) {
	var result []int
	viterx.ForEach(context.Background(), slices.Values([]int{1, 2, 3}), func(v int) {
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
	viterx.ForEach(ctx, slices.Values([]int{1, 2, 3}), func(v int) {
		calls++
	})

	if calls != 0 {
		t.Fatalf("expected 0 calls on cancelled context, got %d", calls)
	}
}

func TestReverse(t *testing.T) {
	var result []int
	for v := range viterx.Reverse(context.Background(), slices.Values([]int{1, 2, 3, 4, 5})) {
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
	for v := range viterx.Reverse(ctx, slices.Values([]int{1, 2, 3, 4, 5})) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestDrain_Basic(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3})
	var result []int

	err := viterx.Drain(context.Background(), numbers, func(n int) error {
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

	err := viterx.Drain(context.Background(), numbers, func(n int) error {
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

	err := viterx.Drain(ctx, numbers, func(n int) error {
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
