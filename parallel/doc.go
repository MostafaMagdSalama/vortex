// Package parallel provides concurrent, context-aware sequence transforms for Go iter.Seq and iter.Seq2 values.
//
// The parallel package is useful when each item in a sequence needs independent CPU or I/O work and you want to spread that work across a fixed number of workers. It keeps the same lazy iteration style as the rest of the library while letting callers control concurrency with standard context cancellation.
//
// Use ParallelMapSeq, BatchMapSeq, and OrderedParallelMapSeq with plain iter.Seq[T]
// values such as slices.Values(...). Use ParallelMap, BatchMap, and
// OrderedParallelMap with iter.Seq2[T, error] values so source errors can flow
// through the pipeline instead of being dropped.
//
// Example:
//
//	ctx := context.Background()
//	numbers := slices.Values([]int{1, 2, 3, 4})
//
//	for v := range parallel.ParallelMapSeq(ctx, numbers, func(n int) int {
//		return n * 2
//	}, 2) {
//		fmt.Println(v)
//	}
//
// ParallelMapSeq and ParallelMap may yield results out of order when multiple
// workers are used, and they stop quickly when the context is cancelled or the
// consumer stops early. BatchMapSeq and BatchMap preserve batch order and do
// not spin up worker goroutines by themselves.
package parallel
