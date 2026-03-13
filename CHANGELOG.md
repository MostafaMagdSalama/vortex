# Changelog

## v1.0.0 — 2024-XX-XX

First stable release.

### Added
- `iter.Filter` — lazy filter with context
- `iter.Map` — lazy map with context
- `iter.Take` — lazy take with context
- `iter.FlatMap` — lazy flat map with context
- `iter.TakeWhile` — lazy take while with context
- `iter.Zip` — zip two sequences with context
- `iter.Validate` — validate items with error callback
- `iter.Chunk` — split sequence into batches
- `iter.Flatten` — flatten sequence of slices
- `iter.Distinct` — remove duplicates
- `iter.Contains` — check if item exists
- `iter.ForEach` — iterate with side effects
- `iter.Reverse` — reverse a sequence
- `iter.Drain` — consume sequence with error handling
- `parallel.ParallelMap` — concurrent map with context
- `parallel.BatchMap` — batch processing with context
- `parallel.WorkerPoolMap` — worker pool map with context
- `resilience.Retry` — retry with exponential backoff
- `resilience.CircuitBreaker` — circuit breaker with half-open fix
- `sources.DBRows` — lazy DB rows with context
- `sources.DBRowsWithArgs` — lazy DB rows with args and context
- `sources.CSVRows` — lazy CSV rows with context
- `sources.FileLines` — lazy file lines with context
- `sources.Lines` — lazy lines from any reader
- `sources.Stdin` — lazy stdin lines

### Breaking changes
- context is required as first parameter on all functions

### Fixed
- `CircuitBreaker` half-open state allowed multiple concurrent trial requests
- `ParallelMap` deadlock on early stop
- `Lines` silent failure on oversized lines