# Changelog

## v1.0.0 — 2026-03-17

First stable release of vortex — a zero-dependency Go 1.23 library for lazy, memory-efficient data pipelines.

### iterx

| Function | Description |
|---|---|
| `Filter` | Lazy filter for `iter.Seq2[T, error]` — errors pass through untouched |
| `Map` | Lazy map for `iter.Seq2[T, error]` — errors pass through untouched |
| `Take` | Lazy take for `iter.Seq2[T, error]` — stops upstream immediately after n valid items |
| `FlatMap` | Lazy flat map for `iter.Seq2[T, error]` — errors pass through untouched |
| `TakeWhile` | Yields items while predicate is true — errors pass through untouched |
| `Zip` | Combines two `iter.Seq2` sequences into pairs — errors from both pass through |
| `Validate` | Validates each item — invalid items trigger callback, valid items continue |
| `Chunk` | Splits sequence into batches of size n — errors pass through untouched |
| `Flatten` | Flattens `iter.Seq2[[]T, error]` into a flat sequence |
| `Distinct` | Removes duplicates keeping first occurrence — errors pass through untouched |
| `Contains` | Returns true if target exists — stops as soon as match is found |
| `ForEach` | Runs side effect on each item — stops on upstream error |
| `Reverse` | Buffers full sequence then yields in reverse order |
| `Drain` | Consumes sequence calling fn on each item — stops on first error |
| `FilterSeq` | Lazy filter for plain `iter.Seq[T]` |
| `MapSeq` | Lazy map for plain `iter.Seq[T]` |
| `TakeSeq` | Lazy take for plain `iter.Seq[T]` |
| `FlatMapSeq` | Lazy flat map for plain `iter.Seq[T]` |
| `TakeWhileSeq` | Yields items while predicate is true for plain `iter.Seq[T]` |
| `ZipSeq` | Combines two plain `iter.Seq` sequences into pairs |
| `ValidateSeq` | Validates each item in plain `iter.Seq[T]` |
| `ChunkSeq` | Splits plain `iter.Seq[T]` into batches of size n |
| `FlattenSeq` | Flattens plain `iter.Seq[[]T]` into a flat sequence |
| `DistinctSeq` | Removes duplicates from plain `iter.Seq[T]` |
| `ContainsSeq` | Returns true if target exists in plain `iter.Seq[T]` |
| `ForEachSeq` | Runs side effect on each item in plain `iter.Seq[T]` |
| `ReverseSeq` | Buffers full plain `iter.Seq[T]` then yields in reverse order |
| `DrainSeq` | Consumes plain `iter.Seq[T]` calling fn on each item |

### parallel

| Function | Description |
|---|---|
| `ParallelMap` | Concurrent unordered map for `iter.Seq2[T, error]` — n workers, errors pass through |
| `BatchMap` | Sequential batch processing for `iter.Seq2[T, error]` — no goroutines |
| `OrderedParallelMap` | Concurrent ordered map for `iter.Seq2[T, error]` — preserves input order |
| `ParallelMapSeq` | Concurrent unordered map for plain `iter.Seq[T]` — n workers |
| `BatchMapSeq` | Sequential batch processing for plain `iter.Seq[T]` — no goroutines |
| `OrderedParallelMapSeq` | Concurrent ordered map for plain `iter.Seq[T]` — preserves input order |

### resilience

| Function / Type | Description |
|---|---|
| `Retry` | Retries fn up to MaxAttempts — only retries errors wrapped with `Retryable()` |
| `Backoff` | Configurable exponential backoff — 50-100% jitter range prevents thundering herd |
| `DefaultRetry` | 3 attempts, 100ms base, 30s max, jitter enabled |
| `DefaultBackoff` | 100ms base, 30s max, 2x multiplier, jitter enabled |
| `RetryableError` | Wraps an error to signal it should be retried |
| `Retryable` | Helper to wrap errors as retryable |
| `IsRetryable` | Reports whether an error is retryable |
| `CircuitBreaker` | Opens after n failures, half-open after timeout — one trial request at a time |
| `NewCircuitBreaker` | Creates a new circuit breaker with maxFailures and timeout |
| `Stats` | Runtime snapshot — requests, failures, successes, rejected, state |

### sources

| Function | Description |
|---|---|
| `DBRows` | Lazy DB rows from any `*sql.DB` or `*sql.Tx` — variadic query args |
| `CSVRows` | Lazy CSV rows from any `io.Reader` — validates consistent column count |
| `JSONLines` | Lazy JSONL decoding from any `io.Reader` — skips empty lines, surfaces decode errors |
| `JSONLinesFile` | Lazy JSONL decoding from file path |
| `Lines` | Lazy lines from any `io.Reader` |
| `FileLines` | Lazy lines from file path |
| `Stdin` | Lazy lines from standard input |

### root package (vortex)

| Symbol | Description |
|---|---|
| `vortex.Error` | Wraps all library errors with operation name and underlying cause |
| `vortex.Wrap` | Helper to construct `vortex.Error` — returns nil if err is nil |
| `vortex.ErrCancelled` | Returned when pipeline context is cancelled |
| `vortex.ErrValidation` | Returned when validation fails |
| `vortex.ErrCircuitOpen` | Returned when circuit breaker is open |

---

### Design decisions

- All functions accept `context.Context` as first parameter — cancellation is first class
- Sources return `iter.Seq2[T, error]` — errors never silently dropped
- `*Seq` variants accept plain `iter.Seq[T]` — for slices, generators, sources without errors
- All sources accept `io.Reader` — files, HTTP responses, network streams work without changes
- `DBRows` accepts `querier` interface — both `*sql.DB` and `*sql.Tx` work transparently
- Zero external dependencies — stdlib only
- Requires Go 1.23 or later
- Race tested — `CGO_ENABLED=1 go test -race ./...` passes clean