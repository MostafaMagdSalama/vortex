// Package sources provides lazy, context-aware iter.Seq values for common input sources such as readers, files, stdin, and SQL queries.
//
// The sources package is useful when you want to turn external data into sequences that can be consumed directly or composed with the seq and parallel packages. It lets you process CSV rows, lines, and database results incrementally instead of loading entire inputs into memory up front.
//
// Example:
//
//	ctx := context.Background()
//	input := strings.NewReader("name,age\nAlice,30\nBob,25\n")
//
//	for row := range sources.CSVRows(ctx, input) {
//		fmt.Println(row[0], row[1])
//	}
//
// Any important notes about behaviour such as laziness, context cancellation, memory usage. These sequences read incrementally and stop immediately when the context is cancelled. File and database helpers acquire resources when iteration starts, and silent variants drop read or scan errors while the WithError variants surface them through iter.Seq2.
package sources
