package iterx_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex"
	"github.com/MostafaMagdSalama/vortex/iterx"
)

// ── FilterSeq ────────────────────────────────────────────────────────────────

func TestFilterSeq_StopsEarly(t *testing.T) {
	callCount := 0
	seq := func(yield func(int) bool) {
		for _, v := range []int{2, 4, 6, 8, 10} {
			callCount++
			if !yield(v) {
				return
			}
		}
	}
	var got []int
	for v := range iterx.FilterSeq(context.Background(), seq, func(n int) bool { return true }) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{2, 4}) {
		t.Fatalf("expected [2 4], got %v", got)
	}
	if callCount > 3 {
		t.Fatalf("expected upstream to stop early, got %d calls", callCount)
	}
}

// ── MapSeq ───────────────────────────────────────────────────────────────────

func TestMapSeq_StopsEarly(t *testing.T) {
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
	for v := range iterx.MapSeq(context.Background(), seq, func(n int) int { return n * 2 }) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{2, 4}) {
		t.Fatalf("expected [2 4], got %v", got)
	}
	if callCount > 3 {
		t.Fatalf("expected upstream to stop early, got %d calls", callCount)
	}
}

// ── ChunkSeq ─────────────────────────────────────────────────────────────────

func TestChunkSeq_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var got [][]int
	for v := range iterx.ChunkSeq(ctx, slices.Values([]int{1, 2, 3}), 2) {
		got = append(got, v)
	}
	if len(got) != 0 {
		t.Fatalf("expected no chunks with cancelled context, got %v", got)
	}
}

func TestChunkSeq_StopsEarly(t *testing.T) {
	callCount := 0
	seq := func(yield func(int) bool) {
		for _, v := range []int{1, 2, 3, 4, 5, 6} {
			callCount++
			if !yield(v) {
				return
			}
		}
	}
	var got [][]int
	for v := range iterx.ChunkSeq(context.Background(), seq, 2) {
		got = append(got, slices.Clone(v))
		if len(got) == 1 {
			break
		}
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(got))
	}
	if callCount > 4 {
		t.Fatalf("expected upstream to stop early, got %d calls", callCount)
	}
}

func TestChunkSeq_ZeroN(t *testing.T) {
	var got [][]int
	for v := range iterx.ChunkSeq(context.Background(), slices.Values([]int{1, 2, 3}), 0) {
		got = append(got, v)
	}
	if len(got) != 0 {
		t.Fatalf("expected no output for n=0, got %v", got)
	}
}

// ── ContainsSeq ──────────────────────────────────────────────────────────────

func TestContainsSeq_Found(t *testing.T) {
	found := iterx.ContainsSeq(context.Background(), slices.Values([]int{1, 2, 3}), 2)
	if !found {
		t.Fatal("expected true, got false")
	}
}

func TestContainsSeq_NotFound(t *testing.T) {
	found := iterx.ContainsSeq(context.Background(), slices.Values([]int{1, 2, 3}), 99)
	if found {
		t.Fatal("expected false, got true")
	}
}

func TestContainsSeq_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	found := iterx.ContainsSeq(ctx, slices.Values([]int{1, 2, 3}), 1)
	if found {
		t.Fatal("expected false with cancelled context, got true")
	}
}

// ── DistinctSeq ──────────────────────────────────────────────────────────────

func TestDistinctSeq_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var got []int
	for v := range iterx.DistinctSeq(ctx, slices.Values([]int{1, 1, 2, 3})) {
		got = append(got, v)
	}
	if len(got) != 0 {
		t.Fatalf("expected no output with cancelled context, got %v", got)
	}
}

