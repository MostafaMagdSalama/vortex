// Package sources provides lazy data sources that produce iter.Seq2[T, error] sequences.
//
// Sources return iter.Seq2[T, error] so read, decode, scan, and cancellation
// failures can flow through the pipeline. When composing these sources with
// vortex/iterx, use the helpers that accept iter.Seq2 such as iterx.Filter,
// iterx.Map, iterx.Take, iterx.Validate, iterx.Drain, and iterx.ForEach.
//
// For plain iter.Seq[T] values that cannot yield errors, vortex keeps parallel
// helper variants like iterx.FilterSeq, iterx.MapSeq, and iterx.TakeSeq.
//
// Sources stream data one item at a time — they never load everything into memory.
// Supports CSV files, JSONL files, databases, text files, and standard input.
//
// All sources accept any io.Reader — files, HTTP responses, and network streams
// all work without any changes to your code.
//
// # Benchmarks
//
// Measured on Windows with 1,000,000 line JSONL file:
//
//	                 vortex      eager
//	early stop       24ms        909ms     (37x faster)
//	peak memory      1 MB        194 MB    (194x less)
//
//	full scan        703ms       808ms     (similar)
//	peak memory      2 MB        168 MB    (84x less)
//
// # Example — JSONL file
//
//	file, err := os.Open("logs.jsonl")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	// find first 100 errors — reads only ~991 lines from a 1M line file
//	logs   := sources.JSONLines[LogEntry](ctx, file)
//	errors := iterx.Filter(ctx, logs, func(e LogEntry) bool { return e.Level == "error" })
//	first100 := iterx.Take(ctx, errors, 100)
//
//	for entry, err := range first100 {
//	    if err != nil { log.Fatal(err) }
//	    fmt.Println(entry.Message)
//	}
//
// # Example — database
//
//	for row, err := range sources.DBRows(ctx, db, "SELECT * FROM users", scan) {
//	    if err != nil { log.Fatal(err) }
//	    process(row)
//	}
//
// # Example — CSV from HTTP
//
//	resp, _ := http.DefaultClient.Do(req)
//	defer resp.Body.Close()
//
//	for row, err := range sources.CSVRows(ctx, resp.Body) {
//	    if err != nil { log.Fatal(err) }
//	    fmt.Println(row)
//	}
package sources
