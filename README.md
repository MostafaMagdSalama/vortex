# vortex

 vortex is a zero-dependency Go 1.23 library that brings lazy evaluation,
structured concurrency, and fault tolerance to data pipeline development.

Built on Go 1.23's iter.Seq and iter.Seq2 interfaces, vortex treats every
data source, database cursors, CSV streams, JSONL files, HTTP response
as a unified lazy sequence. Transformations compose without intermediate
allocations. Pipelines cancel cleanly through context propagation.
Workers coordinate without leaking goroutines.

The result is pipelines that scale from a single row to a billion rows
with flat memory, predictable latency, and production-grade error handling all without leaving idiomatic Go.

## Architecture

```text
 ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐     ┌──────────────────┐
 │     Source      │ ──► │ Transformation  │ ──► │ Transformation  │ ──► │     Terminal     │
 │    (iter.Seq)   │     │  (Filter/Map)   │     │  (Take/Chunk)   │     │  (Drain/Range)   │
 └─────────────────┘     └─────────────────┘     └─────────────────┘     └──────────────────┘
          ▲                       ▲                       ▲                        ▲
          │                       │                       │                        │
  slices.Values           iterx.FilterSeq         iterx.TakeSeq            iterx.DrainSeq
  sources.CSVRows         iterx.Filter            iterx.Take               iterx.Drain
  sources.DBRows          iterx.Map               iterx.Chunk              iterx.ForEach
  slices.Values           parallel.BatchMap       parallel.ParallelMap     for v := range
```

## Install
```bash
go get github.com/MostafaMagdSalama/vortex@latest
```

## Requirements

Go 1.24 or later.

## Packages

| Package | What it does |
|---|---|
| `vortex/iterx` | Lazy sequences — Filter, Map, Take, FlatMap |
| `vortex/parallel` | Parallel processing — ParallelMap, BatchMap, WorkerPoolMap |
| `vortex/resilience` | Fault tolerance — Retry, Backoff, CircuitBreaker |
| `vortex/sources` | Data sources — CSVRows, DBRows, Lines, FileLines |

`iterx` now includes paired APIs: use `*Seq` helpers like `FilterSeq` and `MapSeq` with plain `iter.Seq[T]`, and use the original names like `Filter` and `Map` with `iter.Seq2[T, error]`.

## Benchmarks

### CSV file — 1,000,000 rows (Windows)

| Approach | Peak memory | Rows read | Notes |
|---|---|---|---|
| Eager (load all) | 287 MB | 1,000,000 | loads entire file into RAM |
| Lazy (vortex) | 3 MB | 1,000,000 | streams one row at a time |

**95x less memory** with lazy processing.

<details>
<summary><b>View detailed benchmark scaling</b></summary>
<br>

```text
╔══════════════════════════════════════════╗
║          memory scaling data             ║
╚══════════════════════════════════════════╝
file size     eager peak     vortex peak
──────────    ──────────     ───────────
1M rows         287 MB           3 MB
10M rows       ~2.8 GB           3 MB
100M rows    out of memory       3 MB
```
</details>

### Database — 1,000,000 rows (Windows)

| Approach | Peak memory | Rows read | Notes |
|---|---|---|---|
| Eager (load all) | 247 MB | 1,000,000 | loads all rows before processing |
| Lazy (vortex) | ~397 KB | 10 | stops the moment it has what it needs |

**636x less memory** with lazy processing.

<details>
<summary><b>View detailed benchmark output</b></summary>
<br>

```text
╔══════════════════════════════════════════╗
║         with vortex (lazy)               ║
╚══════════════════════════════════════════╝
memory after creating source:   393 KB
memory after defining filter:   393 KB
memory after defining map:      393 KB
memory after defining take:     393 KB
peak memory: 397 KB
rows read: 10 out of 1,000,000

╔══════════════════════════════════════════╗
║         without vortex (eager)           ║
╚══════════════════════════════════════════╝
memory after loading all rows:  134 MB
memory after filtering:         204 MB
memory after extracting names:  247 MB
rows loaded: 1,000,000
```
</details>