func TestDistinctSeq_StopsEarly(t *testing.T) {
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
	for v := range iterx.DistinctSeq(context.Background(), seq) {
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

// ── DrainSeq ─────────────────────────────────────────────────────────────────

func TestDrainSeq_Basic(t *testing.T) {
	var got []int
	err := iterx.DrainSeq(context.Background(), slices.Values([]int{1, 2, 3}), func(n int) error {
		got = append(got, n)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("expected [1 2 3], got %v", got)
	}
}

func TestDrainSeq_CancelledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := iterx.DrainSeq(ctx, slices.Values([]int{1, 2, 3}), func(n int) error { return nil })
	if !errors.Is(err, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled before start, got %v", err)
	}
}

func TestDrainSeq_CancelledMidIteration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	count := 0
	seq := func(yield func(int) bool) {
		for _, v := range []int{1, 2, 3, 4, 5} {
			if !yield(v) {
				return
			}
		}
	}
	err := iterx.DrainSeq(ctx, seq, func(n int) error {
		count++
		if count == 2 {
			cancel()
		}
		return nil
	})
	if !errors.Is(err, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled mid-iteration, got %v", err)
	}
}

// ── ForEachSeq ───────────────────────────────────────────────────────────────

func TestForEachSeq_Basic(t *testing.T) {
	var got []int
	err := iterx.ForEachSeq(context.Background(), slices.Values([]int{1, 2, 3}), func(n int) {
		got = append(got, n)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("expected [1 2 3], got %v", got)
	}
}

func TestForEachSeq_CancelledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := iterx.ForEachSeq(ctx, slices.Values([]int{1, 2, 3}), func(n int) {})
	if !errors.Is(err, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled before start, got %v", err)
	}
}

func TestForEachSeq_CancelledMidIteration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	count := 0
	err := iterx.ForEachSeq(ctx, slices.Values([]int{1, 2, 3, 4, 5}), func(n int) {
		count++
		if count == 2 {
			cancel()
		}
	})
	if !errors.Is(err, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled mid-iteration, got %v", err)
	}
}

// ── FlattenSeq (inner path) ──────────────────────────────────────────────────

func TestFlattenSeq_CancelledDuringInner(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	innerCount := 0
	seq := func(yield func([]int) bool) {
		for _, s := range [][]int{{1, 2, 3}, {4, 5, 6}} {
			if !yield(s) {
				return
			}
		}
	}
	var got []int
	for v := range iterx.FlattenSeq(ctx, seq) {
		innerCount++
		if innerCount == 2 {
			cancel() // cancel after second item from first inner slice
		}
		got = append(got, v)
	}
	// After cancellation, no more items should be produced
	if len(got) > 3 {
		t.Fatalf("expected at most 3 items before cancellation takes effect, got %v", got)
	}
}

func TestFlattenSeq_StopsEarly(t *testing.T) {
	callCount := 0
	seq := func(yield func([]int) bool) {
		for _, s := range [][]int{{1, 2}, {3, 4}, {5, 6}} {
			callCount++
			if !yield(s) {
				return
			}
		}
	}
	var got []int
	for v := range iterx.FlattenSeq(context.Background(), seq) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("expected [1 2], got %v", got)
	}
	if callCount > 2 {
		t.Fatalf("expected upstream to stop after first slice, got %d calls", callCount)
	}
}

func TestDrainSeq_FnError(t *testing.T) {
	sentinelErr := errors.New("fn failed")
	err := iterx.DrainSeq(context.Background(), slices.Values([]int{1, 2, 3}), func(n int) error {
		if n == 2 {
			return sentinelErr
		}
		return nil
	})
	if !errors.Is(err, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestChunkSeq_CancelledBeforeTrailingBatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	seq := func(yield func(int) bool) {
		if !yield(1) {
			return
		}
		if !yield(2) {
			return
		}
		if !yield(3) {
			return
		}
		cancel() // cancel after last item, before trailing batch is yielded
	}
	var got [][]int
	for v := range iterx.ChunkSeq(ctx, seq, 2) {
		got = append(got, slices.Clone(v))
	}
	if len(got) != 1 || !slices.Equal(got[0], []int{1, 2}) {
		t.Fatalf("expected [[1 2]] only, got %v", got)
	}
}

// ── ZipSeq (second seq shorter path) ─────────────────────────────────────────

func TestZipSeq_StopsEarly(t *testing.T) {
	callCountA := 0
	seqA := func(yield func(int) bool) {
		for _, v := range []int{1, 2, 3, 4, 5} {
			callCountA++
			if !yield(v) {
				return
			}
		}
	}
	var got []iterx.Pair[int, int]
	for p := range iterx.ZipSeq(context.Background(), seqA, slices.Values([]int{10, 20, 30, 40, 50})) {
		got = append(got, p)
		if len(got) == 2 {
			break
		}
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(got))
	}
	if callCountA > 3 {
		t.Fatalf("expected seqA to stop early, got %d calls", callCountA)
	}
}
