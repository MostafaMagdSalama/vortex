# Structured Concurrency in Vortex

Vortex's `parallel` package makes it trivial to fan-out sequence processing to multiple goroutines while maintaining strict limits on memory usage, concurrency, and goroutine lifecycles.

## The Core Philosophy

In Vortex, **concurrency is bounded**. The `parallel` package guarantees:
1. **No leaking goroutines**: Schedulers and worker pools always clean up completely after completion or context cancellation.
2. **Predictable memory**: Parallel pipelines pull from iterating wrappers (like `csv` or `db`) lazily. It never loads the entire dataset into RAM first.
3. **Panic safety**: By keeping execution explicitly bounded within `iter.Seq` wrappers, panics are contained.

## Available Processors

### `parallel.ParallelMapSeq` and `parallel.ParallelMap`

`ParallelMapSeq[T, U any](ctx, seq, mapFn, workers)` applies a transformation function concurrently across `n` workers for plain `iter.Seq[T]` inputs.

`ParallelMap[T, U any](ctx, seq, mapFn, workers)` does the same for `iter.Seq2[T, error]` inputs and passes upstream errors through inline.

**When to use it:**
- I/O-bound operations (calling external APIs, reading files) where order doesn't matter.
- Use `ParallelMapSeq` for `slices.Values(...)` and other non-erroring generators.
- Use `ParallelMap` when the input comes from `vortex/sources` or any other `iter.Seq2[T, error]`.
- **Order guarantee**: both versions yield results as soon as they finish. Results will likely be out of order from the input sequence.

```go
results := parallel.ParallelMapSeq(ctx, urls, func(url string) Status {
    return fetchStatus(url)
}, 10) // 10 concurrent requests
```

### `parallel.OrderedParallelMapSeq` and `parallel.OrderedParallelMap`

`OrderedParallelMapSeq[T, U any](ctx, seq, mapFn, workers)` preserves input order for plain `iter.Seq[T]` inputs.

`OrderedParallelMap[T, U any](ctx, seq, mapFn, workers)` preserves input order for `iter.Seq2[T, error]` inputs while also preserving inline errors in the same order.

**When to use it:**
- Slower I/O or CPU operations where output MUST remain sorted (like maintaining file line numbers).
- **Trade-off**: both ordered versions require buffering internally. If task #1 is very slow, but tasks #2–#100 finish instantly, the library must buffer #2–#100 in memory until #1 completes to yield them in order.

### `parallel.BatchMapSeq` and `parallel.BatchMap`

`BatchMapSeq[T, U any](ctx, seq, batchFn, batchSize)` collects plain `iter.Seq[T]` values into slices of size `n`, then passes each batch to your function.

`BatchMap[T, U any](ctx, seq, batchFn, batchSize)` does the same for `iter.Seq2[T, error]` inputs and yields upstream errors inline instead of dropping them.

**When to use it:**
- Database bulk INSERTS or API bulk updates. Whenever the downstream system handles chunks better than individual requests.

```go
// Groups 500 rows at a time and inserts them
parallel.BatchMap(ctx, dbRows, func(batch []User) []User {
    bulkInsert(batch)
    return batch
}, 500)
```

## Context Cancellation

Every parallel operation in Vortex immediately halts when the `context.Context` is cancelled. Background goroutines will shut down gracefully, channels will close, and memory will be garbage collected safely.
