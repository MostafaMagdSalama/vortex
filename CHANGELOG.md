# Changelog

## v1.0.0 — 2024-XX-XX

First stable release.

### Added
- `iterx.Filter` — lazy filter with context
- `iterx.Map` — lazy map with context
- `iterx.Take` — lazy take with context
- `iterx.FlatMap` — lazy flat map with context
- `iterx.TakeWhile` — lazy take while with context
- `iterx.Zip` — zip two sequences with context
- `iterx.Validate` — validate items with error callback
- `iterx.Chunk` — split sequence into batches
- `iterx.Flatten` — flatten sequence of slices
- `iterx.Distinct` — remove duplicates
- `iterx.Contains` — check if item exists
- `iterx.ForEach` — iterate with side effects
- `iterx.Reverse` — reverse a sequence
- `iterx.Drain` — consume sequence with error handling
- `parallel.ParallelMap` — concurrent map with context
- `parallel.BatchMap` — batch processing with context
- `parallel.OrderedParallelMap` — ordered parallel map with context
- `resilience.Retry` — retry with exponential backoff
- `resilience.CircuitBreaker` — circuit breaker with half-open fix
- `sources.DBRows` — lazy DB rows with context
- `sources.DBRows` — lazy DB rows with variadic query args and context
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
