package iter_test

import (
    "slices"
    "testing"

    "vortex/iter"
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