The lazy approach stops reading from the database the moment
it has enough results — it never touches the remaining 999,990 rows.

### JSON Lines — 1,000,000 rows (Windows)

| Approach | Peak memory | Time | Notes |
|---|---|---|---|
| Eager (load all) | 194 MB | ~909 ms | decodes the entire file into memory before processing |
| Lazy (vortex) | 1 MB | ~24 ms | streams one line at a time |

**194x less memory** and **~37x faster** with lazy processing.

<details>
<summary><b>View detailed benchmark output</b></summary>
<br>

```text
╔══════════════════════════════════════════╗
║         with vortex (lazy)               ║
╚══════════════════════════════════════════╝
memory before:                  3 MB
memory after open:              3 MB
memory after JSONLines:         3 MB
memory after unwrap:            3 MB
memory after Filter:            3 MB
memory after Take:              3 MB
memory before range:            3 MB
memory after range:             1 MB

result:
errors found:   100
time:           24.7103ms
peak memory:    1 MB

╔══════════════════════════════════════════╗
║         without vortex (eager)           ║
╚══════════════════════════════════════════╝
memory before:                  274 KB
memory after ReadFile:          57 MB
memory after Split:             72 MB
memory after decode all:        168 MB
memory after filter:            194 MB
memory after take:              194 MB

result:
errors found:   100
total lines:    1000000
time:           909.2772ms
peak memory:    194 MB
```
</details>

## iterx API split

Use `iterx.FilterSeq`, `iterx.MapSeq`, `iterx.TakeSeq`, `iterx.DrainSeq`, and friends with plain `iter.Seq[T]` values such as `slices.Values(...)`.

Use `iterx.Filter`, `iterx.Map`, `iterx.Take`, `iterx.Drain`, and friends with `iter.Seq2[T, error]` values such as `sources.CSVRows`, `sources.DBRows`, and `sources.JSONLines`.

## Error Handling

Vortex provides a unified error handling architecture to ensure safety and transparency across pipelines. All library packages bubble up errors rather than failing silently.

### Expected Errors (`vortex.Error`)

All underlying errors (like network failures or database disconnects) are wrapped in `vortex.Error`. You can use `errors.As` to retrieve the original error and the operation that failed:

```go
import (
    "errors"
    "github.com/MostafaMagdSalama/vortex"
)

// inside your pipeline execution
err := iterx.Drain(ctx, mySeq, processor)

var vErr *vortex.Error
if errors.As(err, &vErr) {
    fmt.Printf("Failed operation: %s\n", vErr.Op)
    fmt.Printf("Underlying root cause: %v\n", vErr.Err)
}
```

### Sentinel Errors

Vortex exposes common failure states as sentinel errors in the root package. You can check for them using `errors.Is`:

*   `vortex.ErrCancelled`: Returned when the pipeline's context is cancelled.
*   `vortex.ErrValidation`: Returned when validation conditions fail (e.g. `iterx.Validate`).
*   `vortex.ErrCircuitOpen`: Returned by `resilience.CircuitBreaker` when the service is rejecting traffic.

```go
if errors.Is(err, vortex.ErrCircuitOpen) {
    // Serve from cache instead
}
```

## Examples

### Lazy filtering
```go
import (
    "slices"
    "github.com/MostafaMagdSalama/vortex/iterx"
)

numbers := slices.Values([]int{1, 2, 3, 4, 5})

for v := range iterx.FilterSeq(context.Background(), numbers, func(n int) bool { return n > 2 }) {
    fmt.Println(v) // 3, 4, 5
}
```

### Parallel processing
```go
import (
    "slices"
    "github.com/MostafaMagdSalama/vortex/parallel"
)

numbers := slices.Values([]int{1, 2, 3, 4, 5})

for v := range parallel.ParallelMap(context.Background(), numbers, func(n int) int {
    return n * 2
}, 4) {
    fmt.Println(v) // 2, 4, 6, 8, 10 (unordered)
}
```

