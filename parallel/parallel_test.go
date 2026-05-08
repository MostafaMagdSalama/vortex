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

	"github.com/MostafaMagdSalama/vortex"
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

func TestParallelMapSeq_ZeroWorkersPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for workers=0, got none")
		}
	}()
	parallel.ParallelMapSeq(context.Background(), slices.Values([]int{}), func(n int) int { return n }, 0)
}

func TestParallelMap_ZeroWorkersPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for workers=0, got none")
		}
	}()
	parallel.ParallelMap(context.Background(), seq2FromSlice([]int{}), func(n int) int { return n }, 0)
}

func TestOrderedParallelMapSeq_ZeroWorkersPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for workers=0, got none")
		}
	}()
	parallel.OrderedParallelMapSeq(context.Background(), slices.Values([]int{}), func(n int) int { return n }, 0)
}

func TestOrderedParallelMap_ZeroWorkersPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for workers=0, got none")
		}
	}()
	parallel.OrderedParallelMap(context.Background(), seq2FromSlice([]int{}), func(n int) int { return n }, 0)
}

func TestBatchMapSeq_ZeroBatchSizePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for batchSize=0, got none")
		}
	}()
	parallel.BatchMapSeq(context.Background(), slices.Values([]int{}), func(b []int) []int { return b }, 0)
}

func TestBatchMap_ZeroBatchSizePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for batchSize=0, got none")
		}
	}()
	parallel.BatchMap(context.Background(), seq2FromSlice([]int{}), func(b []int) []int { return b }, 0)
}

func TestParallelMap_WorkerPanicPropagatesAsError(t *testing.T) {
	seq := seq2FromSlice([]int{1, 2, 3, 4, 5})

	var gotErr error
	for _, err := range parallel.ParallelMap(context.Background(), seq, func(n int) int {
		if n == 3 {
			panic("intentional panic in worker")
		}
		return n * 10
	}, 2) {
		if err != nil {
			gotErr = err
			break
		}
	}

	if gotErr == nil {
		t.Fatal("expected error from worker panic, got nil")
	}
	if !strings.Contains(gotErr.Error(), "worker panic") {
		t.Fatalf("expected 'worker panic' in error, got: %v", gotErr)
	}
}

func TestOrderedParallelMap_WorkerPanicPropagatesAsError(t *testing.T) {
	seq := seq2FromSlice([]int{1, 2, 3, 4, 5})

	var gotErr error
	for _, err := range parallel.OrderedParallelMap(context.Background(), seq, func(n int) int {
		if n == 3 {
			panic("intentional panic in ordered worker")
		}
		return n * 10
	}, 2) {
		if err != nil {
			gotErr = err
			break
		}
	}

	if gotErr == nil {
		t.Fatal("expected error from worker panic, got nil")
	}
	if !strings.Contains(gotErr.Error(), "worker panic") {
		t.Fatalf("expected 'worker panic' in error, got: %v", gotErr)
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

func TestBatchMap_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var gotErr error
	for _, err := range parallel.BatchMap(ctx, seq2FromSlice([]int{1, 2, 3, 4}), func(b []int) []int { return b }, 2) {
		if err != nil {
			gotErr = err
			break
		}
	}

	if gotErr == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}

func TestBatchMap_StopsEarlyOnValue(t *testing.T) {
	callCount := 0
	input := seq2FromSlice([]int{1, 2, 3, 4, 5, 6})

	var got []int
	for v, err := range parallel.BatchMap(context.Background(), input, func(b []int) []int {
		callCount++
		return b
	}, 2) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if callCount > 2 {
		t.Fatalf("expected fn to stop being called after consumer break, got %d calls", callCount)
	}
}

func TestBatchMap_StopsEarlyOnError(t *testing.T) {
	sentinelErr := errors.New("upstream error")
	input := seq2WithError([]int{1, 2, 3, 4}, 1, sentinelErr)

	var gotErr error
	var got []int
	for v, err := range parallel.BatchMap(context.Background(), input, func(b []int) []int { return b }, 2) {
		if err != nil {
			gotErr = err
			break // consumer stops on first error
		}
		got = append(got, v)
	}

	if !errors.Is(gotErr, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", gotErr)
	}
	// batch [1] was flushed before the error, so got=[1]
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("expected [1] before error, got %v", got)
	}
}

func TestBatchMapSeq_StopsEarly(t *testing.T) {
	callCount := 0
	input := slices.Values([]int{1, 2, 3, 4, 5, 6})

	var got []int
	for v := range parallel.BatchMapSeq(context.Background(), input, func(b []int) []int {
		callCount++
		return b
	}, 2) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if callCount > 2 {
		t.Fatalf("expected fn to stop being called after consumer break, got %d calls", callCount)
	}
}

