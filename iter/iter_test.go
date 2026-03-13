package iter_test

import (
    "slices"
    "testing"

    "github.com/MostafaMagdSalama/vortex/iter"
)

func TestFilter(t *testing.T) {
    source := slices.Values([]int{1, 2, 3, 4, 5})

    var result []int
    for v := range iter.Filter(source, func(n int) bool { return n > 2 }) {
        result = append(result, v)
    }

    // result should be [3, 4, 5]
    if len(result) != 3 || result[0] != 3 {
        t.Fatalf("got %v", result)
    }
}
// ─── Chunk ───────────────────────────────

func TestChunk_EvenSplit(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3, 4, 5, 6})

	var result [][]int
	for batch := range iter.Chunk(numbers, 2) {
		result = append(result, batch)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(result))
	}
	if len(result[0]) != 2 || result[0][0] != 1 || result[0][1] != 2 {
		t.Fatalf("unexpected first batch: %v", result[0])
	}
}

func TestChunk_UnevenSplit(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	var result [][]int
	for batch := range iter.Chunk(numbers, 2) {
		result = append(result, batch)
	}

	// [1 2], [3 4], [5]
	if len(result) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(result))
	}
	if len(result[2]) != 1 || result[2][0] != 5 {
		t.Fatalf("expected last batch [5], got %v", result[2])
	}
}

func TestChunk_EarlyStop(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3, 4, 5, 6})

	count := 0
	for range iter.Take(iter.Chunk(numbers, 2), 1) {
		count++
	}

	if count != 1 {
		t.Fatalf("expected 1 batch, got %d", count)
	}
}

// ─── Flatten ─────────────────────────────

func TestFlatten_Basic(t *testing.T) {
	groups := slices.Values([][]int{{1, 2}, {3, 4}, {5}})

	var result []int
	for v := range iter.Flatten(groups) {
		result = append(result, v)
	}

	expected := []int{1, 2, 3, 4, 5}
	if len(result) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
	for i, v := range result {
		if v != expected[i] {
			t.Fatalf("index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestFlatten_EarlyStop(t *testing.T) {
	groups := slices.Values([][]int{{1, 2}, {3, 4}, {5, 6}})

	var result []int
	for v := range iter.Take(iter.Flatten(groups), 3) {
		result = append(result, v)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}
}

func TestFlatten_EmptySlices(t *testing.T) {
	groups := slices.Values([][]int{{}, {1, 2}, {}, {3}})

	var result []int
	for v := range iter.Flatten(groups) {
		result = append(result, v)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}
}

// ─── Distinct ────────────────────────────

func TestDistinct_Basic(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 1, 3, 2, 4})

	var result []int
	for v := range iter.Distinct(numbers) {
		result = append(result, v)
	}

	if len(result) != 4 {
		t.Fatalf("expected 4 unique values, got %d", len(result))
	}
}

func TestDistinct_AllDuplicates(t *testing.T) {
	numbers := slices.Values([]int{1, 1, 1, 1})

	var result []int
	for v := range iter.Distinct(numbers) {
		result = append(result, v)
	}

	if len(result) != 1 || result[0] != 1 {
		t.Fatalf("expected [1], got %v", result)
	}
}

func TestDistinct_NoDuplicates(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3, 4})

	var result []int
	for v := range iter.Distinct(numbers) {
		result = append(result, v)
	}

	if len(result) != 4 {
		t.Fatalf("expected 4 items, got %d", len(result))
	}
}

func TestDistinct_Strings(t *testing.T) {
	words := slices.Values([]string{"a", "b", "a", "c", "b"})

	var result []string
	for v := range iter.Distinct(words) {
		result = append(result, v)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 unique strings, got %d", len(result))
	}
}

// ─── Contains ────────────────────────────

func TestContains_Found(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	if !iter.Contains(numbers, 3) {
		t.Fatal("expected true, got false")
	}
}

func TestContains_NotFound(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	if iter.Contains(numbers, 9) {
		t.Fatal("expected false, got true")
	}
}

func TestContains_EmptySequence(t *testing.T) {
	numbers := slices.Values([]int{})

	if iter.Contains(numbers, 1) {
		t.Fatal("expected false for empty sequence")
	}
}

func TestContains_StopsEarly(t *testing.T) {
	// if Contains stops early, this counter should be low
	count := 0
	seq := func(yield func(int) bool) {
		for i := 1; i <= 100; i++ {
			count++
			if !yield(i) {
				return
			}
		}
	}

	iter.Contains(seq, 3)

	// should have stopped after finding 3 — not read all 100
	if count > 5 {
		t.Fatalf("expected early stop, read %d items", count)
	}
}

// ─── ForEach ─────────────────────────────

func TestForEach_Basic(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	var result []int
	iter.ForEach(numbers, func(v int) {
		result = append(result, v*2)
	})

	if len(result) != 5 {
		t.Fatalf("expected 5 items, got %d", len(result))
	}
	if result[0] != 2 || result[4] != 10 {
		t.Fatalf("unexpected result: %v", result)
	}
}

func TestForEach_Empty(t *testing.T) {
	numbers := slices.Values([]int{})

	count := 0
	iter.ForEach(numbers, func(v int) {
		count++
	})

	if count != 0 {
		t.Fatalf("expected 0 calls, got %d", count)
	}
}

func TestForEach_SideEffect(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3})

	sum := 0
	iter.ForEach(numbers, func(v int) {
		sum += v
	})

	if sum != 6 {
		t.Fatalf("expected sum 6, got %d", sum)
	}
}

// ─── Reverse ─────────────────────────────

func TestReverse_Basic(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	var result []int
	for v := range iter.Reverse(numbers) {
		result = append(result, v)
	}

	expected := []int{5, 4, 3, 2, 1}
	for i, v := range result {
		if v != expected[i] {
			t.Fatalf("index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestReverse_SingleItem(t *testing.T) {
	numbers := slices.Values([]int{42})

	var result []int
	for v := range iter.Reverse(numbers) {
		result = append(result, v)
	}

	if len(result) != 1 || result[0] != 42 {
		t.Fatalf("expected [42], got %v", result)
	}
}

func TestReverse_Empty(t *testing.T) {
	numbers := slices.Values([]int{})

	var result []int
	for v := range iter.Reverse(numbers) {
		result = append(result, v)
	}

	if len(result) != 0 {
		t.Fatalf("expected empty, got %v", result)
	}
}

func TestReverse_EarlyStop(t *testing.T) {
	numbers := slices.Values([]int{1, 2, 3, 4, 5})

	var result []int
	for v := range iter.Take(iter.Reverse(numbers), 3) {
		result = append(result, v)
	}

	// should get 5, 4, 3
	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}
	if result[0] != 5 || result[1] != 4 || result[2] != 3 {
		t.Fatalf("expected [5 4 3], got %v", result)
	}
}