### Ordered parallel processing
```go
for v := range parallel.OrderedParallelMap(context.Background(), numbers, func(n int) int {
    return n * 2
}, 4) {
    fmt.Println(v) // 2, 4, 6, 8, 10 (strictly ordered)
}
```

### Batch processing
```go
for v := range parallel.BatchMap(context.Background(), numbers, func(batch []int) []int {
    results := make([]int, len(batch))
    for i, v := range batch {
        results[i] = v * 2
    }
    return results
}, 3) {
    fmt.Println(v)
}
```

## CSV

### More iterx Examples

To see real-world, runnable examples for both the `*Seq` helpers and the error-aware `iter.Seq2` helpers, visit the `pkg.go.dev` documentation or explore `iterx/example_test.go` in the repository.

`sources.CSVRows` accepts any `io.Reader` and returns a lazy sequence of rows.
The source is always streamed - never fully loaded into memory.

### Local file

```go
file, err := os.Open("users.csv")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

for row, err := range sources.CSVRows(ctx, file) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(row)
}
```

### User uploads a CSV file (HTTP multipart)

```go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    file, _, err := r.FormFile("csv")
    if err != nil {
        http.Error(w, err.Error(), 400)
        return
    }
    defer file.Close()

    for row, err := range sources.CSVRows(r.Context(), file) {
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        fmt.Println(row)
    }
}
```

### Presigned URL or any HTTP URL

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://s3.amazonaws.com/bucket/file.csv", nil)
if err != nil {
    log.Fatal(err)
}

resp, err := http.DefaultClient.Do(req)
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

for row, err := range sources.CSVRows(ctx, resp.Body) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(row)
}
```

### Pipeline - CSV -> filter -> map -> take

```go
file, _ := os.Open("users.csv")
defer file.Close()

// skip header row, filter active users, take first 10 names
rows := sources.CSVRows(ctx, file)

first := true
names := iterx.Map(ctx,
    iterx.Take(ctx,
        iterx.Filter(ctx, rows, func(row []string) bool {
            if first {
                first = false
                return false
            }
            return row[3] == "active" // status column
        }),
        10,
    ),
    func(row []string) string {
        return row[1] // name column
    },
)

for name := range names {
    fmt.Println(name)
}
```

### Why it is always lazy

All three sources satisfy `io.Reader`. `CSVRows` reads one record at a
time regardless of whether the source is a file, an HTTP upload, or a network
stream.

```
multipart upload  -> io.Reader -> CSVRows -> one row at a time
presigned URL     -> io.Reader -> CSVRows -> one row at a time
local file        -> io.Reader -> CSVRows -> one row at a time
```

### Database pipeline
```go
import (
    "github.com/MostafaMagdSalama/vortex/iterx"
    "github.com/MostafaMagdSalama/vortex/sources"
)

// reads one row at a time — stops as soon as Take is satisfied
names := iterx.Map(
    context.Background(),
    iterx.Take(
        context.Background(),
        iterx.Filter(
            context.Background(),
            sources.DBRows(context.Background(), db, "SELECT id, name, email, status FROM users", scanUser),
            func(u User) bool { return u.Status == "active" },
        ),
        5,
    ),
    func(u User) string { return u.Name },
)

for name := range names {
    fmt.Println(name)
}
```

### Retry with backoff
```go
import (
    "context"
    "github.com/MostafaMagdSalama/vortex/resilience"
)

err := resilience.Retry(context.Background(), resilience.DefaultRetry, func() error {
    return callSomeAPI()
})
```

### Circuit breaker
```go
cb := resilience.NewCircuitBreaker(5, 10*time.Second)

err := cb.Execute(func() error {
    return callSomeAPI()
})

if errors.Is(err, resilience.ErrCircuitOpen) {
    // service is down, circuit is open
}
```

### Composing retry + circuit breaker
```go
cb := resilience.NewCircuitBreaker(5, 10*time.Second)

err := resilience.Retry(ctx, resilience.DefaultRetry, func() error {
    return cb.Execute(func() error {
        return callSomeAPI()
    })
})
```

## License

MIT
