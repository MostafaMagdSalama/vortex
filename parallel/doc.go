// Package parallel provides concurrent, context-aware sequence transforms for Go 1.23 iter.Seq values.
//
// The parallel package is useful when each item in a sequence needs independent CPU or I/O work and you want to spread that work across a fixed number of workers. It keeps the same lazy iteration style as the rest of the library while letting callers control concurrency with standard context cancellation.
//
// Example:
//
//	ctx := context.Background()
//	numbers := slices.Values([]int{1, 2, 3, 4})
//
//	for v := range parallel.ParallelMap(ctx, numbers, func(n int) int {
//		return n * 2
//	}, 2) {
//		fmt.Println(v)
//	}
//
// Any important notes about behaviour such as laziness, context cancellation, memory usage. ParallelMap and WorkerPoolMap may yield results out of order when multiple workers are used, and they stop quickly when the context is cancelled or the consumer stops early. BatchMap preserves batch order and does not spin up worker goroutines by itself.
package parallel