// Regression: when fn returns the same slice it was handed (T==U), the
// internal buffer must NOT alias that slice — otherwise the next batch's
// appends overwrite values still being yielded.
func TestBatchMapSeq_NoAliasingWhenFnReturnsInput(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5, 6})

	var got []int
	for v := range parallel.BatchMapSeq(context.Background(), input, func(b []int) []int {
		return b // identity — would alias the buffer pre-fix
	}, 2) {
		got = append(got, v)
	}

	want := []int{1, 2, 3, 4, 5, 6}
	if !slices.Equal(got, want) {
		t.Fatalf("aliasing detected: got %v, want %v", got, want)
	}
}

func TestBatchMap_NoAliasingWhenFnReturnsInput(t *testing.T) {
	input := seq2FromSlice([]int{1, 2, 3, 4, 5, 6})

	var got []int
	for v, err := range parallel.BatchMap(context.Background(), input, func(b []int) []int {
		return b
	}, 2) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}

	want := []int{1, 2, 3, 4, 5, 6}
	if !slices.Equal(got, want) {
		t.Fatalf("aliasing detected: got %v, want %v", got, want)
	}
}

// panicSeq yields a few values and then panics. Used to verify producer
// goroutines recover from source panics rather than crashing the program.
func panicSeq(items []int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i, v := range items {
			if i == 2 {
				panic("source explosion")
			}
			if !yield(v) {
				return
			}
		}
	}
}

func panicSeq2(items []int) iter.Seq2[int, error] {
	return func(yield func(int, error) bool) {
		for i, v := range items {
			if i == 2 {
				panic("source explosion")
			}
			if !yield(v, nil) {
				return
			}
		}
	}
}

func TestParallelMapSeq_RecoversSourcePanic(t *testing.T) {
	// Without the producer recover this test would crash the test binary.
	for v := range parallel.ParallelMapSeq(context.Background(), panicSeq([]int{1, 2, 3, 4, 5}), func(n int) int { return n }, 2) {
		_ = v
	}
}

func TestParallelMap_RecoversSourcePanicAsError(t *testing.T) {
	var sawErr bool
	for _, err := range parallel.ParallelMap(context.Background(), panicSeq2([]int{1, 2, 3, 4, 5}), func(n int) int { return n }, 2) {
		if err != nil {
			sawErr = true
			if !strings.Contains(err.Error(), "source panic") {
				t.Fatalf("expected 'source panic' in error, got %v", err)
			}
		}
	}
	if !sawErr {
		t.Fatal("expected an error from a panicking source, got none")
	}
}

func TestOrderedParallelMapSeq_RecoversSourcePanic(t *testing.T) {
	for v := range parallel.OrderedParallelMapSeq(context.Background(), panicSeq([]int{1, 2, 3, 4, 5}), func(n int) int { return n }, 2) {
		_ = v
	}
}

func TestOrderedParallelMap_RecoversSourcePanicAsError(t *testing.T) {
	var sawErr bool
	for _, err := range parallel.OrderedParallelMap(context.Background(), panicSeq2([]int{1, 2, 3, 4, 5}), func(n int) int { return n }, 2) {
		if err != nil {
			sawErr = true
			if !strings.Contains(err.Error(), "source panic") {
				t.Fatalf("expected 'source panic' in error, got %v", err)
			}
		}
	}
	if !sawErr {
		t.Fatal("expected an error from a panicking source, got none")
	}
}

// ---- Phase 2: ParallelMapErr / OrderedParallelMapErr ----

func TestParallelMapErr(t *testing.T) {
	input := seq2FromSlice([]int{1, 2, 3, 4, 5})
	var got []int
	for v, err := range parallel.ParallelMapErr(
		context.Background(), input,
		func(n int) (int, error) { return n * 2, nil },
		3,
	) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got = append(got, v)
	}
	sort.Ints(got)
	if !slices.Equal(got, []int{2, 4, 6, 8, 10}) {
		t.Fatalf("got %v", got)
	}
}

func TestOrderedParallelMapErr(t *testing.T) {
	input := seq2FromSlice([]int{1, 2, 3, 4, 5})
	var got []int
	for v, err := range parallel.OrderedParallelMapErr(
		context.Background(), input,
		func(n int) (int, error) { return n * 2, nil },
		3,
	) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got = append(got, v)
	}
	if !slices.Equal(got, []int{2, 4, 6, 8, 10}) {
		t.Fatalf("got %v (order should be preserved)", got)
	}
}

