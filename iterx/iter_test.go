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
	for v := range viterx.FilterSeq(context.Background(), slices.Values([]int{1, 2, 3, 4, 5}), func(n int) bool { return n > 2 }) {
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
	for v := range viterx.FilterSeq(ctx, slices.Values([]int{1, 2, 3, 4, 5}), func(n int) bool { return true }) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestMap(t *testing.T) {
	var result []int
	for v := range viterx.MapSeq(context.Background(), slices.Values([]int{1, 2, 3}), func(n int) int { return n * 2 }) {
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
	for v := range viterx.MapSeq(ctx, slices.Values([]int{1, 2, 3}), func(n int) int { return n * 2 }) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestTake(t *testing.T) {
	var result []int
	for v := range viterx.TakeSeq(context.Background(), slices.Values([]int{1, 2, 3, 4}), 2) {
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
	for v := range viterx.TakeSeq(ctx, slices.Values([]int{1, 2, 3, 4}), 2) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestFlatMap(t *testing.T) {
	var result []int
	for v := range viterx.FlatMapSeq(context.Background(), slices.Values([]int{1, 2, 3}), func(n int) iter.Seq[int] {
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
	for v := range viterx.FlatMapSeq(ctx, slices.Values([]int{1, 2, 3}), func(n int) iter.Seq[int] {
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
	for v := range viterx.TakeWhileSeq(context.Background(), slices.Values([]int{1, 2, 3, 0, 4}), func(n int) bool { return n > 0 }) {
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
	for v := range viterx.TakeWhileSeq(ctx, slices.Values([]int{1, 2, 3}), func(n int) bool { return true }) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestZip(t *testing.T) {
	var result [][2]any
	for pair := range viterx.ZipSeq(context.Background(), slices.Values([]int{1, 2, 3}), slices.Values([]string{"a", "b"})) {
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
	for pair := range viterx.ZipSeq(ctx, slices.Values([]int{1, 2, 3}), slices.Values([]string{"a", "b"})) {
		result = append(result, pair)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestChunk_EvenSplit(t *testing.T) {
	var result [][]int
	for batch := range viterx.ChunkSeq(context.Background(), slices.Values([]int{1, 2, 3, 4, 5, 6}), 2) {
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
	for batch := range viterx.ChunkSeq(ctx, slices.Values([]int{1, 2, 3, 4}), 2) {
		result = append(result, batch)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestFlatten(t *testing.T) {
	var result []int
	for v := range viterx.FlattenSeq(context.Background(), slices.Values([][]int{{1, 2}, {3}, {4, 5}})) {
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
	for v := range viterx.FlattenSeq(ctx, slices.Values([][]int{{1, 2}, {3}})) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestDistinct(t *testing.T) {
	var result []int
	for v := range viterx.DistinctSeq(context.Background(), slices.Values([]int{1, 2, 1, 3, 2, 4})) {
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
	for v := range viterx.DistinctSeq(ctx, slices.Values([]int{1, 2, 1, 3})) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestContains(t *testing.T) {
	if !viterx.ContainsSeq(context.Background(), slices.Values([]int{1, 2, 3, 4}), 3) {
		t.Fatal("expected true")
	}
}

func TestContains_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if viterx.ContainsSeq(ctx, slices.Values([]int{1, 2, 3, 4}), 3) {
		t.Fatal("expected false on cancelled context")
	}
}

func TestForEach(t *testing.T) {
	var result []int
	viterx.ForEachSeq(context.Background(), slices.Values([]int{1, 2, 3}), func(v int) {
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
	viterx.ForEachSeq(ctx, slices.Values([]int{1, 2, 3}), func(v int) {
		calls++
	})

	if calls != 0 {
		t.Fatalf("expected 0 calls on cancelled context, got %d", calls)
	}
}

func TestReverse(t *testing.T) {
	var result []int
	for v := range viterx.ReverseSeq(context.Background(), slices.Values([]int{1, 2, 3, 4, 5})) {
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
	for v := range viterx.ReverseSeq(ctx, slices.Values([]int{1, 2, 3, 4, 5})) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(result))
	}
}

func TestDrain_Basic(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3})
	var result []int

	err := viterx.DrainSeq(context.Background(), numbers, func(n int) error {
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

	err := viterx.DrainSeq(context.Background(), numbers, func(n int) error {
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

	err := viterx.DrainSeq(ctx, numbers, func(n int) error {
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

func TestTake_ErrorAfterN(t *testing.T) {
    seq := func(yield func(int, error) bool) {
        if !yield(1, nil) {
            return
        }
        if !yield(2, nil) {
            return
        }
        if !yield(0, errors.New("upstream error")) {
            return
        }
    }

    var values []int
    var errs []error
    for v, err := range viterx.Take(context.Background(), seq, 2) {
        if err != nil {
            errs = append(errs, err)
            continue
        }
        values = append(values, v)
    }

    if !slices.Equal(values, []int{1, 2}) {
        t.Fatalf("expected [1 2], got %v", values)
    }
    if len(errs) != 0 {
        t.Fatalf("expected no errors, got %v", errs)
    }
}
func TestFilter_Seq2_ErrorPassThrough(t *testing.T) {
    seq := func(yield func(int, error) bool) {
        yield(1, nil)
        yield(0, errors.New("upstream error"))
        yield(3, nil)
    }

    var values []int
    var errs []error
    for v, err := range viterx.Filter(context.Background(), seq, func(n int) bool { return n > 0 }) {
        if err != nil {
            errs = append(errs, err)
            continue
        }
        values = append(values, v)
    }

    if !slices.Equal(values, []int{1, 3}) {
        t.Fatalf("expected [1 3], got %v", values)
    }
    if len(errs) != 1 {
        t.Fatalf("expected 1 error, got %d", len(errs))
    }
}
func TestDrain_Seq2_StopsOnUpstreamError(t *testing.T) {
    seq := func(yield func(int, error) bool) {
        if !yield(1, nil) {
            return
        }
        if !yield(0, errors.New("upstream error")) {
            return
        }
        if !yield(3, nil) {
            return
        }
    }

    count := 0
    err := viterx.Drain(context.Background(), seq, func(n int) error {
        count++
        return nil
    })

    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if count != 1 {
        t.Fatalf("expected 1 call before error, got %d", count)
    }
}