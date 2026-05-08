# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Vortex** is a zero-dependency Go 1.23 library for lazy, memory-efficient data pipelines using Go's `iter.Seq` and `iter.Seq2` interfaces. It processes data one item at a time without intermediate allocations, enabling billion-row datasets with flat memory usage.

## Commands

```bash
# Run all tests
go test ./...

# Run tests for a single package
go test ./iterx
go test ./parallel
go test ./sources

# Run a specific test
go test ./iterx -run TestFilter

# Run examples
go run examples/etl_pipeline/main.go
go run examples/create_excel_file/main.go
```

Requirements: Go 1.23+. No external dependencies.

## Architecture

Three packages with clear separation of concerns:

### `vortex/iterx` — Lazy transformations
Composable functions over `iter.Seq[T]` and `iter.Seq2[T, error]`. Every function returns a closure — no work happens until the caller ranges over it. Each iteration checks `ctx.Err()` for clean cancellation.

The API is split by whether errors need to flow:
- **Plain `iter.Seq[T]`** variants (suffix `Seq`): `FilterSeq`, `MapSeq`, `TakeSeq`, etc. — for error-free sequences (slices, custom generators).
- **Error-aware `iter.Seq2[T, error]`** variants (no suffix): `Filter`, `Map`, `Take`, etc. — pass errors through untouched; counting functions (e.g. `Take`) only count non-error items.

Most transformations are O(1) memory. The exceptions:
- `Reverse` buffers the entire sequence to invert it.
- `Distinct` keeps a set of every unique value seen so far (O(unique items)).
- `Chunk` holds at most one batch of size n.

### `vortex/parallel` — Concurrent processing
Three models, all with bounded goroutine pools and guaranteed cleanup on cancellation or early exit:
- **`ParallelMap`/`ParallelMapSeq`** — unordered, best for I/O-bound operations.
- **`ParallelMapErr`/`ParallelMapSeqErr`** — same but `fn` returns `(U, error)`; fn errors yielded inline.
- **`OrderedParallelMap`/`OrderedParallelMapSeq`** — preserves input order (may buffer if workers finish out of order).
- **`OrderedParallelMapErr`/`OrderedParallelMapSeqErr`** — ordered variant where `fn` returns `(U, error)`.
- **`BatchMap`/`BatchMapSeq`** — sequential batching without goroutines, for bulk DB operations.

### `vortex/sources` — Data source adapters
All return `iter.Seq2[T, error]` and accept `io.Reader` (works with files, HTTP responses, buffers):
- `CSVRows` — lazy CSV parsing with column count validation.
- `DBRows` — wraps `*sql.DB` or `*sql.Tx` with a caller-supplied scan function.
- `JSONLines` / `JSONLinesFile` — JSONL streaming, skips blank lines.
- `Lines` / `FileLines` — raw text line streaming.

### `errors.go` — Unified error model
- `vortex.Error{Op, Err}` wraps all library errors with operation context.
- Sentinel errors: `vortex.ErrCancelled`, `vortex.ErrValidation` — check with `errors.Is()`.
- `iterx.ValidationError[T]` carries the failing item and reason; unwraps to `vortex.ErrValidation`.
- `Wrap(op, err)` avoids nested chains; `WrapCancelled(op)` for context cancellations.

## Key Conventions

- **Context first**: every public function takes `context.Context` as its first parameter.
- **Lazy by default**: sequences are closures; computation starts only on `range`.
- **Errors pass through**: transformation functions never swallow errors — they re-yield them unchanged.
- **Goroutine safety**: `parallel` package guarantees no leaks; always pass a cancellable context.
- **Testing**: table-driven tests per function; mock I/O via `strings.NewReader`; cancellation via `context.WithCancel`.