func TestParallelMapErr_FnErrorsWrapped(t *testing.T) {
	input := seq2FromSlice([]int{1, 2, 3, 4, 5})
	boom := errors.New("boom")
	var got []int
	var errs int
	for v, err := range parallel.ParallelMapErr(
		context.Background(), input,
		func(n int) (int, error) {
			if n == 3 {
				return 0, boom
			}
			return n * 2, nil
		},
		2,
	) {
		if err != nil {
			errs++
			if !errors.Is(err, boom) {
				t.Fatalf("expected wrapped boom, got %v", err)
			}
			continue
		}
		got = append(got, v)
	}
	sort.Ints(got)
	if !slices.Equal(got, []int{2, 4, 8, 10}) {
		t.Fatalf("got values %v", got)
	}
	if errs != 1 {
		t.Fatalf("expected 1 error, got %d", errs)
	}
}

func TestOrderedParallelMapErr_OrderPreservedAcrossErrors(t *testing.T) {
	input := seq2FromSlice([]int{1, 2, 3, 4, 5})
	boom := errors.New("boom")
	type out struct {
		v   int
		err bool
	}
	var got []out
	for v, err := range parallel.OrderedParallelMapErr(
		context.Background(), input,
		func(n int) (int, error) {
			if n == 3 {
				return 0, boom
			}
			return n * 10, nil
		},
		4,
	) {
		if err != nil {
			got = append(got, out{err: true})
			continue
		}
		got = append(got, out{v: v})
	}
	want := []out{{v: 10}, {v: 20}, {err: true}, {v: 40}, {v: 50}}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("at %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParallelMapErr_SourceErrorsPassThrough(t *testing.T) {
	boom := errors.New("source-boom")
	input := seq2WithError([]int{1, 2, 3}, 1, boom)
	var sawErr bool
	for _, err := range parallel.ParallelMapErr(
		context.Background(), input,
		func(n int) (int, error) { return n, nil },
		2,
	) {
		if err != nil && errors.Is(err, boom) {
			sawErr = true
		}
	}
	if !sawErr {
		t.Fatal("expected source error to surface")
	}
}

func TestParallelMapErr_RecoversWorkerPanic(t *testing.T) {
	input := seq2FromSlice([]int{1, 2, 3, 4, 5})
	var sawPanic bool
	for _, err := range parallel.ParallelMapErr(
		context.Background(), input,
		func(n int) (int, error) {
			if n == 3 {
				panic("worker explosion")
			}
			return n, nil
		},
		2,
	) {
		if err != nil && strings.Contains(err.Error(), "worker panic") {
			sawPanic = true
		}
	}
	if !sawPanic {
		t.Fatal("expected worker panic to surface as error")
	}
}

func TestParallelMapErr_PreCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var gotErr error
	for _, err := range parallel.ParallelMapErr(
		ctx, seq2FromSlice([]int{1, 2, 3}),
		func(n int) (int, error) { return n, nil },
		2,
	) {
		if err != nil {
			gotErr = err
			break
		}
	}
	if !errors.Is(gotErr, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", gotErr)
	}
}

func TestParallelMapSeqErr(t *testing.T) {
	boom := errors.New("boom")
	var got []int
	var errs int
	for v, err := range parallel.ParallelMapSeqErr(
		context.Background(), slices.Values([]int{1, 2, 3, 4, 5}),
		func(n int) (int, error) {
			if n == 3 {
				return 0, boom
			}
			return n * 2, nil
		},
		2,
	) {
		if err != nil {
			errs++
			continue
		}
		got = append(got, v)
	}
	sort.Ints(got)
	if !slices.Equal(got, []int{2, 4, 8, 10}) || errs != 1 {
		t.Fatalf("got values %v errs %d", got, errs)
	}
}

func TestOrderedParallelMapSeqErr(t *testing.T) {
	var got []int
	for v, err := range parallel.OrderedParallelMapSeqErr(
		context.Background(), slices.Values([]int{1, 2, 3}),
		func(n int) (int, error) { return n * 2, nil },
		2,
	) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	if !slices.Equal(got, []int{2, 4, 6}) {
		t.Fatalf("got %v", got)
	}
}

func TestParallelMapErr_ZeroWorkersPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for workers <= 0")
		}
	}()
	parallel.ParallelMapErr(context.Background(), seq2FromSlice([]int{1}),
		func(n int) (int, error) { return n, nil }, 0)
}

func TestOrderedParallelMapErr_ZeroWorkersPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for workers <= 0")
		}
	}()
	parallel.OrderedParallelMapErr(context.Background(), seq2FromSlice([]int{1}),
		func(n int) (int, error) { return n, nil }, 0)
}

func TestParallelMapSeqErr_ZeroWorkersPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for workers <= 0")
		}
	}()
	parallel.ParallelMapSeqErr(context.Background(), slices.Values([]int{1}),
		func(n int) (int, error) { return n, nil }, 0)
}

func TestOrderedParallelMapSeqErr_ZeroWorkersPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for workers <= 0")
		}
	}()
	parallel.OrderedParallelMapSeqErr(context.Background(), slices.Values([]int{1}),
		func(n int) (int, error) { return n, nil }, 0)
}
