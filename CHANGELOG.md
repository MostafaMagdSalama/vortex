# Changelog

## v1.0.0 — 2024-XX-XX

First stable release.

### Added
- `interx.Filter` — lazy filter with context
- `interx.Map` — lazy map with context
- `interx.Take` — lazy take with context
- `interx.FlatMap` — lazy flat map with context
- `interx.TakeWhile` — lazy take while with context
- `interx.Zip` — zip two sequences with context
- `interx.Validate` — validate items with error callback
- `interx.Chunk` — split sequence into batches
- `interx.Flatten` — flatten sequence of slices
- `interx.Distinct` — remove duplicates
- `interx.Contains` — check if item exists
- `interx.ForEach` — iterate with side effects
- `interx.Reverse` — reverse a sequence
- `interx.Drain` — consume sequence with error handling
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
