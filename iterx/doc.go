// Package iterx provides lazy, context-aware sequence transformations for Go iter.Seq and iter.Seq2 values.
//
// The iterx package is useful when you want to build allocation-light data pipelines over slices, files, database rows, or generated streams without materializing every intermediate result. Each helper returns a new lazy sequence or consumes one directly, so work only happens when the caller ranges over the final sequence.
//
// # API split
//
// iterx provides two variants of every function:
//
//   - Plain variants (FilterSeq, MapSeq, TakeSeq, ...) accept iter.Seq[T] and are
//     suitable for slices, custom generators, or any source that does not yield errors.
//
//   - Error-aware variants (Filter, Map, Take, ...) accept iter.Seq2[T, error] and are
//     suitable for pipelines that start from vortex/sources. Errors pass through the
//     pipeline untouched — fn is only called on valid values.
//
// # Plain iter.Seq example
//
//	ctx := context.Background()
//	numbers := slices.Values([]int{1, 2, 3, 4, 5})
//
//	for v := range iterx.FilterSeq(ctx, numbers, func(n int) bool {
//		return n%2 == 0
//	}) {
//		fmt.Println(v) // 2, 4
//	}
//
// # iter.Seq2 pipeline example (with vortex/sources)
//
//	ctx := context.Background()
//	file, _ := os.Open("users.csv")
//	defer file.Close()
//
//	rows     := sources.CSVRows(ctx, file)
//	filtered := iterx.Filter(ctx, rows, func(row []string) bool { return row[2] == "active" })
//	taken    := iterx.Take(ctx, filtered, 10)
//
//	for row, err := range taken {
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Println(row)
//	}
//
// # Memory notes
//
// All functions stream values one at a time except Reverse, which buffers the
// full sequence in memory before yielding results. Reverse is not suitable for
// infinite or very large sequences.
//
// # Context cancellation
//
// All functions check ctx.Err() at every iteration. Pipelines cancel cleanly
// the moment the context is cancelled — no goroutines are leaked and no extra
// items are processed.
package iterx