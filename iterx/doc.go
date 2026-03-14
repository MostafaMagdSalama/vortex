// Package interx provides lazy, context-aware sequence transformations for Go 1.23 iter.Seq values.
//
// The interx package is useful when you want to build allocation-light data pipelines over slices, files, database rows, or generated streams without materializing every intermediate result. Each helper returns a new lazy sequence or consumes one directly, so work only happens when the caller ranges over the final sequence.
//
// Example:
//
//	ctx := context.Background()
//	numbers := slices.Values([]int{1, 2, 3, 4, 5})
//
//	for v := range interx.Filter(ctx, numbers, func(n int) bool {
//		return n%2 == 0
//	}) {
//		fmt.Println(v)
//	}
//
// Any important notes about behaviour such as laziness, context cancellation, memory usage. Most functions are lazy and stop as soon as the context is cancelled or the consumer breaks early. Reverse buffers the full sequence in memory before yielding results, while helpers like Filter, Map, and Take stream values one at a time.
package iterx
