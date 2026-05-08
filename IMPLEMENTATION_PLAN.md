# Vortex — Phase 2 & 3 Implementation Plan

This is a hand-off spec for finishing the production-readiness work on
`github.com/MostafaMagdSalama/vortex`. **Phase 1 is already merged**
(see "Phase 1 — done" at the bottom for context). This document covers
Phase 2 (API addition) and Phase 3 (documentation).

Read top to bottom and implement items in the order they appear. Every
item lists: target file(s), the code change (with diffs or full
listings), the test to add, and acceptance criteria.

---

## Repository facts you need

- Module path: `github.com/MostafaMagdSalama/vortex`
- Go version: `1.23.0`
- No external dependencies (the SQLite example imports `modernc.org/sqlite`
  but that's only inside `examples/` and is not in `go.mod`).
- Tests: `go test ./...`. The race detector requires cgo (Windows dev
  machine has no gcc) — run `-race` in CI.
- Branch in use: `main`. Make commits directly on `main` unless told
  otherwise.

### Package layout

```
vortex/
├── errors.go                 vortex.Error, ErrCancelled, ErrValidation, Wrap, WrapCancelled
├── iterx/                    Lazy transformations over iter.Seq / iter.Seq2
│   └── iter.go               Filter, Map, Take, FlatMap, Zip, Validate, ...
├── parallel/                 Concurrent processing
│   └── parallel.go           ParallelMap[Seq], OrderedParallelMap[Seq], BatchMap[Seq]
└── sources/                  Data source adapters
    ├── csv.go                CSVRows
    ├── db.go                 DBRows
    ├── jsonlines.go          JSONLines, JSONLinesFile
    └── lines.go              Lines, FileLines, Stdin
```

### Naming convention (already established)

- Plain `iter.Seq[T]` variants → suffix `Seq` (`FilterSeq`, `MapSeq`, `ParallelMapSeq`).
- Error-aware `iter.Seq2[T, error]` variants → no suffix (`Filter`, `Map`, `ParallelMap`).
- Phase 2 follows the same pattern with an `Err` suffix marking the
  fn-returns-error variant: `ParallelMapErr`, `ParallelMapSeqErr`, etc.

### Internal types (already defined in `parallel/parallel.go`)

```go
type task[T any] struct {
    index int
    value T
}

type result[U any] struct {
    index int
    value U
    err   error
}
```

Reuse these. Do not redefine.

### Op names for `vortex.Wrap`

The existing convention: `<package>.<FuncName>`. Examples already in the
codebase: `"parallel.ParallelMap"`, `"sources.CSVRows"`,
`"iterx.Filter"`. Phase 2 ops: `"parallel.ParallelMapErr"`,
`"parallel.OrderedParallelMapErr"`, `"parallel.ParallelMapSeqErr"`,
`"parallel.OrderedParallelMapSeqErr"`.

---

# Phase 2 — Add error-returning parallel variants

## Why

`ParallelMap`/`OrderedParallelMap` accept `fn func(T) U` — fn cannot
fail. The doc comment says the main use case is I/O-bound work, where
failures are routine. Today users have to panic, encode failures into
`U`, or skip the parallel package. Phase 2 closes that gap by adding
four new functions where fn returns `(U, error)`.

These are **additions, not replacements**. The existing four functions
keep their signatures and behavior. No breaking changes.

## Functions to add

All in `parallel/parallel.go`. Add them **after** the existing
`OrderedParallelMap` (i.e., at the bottom of the file).

| New function                | Input                      | fn signature        | Output                  |
|-----------------------------|----------------------------|---------------------|-------------------------|
| `ParallelMapErr`            | `iter.Seq2[T, error]`      | `func(T) (U, error)` | `iter.Seq2[U, error]`   |
| `OrderedParallelMapErr`     | `iter.Seq2[T, error]`      | `func(T) (U, error)` | `iter.Seq2[U, error]`   |
| `ParallelMapSeqErr`         | `iter.Seq[T]`              | `func(T) (U, error)` | `iter.Seq2[U, error]`   |
| `OrderedParallelMapSeqErr`  | `iter.Seq[T]`              | `func(T) (U, error)` | `iter.Seq2[U, error]`   |

**Note on output type:** Even when input is plain `iter.Seq[T]`, the
output is `iter.Seq2[U, error]` because fn can fail. This is the bridge
between error-free and error-aware pipelines.

## Behavior contract (applies to all four)

1. **`workers <= 0`** → `panic("vortex: workers must be > 0")`. Match
   the existing convention exactly (string identical).
2. **Pre-cancelled context** → yield one `(zero, vortex.WrapCancelled(op))`
   then return.
3. **Source `seq` errors** (Seq2 variants only) → wrap with
   `vortex.Wrap(op, err)`, yield inline as a `result[U]{err: ...}`,
   continue. Do **not** terminate the pipeline.
4. **fn returns non-nil error** → wrap with `vortex.Wrap(op, err)`,
   yield inline, continue. Do **not** terminate the pipeline.
5. **Worker panic** → `recover()`, surface as
   `vortex.Wrap(op, fmt.Errorf("worker panic: %v", r))`. Worker exits
   without processing further tasks (matches existing semantics — pool
   shrinks by one).
6. **Producer panic** (panic from `range seq`) → `recover()`, surface as
   `vortex.Wrap(op, fmt.Errorf("source panic: %v", r))` written to the
   `results` channel.
7. **Consumer breaks early** (yield returns false) → call `cancel()` so
   workers and producer drain, then return.
8. **Mid-iteration cancellation** → yield one
   `(zero, vortex.WrapCancelled(op))` then return.
9. **Ordering**:
   - `ParallelMapErr`: unspecified output order; errors are interleaved
     with values as they arrive.
   - `OrderedParallelMapErr`: yields in input order. Errors (from `seq`,
     fn, or panics) occupy the slot where the failing item would have
     been — the rest of the stream is **not** shifted.

## Channel sizing

Match the existing functions exactly:
- Unordered: `jobs := make(chan T, workers)`, `results := make(chan result[U], workers*2)`.
- Ordered: `tasks := make(chan task[T], workers)` (deliberately
  `workers`, not `workers*2`, per the comment in
  `parallel.go:326-327` — bounds the ordering buffer),
  `results := make(chan result[U], workers*2)`.

## Reference implementations

### `ParallelMapErr`

```go
// ParallelMapErr applies fn concurrently with `workers` goroutines. Both
// errors yielded by seq and errors returned by fn are wrapped with
// vortex.Wrap and yielded inline. Output order is unspecified.
//
// fn must be safe to call concurrently. A panic inside fn is recovered
// and surfaced as a worker error so it cannot crash the program.
func ParallelMapErr[T, U any](
    ctx context.Context,
    seq iter.Seq2[T, error],
    fn func(T) (U, error),
    workers int,
) iter.Seq2[U, error] {
    if workers <= 0 {
        panic("vortex: workers must be > 0")
    }
    return func(yield func(U, error) bool) {
        if ctx.Err() != nil {
            var zero U
            yield(zero, vortex.WrapCancelled("parallel.ParallelMapErr"))
            return
        }

        jobs := make(chan T, workers)
        results := make(chan result[U], workers*2)
        var wg sync.WaitGroup
        ctx, cancel := context.WithCancel(ctx)
        defer cancel()

        for i := 0; i < workers; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                defer func() {
                    if r := recover(); r != nil {
                        select {
                        case results <- result[U]{err: vortex.Wrap("parallel.ParallelMapErr", fmt.Errorf("worker panic: %v", r))}:
                        case <-ctx.Done():
                        }
                    }
                }()
                for {
                    select {
                    case <-ctx.Done():
                        return
                    case v, ok := <-jobs:
                        if !ok {
                            return
                        }
                        u, err := fn(v)
                        if err != nil {
                            select {
                            case results <- result[U]{err: vortex.Wrap("parallel.ParallelMapErr", err)}:
                            case <-ctx.Done():
                                return
                            }
                            continue
                        }
                        select {
                        case results <- result[U]{value: u}:
                        case <-ctx.Done():
                            return
                        }
                    }
                }
            }()
        }

        go func() {
            defer close(jobs)
            defer func() {
                if r := recover(); r != nil {
                    select {
                    case results <- result[U]{err: vortex.Wrap("parallel.ParallelMapErr", fmt.Errorf("source panic: %v", r))}:
                    case <-ctx.Done():
                    }
                }
            }()
            for v, err := range seq {
                if ctx.Err() != nil {
                    return
                }
                if err != nil {
                    select {
                    case results <- result[U]{err: vortex.Wrap("parallel.ParallelMapErr", err)}:
                    case <-ctx.Done():
                    }
                    continue
                }
                select {
                case jobs <- v:
                case <-ctx.Done():
                    return
                }
            }
        }()

        go func() {
            wg.Wait()
            close(results)
        }()

        for {
            if ctx.Err() != nil {
                var zero U
                yield(zero, vortex.WrapCancelled("parallel.ParallelMapErr"))
                return
            }
            r, ok := <-results
            if !ok {
                return
            }
            if r.err != nil {
                var zero U
                if !yield(zero, r.err) {
                    cancel()
                    return
                }
                continue
            }
            if !yield(r.value, nil) {
                cancel()
                return
            }
        }
    }
}
```

### `OrderedParallelMapErr`

Identical scaffolding to the existing `OrderedParallelMap`, except:
- fn signature is `func(T) (U, error)`.
- Worker stores fn's error in `result[U]{index: t.index, err: ...}`
  rather than calling `fn` and assuming success.
- Op name is `"parallel.OrderedParallelMapErr"`.
- The `currentIndex`-tracked panic recovery (already present in
  `OrderedParallelMap` at lines 444-452) stays the same.
- The producer's recover (added in Phase 1) stays the same; just rename
  the op string.

```go
// OrderedParallelMapErr applies fn concurrently with `workers` goroutines
// and yields results and errors in the original input order. Errors
// (from seq, fn, or worker panics) occupy the slot where the failing
// item would have been; surrounding items are not shifted.
func OrderedParallelMapErr[T, U any](
    ctx context.Context,
    seq iter.Seq2[T, error],
    fn func(T) (U, error),
    workers int,
) iter.Seq2[U, error] {
    if workers <= 0 {
        panic("vortex: workers must be > 0")
    }
    return func(yield func(U, error) bool) {
        if ctx.Err() != nil {
            var zero U
            yield(zero, vortex.WrapCancelled("parallel.OrderedParallelMapErr"))
            return
        }

        ctx, cancel := context.WithCancel(ctx)
        defer cancel()

        tasks := make(chan task[T], workers)
        results := make(chan result[U], workers*2)

        var wg sync.WaitGroup

        for i := 0; i < workers; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                var currentIndex int
                defer func() {
                    if r := recover(); r != nil {
                        select {
                        case results <- result[U]{index: currentIndex, err: vortex.Wrap("parallel.OrderedParallelMapErr", fmt.Errorf("worker panic: %v", r))}:
                        case <-ctx.Done():
                        }
                    }
                }()
                for {
                    select {
                    case <-ctx.Done():
                        return
                    case t, ok := <-tasks:
                        if !ok {
                            return
                        }
                        currentIndex = t.index
                        u, err := fn(t.value)
                        if err != nil {
                            select {
                            case results <- result[U]{index: t.index, err: vortex.Wrap("parallel.OrderedParallelMapErr", err)}:
                            case <-ctx.Done():
                                return
                            }
                            continue
                        }
                        select {
                        case results <- result[U]{index: t.index, value: u}:
                        case <-ctx.Done():
                            return
                        }
                    }
                }
            }()
        }

        go func() {
            defer close(tasks)
            i := 0
            defer func() {
                if r := recover(); r != nil {
                    select {
                    case results <- result[U]{index: i, err: vortex.Wrap("parallel.OrderedParallelMapErr", fmt.Errorf("source panic: %v", r))}:
                    case <-ctx.Done():
                    }
                }
            }()

            for v, err := range seq {
                if ctx.Err() != nil {
                    return
                }
                if err != nil {
                    select {
                    case results <- result[U]{index: i, err: vortex.Wrap("parallel.OrderedParallelMapErr", err)}:
                        i++
                    case <-ctx.Done():
                        return
                    }
                    continue
                }

                select {
                case tasks <- task[T]{index: i, value: v}:
                    i++
                case <-ctx.Done():
                    return
                }
            }
        }()

        go func() {
            wg.Wait()
            close(results)
        }()

        next := 0
        buffer := map[int]result[U]{}

        for {
            select {
            case <-ctx.Done():
                var zero U
                yield(zero, vortex.WrapCancelled("parallel.OrderedParallelMapErr"))
                return
            case r, ok := <-results:
                if !ok {
                    return
                }
                buffer[r.index] = r
                for {
                    buffered, exists := buffer[next]
                    if !exists {
                        break
                    }
                    delete(buffer, next)
                    if buffered.err != nil {
                        var zero U
                        if !yield(zero, buffered.err) {
                            cancel()
                            return
                        }
                    } else if !yield(buffered.value, nil) {
                        cancel()
                        return
                    }
                    next++
                }
            }
        }
    }
}
```

### `ParallelMapSeqErr` and `OrderedParallelMapSeqErr`

These accept `iter.Seq[T]` (no errors from the source). Implement them
**by adapting the seq into a Seq2 and delegating** to the Seq2 variants.
This avoids ~200 lines of near-identical code.

```go
// ParallelMapSeqErr is like ParallelMapErr but takes a plain iter.Seq[T].
// fn errors and worker panics are surfaced through the returned
// iter.Seq2[U, error]. Output order is unspecified.
func ParallelMapSeqErr[T, U any](
    ctx context.Context,
    seq iter.Seq[T],
    fn func(T) (U, error),
    workers int,
) iter.Seq2[U, error] {
    return ParallelMapErr(ctx, seqToSeq2(seq), fn, workers)
}

// OrderedParallelMapSeqErr is the ordered variant of ParallelMapSeqErr.
func OrderedParallelMapSeqErr[T, U any](
    ctx context.Context,
    seq iter.Seq[T],
    fn func(T) (U, error),
    workers int,
) iter.Seq2[U, error] {
    return OrderedParallelMapErr(ctx, seqToSeq2(seq), fn, workers)
}
```

**Caveat:** the op name in surfaced errors will be `"parallel.ParallelMapErr"`,
not `"parallel.ParallelMapSeqErr"`. That's a minor mismatch but it's the
trade-off for not duplicating ~200 lines. If exact op names matter,
inline the implementations instead and replace the op string in every
`vortex.Wrap` / `vortex.WrapCancelled` call.

**Helper required.** Add this private helper at the top of
`parallel/parallel.go`, just below the `task`/`result` types:

```go
// seqToSeq2 lifts a plain iter.Seq into iter.Seq2 with always-nil error.
func seqToSeq2[T any](seq iter.Seq[T]) iter.Seq2[T, error] {
    return func(yield func(T, error) bool) {
        for v := range seq {
            if !yield(v, nil) {
                return
            }
        }
    }
}
```

Note: `parallel_test.go` already has a test-local `seq2FromSlice` (line
17) — that one takes a slice. The production helper above takes a
`Seq`. Different signatures, no conflict.

## Tests to add

All in `parallel/parallel_test.go`. Append after the panic-recovery
tests added in Phase 1.

### 1. Happy path

```go
func TestParallelMapErr(t *testing.T) {
    input := seq2FromSlice([]int{1, 2, 3, 4, 5})
    var got []int
    for v, err := range parallel.ParallelMapErr(
        context.Background(), input,
        func(n int) (int, error) { return n * 2, nil },
        3,
    ) {
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        got = append(got, v)
    }
    sort.Ints(got)
    if !slices.Equal(got, []int{2, 4, 6, 8, 10}) {
        t.Fatalf("got %v", got)
    }
}

func TestOrderedParallelMapErr(t *testing.T) {
    input := seq2FromSlice([]int{1, 2, 3, 4, 5})
    var got []int
    for v, err := range parallel.OrderedParallelMapErr(
        context.Background(), input,
        func(n int) (int, error) { return n * 2, nil },
        3,
    ) {
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        got = append(got, v)
    }
    if !slices.Equal(got, []int{2, 4, 6, 8, 10}) {
        t.Fatalf("got %v (order should be preserved)", got)
    }
}
```

### 2. fn errors are wrapped and yielded inline

```go
func TestParallelMapErr_FnErrorsWrapped(t *testing.T) {
    input := seq2FromSlice([]int{1, 2, 3, 4, 5})
    boom := errors.New("boom")
    var got []int
    var errs int
    for v, err := range parallel.ParallelMapErr(
        context.Background(), input,
        func(n int) (int, error) {
            if n == 3 {
                return 0, boom
            }
            return n * 2, nil
        },
        2,
    ) {
        if err != nil {
            errs++
            if !errors.Is(err, boom) {
                t.Fatalf("expected wrapped boom, got %v", err)
            }
            continue
        }
        got = append(got, v)
    }
    sort.Ints(got)
    if !slices.Equal(got, []int{2, 4, 8, 10}) {
        t.Fatalf("got values %v", got)
    }
    if errs != 1 {
        t.Fatalf("expected 1 error, got %d", errs)
    }
}
```

### 3. Order preserved with mixed errors

```go
func TestOrderedParallelMapErr_OrderPreservedAcrossErrors(t *testing.T) {
    input := seq2FromSlice([]int{1, 2, 3, 4, 5})
    boom := errors.New("boom")
    type out struct {
        v   int
        err bool
    }
    var got []out
    for v, err := range parallel.OrderedParallelMapErr(
        context.Background(), input,
        func(n int) (int, error) {
            if n == 3 {
                return 0, boom
            }
            return n * 10, nil
        },
        4,
    ) {
        if err != nil {
            got = append(got, out{err: true})
            continue
        }
        got = append(got, out{v: v})
    }
    want := []out{{v: 10}, {v: 20}, {err: true}, {v: 40}, {v: 50}}
    if len(got) != len(want) {
        t.Fatalf("len mismatch: got %v want %v", got, want)
    }
    for i := range want {
        if got[i] != want[i] {
            t.Fatalf("at %d: got %+v, want %+v", i, got[i], want[i])
        }
    }
}
```

### 4. Source errors pass through

```go
func TestParallelMapErr_SourceErrorsPassThrough(t *testing.T) {
    boom := errors.New("source-boom")
    input := seq2WithError([]int{1, 2, 3}, 1, boom) // existing helper, line 27 of parallel_test.go
    var sawErr bool
    for _, err := range parallel.ParallelMapErr(
        context.Background(), input,
        func(n int) (int, error) { return n, nil },
        2,
    ) {
        if err != nil && errors.Is(err, boom) {
            sawErr = true
        }
    }
    if !sawErr {
        t.Fatal("expected source error to surface")
    }
}
```

### 5. Worker panic recovered as error

```go
func TestParallelMapErr_RecoversWorkerPanic(t *testing.T) {
    input := seq2FromSlice([]int{1, 2, 3, 4, 5})
    var sawPanic bool
    for _, err := range parallel.ParallelMapErr(
        context.Background(), input,
        func(n int) (int, error) {
            if n == 3 {
                panic("worker explosion")
            }
            return n, nil
        },
        2,
    ) {
        if err != nil && strings.Contains(err.Error(), "worker panic") {
            sawPanic = true
        }
    }
    if !sawPanic {
        t.Fatal("expected worker panic to surface as error")
    }
}
```

### 6. Pre-cancelled context

```go
func TestParallelMapErr_PreCancelled(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    var gotErr error
    for _, err := range parallel.ParallelMapErr(
        ctx, seq2FromSlice([]int{1, 2, 3}),
        func(n int) (int, error) { return n, nil },
        2,
    ) {
        if err != nil {
            gotErr = err
            break
        }
    }
    if !errors.Is(gotErr, vortex.ErrCancelled) {
        t.Fatalf("expected ErrCancelled, got %v", gotErr)
    }
}
```

### 7. Plain-Seq variants delegate correctly

```go
func TestParallelMapSeqErr(t *testing.T) {
    boom := errors.New("boom")
    var got []int
    var errs int
    for v, err := range parallel.ParallelMapSeqErr(
        context.Background(), slices.Values([]int{1, 2, 3, 4, 5}),
        func(n int) (int, error) {
            if n == 3 {
                return 0, boom
            }
            return n * 2, nil
        },
        2,
    ) {
        if err != nil {
            errs++
            continue
        }
        got = append(got, v)
    }
    sort.Ints(got)
    if !slices.Equal(got, []int{2, 4, 8, 10}) || errs != 1 {
        t.Fatalf("got values %v errs %d", got, errs)
    }
}

func TestOrderedParallelMapSeqErr(t *testing.T) {
    var got []int
    for v, err := range parallel.OrderedParallelMapSeqErr(
        context.Background(), slices.Values([]int{1, 2, 3}),
        func(n int) (int, error) { return n * 2, nil },
        2,
    ) {
        if err != nil {
            t.Fatal(err)
        }
        got = append(got, v)
    }
    if !slices.Equal(got, []int{2, 4, 6}) {
        t.Fatalf("got %v", got)
    }
}
```

### 8. Workers <= 0 panics

```go
func TestParallelMapErr_ZeroWorkersPanics(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Fatal("expected panic for workers <= 0")
        }
    }()
    parallel.ParallelMapErr(context.Background(), seq2FromSlice([]int{1}),
        func(n int) (int, error) { return n, nil }, 0)
}
```

(Add the same for the other three new functions.)

### Imports needed in `parallel_test.go`

The existing test file already imports `errors`, `slices`, `sort`,
`strings`, `testing`. Make sure these stay; do not remove anything.
The new tests use them all.

## Acceptance criteria for Phase 2

- [ ] Four new exported functions in `parallel/parallel.go` with the
      signatures listed above.
- [ ] One new private helper `seqToSeq2` in `parallel/parallel.go`.
- [ ] All eight reference behaviors (workers<=0 panic, pre-cancellation,
      source errors, fn errors, worker panic, producer panic, consumer
      break, mid-iter cancel) hold for `ParallelMapErr` and
      `OrderedParallelMapErr`.
- [ ] Order is preserved by `OrderedParallelMapErr` even when errors
      and values are interleaved.
- [ ] Plain-Seq variants delegate to Seq2 variants (no duplicated
      worker-pool logic).
- [ ] All new tests pass: `go test ./parallel/...`.
- [ ] Full suite still green: `go test ./...`.
- [ ] `go vet ./parallel/...` clean (the existing vet warnings in some
      `iterx/*_test.go` files are unrelated and out of scope).

## Suggested commit

```
feat(parallel): add ParallelMapErr / OrderedParallelMapErr

Allow the per-item function in parallel maps to return an error.
Errors from fn are wrapped with vortex.Wrap and yielded inline
alongside source-seq errors and recovered worker/producer panics.
OrderedParallelMapErr preserves input order across mixed
value/error output.

The Seq variants (ParallelMapSeqErr, OrderedParallelMapSeqErr)
delegate to the Seq2 variants via a small seqToSeq2 helper.

The existing ParallelMap / OrderedParallelMap signatures and
behavior are unchanged.
```

---

# Phase 3 — Documentation

Six small doc edits. Each one is a few lines. None require code logic
changes.

## 3.1 `iterx/distinct.go` — note O(n) memory

Update both `DistinctSeq` and `Distinct` doc comments. Current text:

```go
// DistinctSeq filters out duplicate values keeping only the first occurrence.
```

Replace with:

```go
// DistinctSeq filters out duplicate values keeping only the first occurrence.
//
// Memory: O(unique items). The internal seen-set grows for the lifetime
// of the iteration and is not bounded. For high-cardinality streams,
// consider chunking with iterx.Chunk and deduplicating per chunk, or use
// a probabilistic structure outside the pipeline.
```

Same paragraph appended to `Distinct` (the Seq2 variant) at the same
file.

## 3.2 `CLAUDE.md` — fix the "O(1) memory" claim

Find this line in the iterx section:

```
`Reverse` is the only function that buffers the full sequence. All others are O(1) memory.
```

Replace with:

```
Most transformations are O(1) memory. The exceptions:
- `Reverse` buffers the entire sequence to invert it.
- `Distinct` keeps a set of every unique value seen so far (O(unique items)).
- `Chunk` holds at most one batch of size n.
```

## 3.3 `sources/lines.go`, `sources/jsonlines.go`, `sources/csv.go` — cancellation caveat

Add this paragraph to the package-level doc comment on each source
function (`Lines`, `FileLines`, `JSONLines`, `JSONLinesFile`, `CSVRows`).
For functions without an existing doc paragraph, add one.

```
// Cancellation: ctx is checked between records. A blocked read on the
// underlying io.Reader (e.g. a slow HTTP body) will not be interrupted
// by ctx alone — close the reader to unblock it. For HTTP, do that by
// canceling the request context that produced the response, or by
// calling resp.Body.Close from another goroutine.
```

For `DBRows` in `sources/db.go`, the situation is different —
`db.QueryContext` does propagate ctx — so do **not** add this paragraph
there. Optionally add a one-line note instead:

```
// Cancellation: ctx is honored by both the QueryContext call and the
// per-row scan loop.
```

## 3.4 `parallel/doc.go` — worker panic semantics

The file `parallel/doc.go` exists (per the glob output). Add a section
to the package documentation:

```go
// # Panic safety
//
// Worker, producer, and source-seq panics are recovered and surfaced as
// wrapped errors through the returned iter.Seq2 (for Seq2 and *Err
// variants) or cause the pipeline to cancel cleanly (for plain-Seq
// variants without an error channel). A worker that panics exits
// without processing further items — the pool shrinks by one for the
// rest of that call. Throughput degrades but ordering and error
// delivery remain correct.
```

If `parallel/doc.go` only contains `package parallel`, prepend the
comment block before that line. If it has existing package-level docs,
append this section after them.

## 3.5 `iterx/doc.go` — extend the existing "# Memory notes" section

`iterx/doc.go` already has this paragraph:

```
// # Memory notes
//
// All functions stream values one at a time except Reverse, which buffers the
// full sequence in memory before yielding results. Reverse is not suitable for
// infinite or very large sequences.
```

**Replace it with:**

```
// # Memory notes
//
// Most functions stream values one at a time and use O(1) memory. The
// exceptions:
//
//   - Reverse — buffers the entire sequence before yielding (O(n)). Not
//     suitable for infinite or very large sequences.
//   - Distinct / DistinctSeq — keeps a hash set of every unique value
//     seen so far (O(unique items)). For high-cardinality streams,
//     deduplicate per chunk (via iterx.Chunk) or use a probabilistic
//     structure outside the pipeline.
//   - Chunk / ChunkSeq — holds at most one batch of size n at a time.
```

Do not duplicate the section — replace the existing one in place.

## 3.6 `parallel/parallel.go` — comment fix for `BatchMap` (Phase 1 follow-up)

Phase 1 changed `batch = batch[:0]` to `batch = make([]T, 0, batchSize)`
in both `BatchMapSeq` and `BatchMap`, with an inline rationale comment.
Verify the rationale comment exists on both. If only one has it, copy
it to the other. (Phase 1 should already have done this — this is a
double-check item.)

## Tests for Phase 3

No new tests needed. Documentation only. Verify with:

```
go test ./...        # nothing should break
go vet ./...         # should still pass
godoc -http=:6060    # eyeball the doc comments render correctly (optional)
```

## Suggested commit

```
docs: clarify memory bounds, cancellation, and panic semantics

- iterx/Distinct: note O(unique items) memory, not O(1).
- sources: document that ctx is checked between records, not during
  blocking reads on the underlying io.Reader.
- parallel: document worker panic shrinks the pool by one and is
  surfaced as a wrapped error.
- CLAUDE.md and package docs aligned with these clarifications.
```

---

# Phase 1 — done (context for the next implementer)

Already merged. Listed here so you don't redo it.

| File                       | Change                                                                                                                            |
|----------------------------|-----------------------------------------------------------------------------------------------------------------------------------|
| `sources/csv.go`           | `CSVRows` distinguishes `*csv.ParseError` (recoverable, continue) from other errors (terminal, yield once and return).            |
| `iterx/iter.go`            | `Take` and `TakeSeq` short-circuit when `n <= 0` — no source pull.                                                                |
| `parallel/parallel.go`     | `BatchMap`/`BatchMapSeq` allocate a fresh batch instead of `batch[:0]`, removing aliasing when fn returns its input slice.        |
| `parallel/parallel.go`     | All four producer goroutines (`ParallelMapSeq`, `ParallelMap`, `OrderedParallelMapSeq`, `OrderedParallelMap`) now `recover()` from source-seq panics. Seq2 variants surface the panic as a wrapped error; plain-Seq variants cancel the pool. |

New tests added in Phase 1 (do not duplicate):
- `TestCSVRows_StopsOnIOError`, `TestCSVRows_ContinuesOnParseError`
- `TestTake_ZeroDoesNotPullSource`
- `TestBatchMapSeq_NoAliasingWhenFnReturnsInput`, `TestBatchMap_NoAliasingWhenFnReturnsInput`
- `TestParallelMapSeq_RecoversSourcePanic`, `TestParallelMap_RecoversSourcePanicAsError`
- `TestOrderedParallelMapSeq_RecoversSourcePanic`, `TestOrderedParallelMap_RecoversSourcePanicAsError`

Helper added in Phase 1 tests: `panicSeq` and `panicSeq2` in
`parallel/parallel_test.go`. Reuse them for any panic-related Phase 2
tests rather than redefining.

---

# Verification checklist (run after each phase)

```bash
# build + unit tests
go test ./...

# static analysis
go vet ./...

# race detector — requires cgo (gcc/clang on PATH)
# This MUST pass in CI before tagging a release.
CGO_ENABLED=1 go test -race ./...

# stress the parallel package
go test -count=20 -run "Parallel|Ordered|BatchMap" ./parallel
```

Pre-existing `go vet` warnings in `iterx/{flatten,reverse,seq_variants,zip}_test.go`
about unused `cancel` functions are out of scope for these phases.
Track separately if desired; do not let them block Phase 2 or 3.

---

# Out of scope (do not do)

- Do **not** change the existing `ParallelMap` / `OrderedParallelMap`
  signatures or rewire them on top of the new `Err` variants. The
  resulting op-name change in error messages is a breaking change.
- Do **not** add a `*Err` suffix to `BatchMap` — batch processing
  already supports per-batch error handling via the slice return; a
  fn-returns-error variant there is a separate proposal.
- Do **not** introduce new dependencies. The library is zero-dep and
  must stay that way.
- Do **not** modify `examples/`. The examples are demos, not part of
  the public surface.
