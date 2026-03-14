# vortex

A Go library for lazy iterators, parallel processing, and resilient pipelines.
Built on Go 1.23's `iter.Seq` — zero external dependencies.

## Install
```bash
go get github.com/MostafaMagdSalama/vortex@latest
```

## Requirements

Go 1.23 or later.

## Packages

| Package | What it does |
|---|---|
| `vortex/iterx` | Lazy sequences — Filter, Map, Take, FlatMap |
| `vortex/parallel` | Parallel processing — ParallelMap, BatchMap, WorkerPoolMap |
| `vortex/resilience` | Fault tolerance — Retry, Backoff, CircuitBreaker |
| `vortex/sources` | Data sources — CSVRows, DBRows, Lines, FileLines |

## Benchmarks

### CSV file — 1,000,000 rows (Windows)

| Approach | Peak memory | Notes |
|---|---|---|
| Eager (load all) | 287 MB | loads entire file into RAM |
| Lazy (vortex) | 3 MB | one row at a time |

**95x less memory** with lazy processing.
```
file size     eager peak     vortex peak
──────────    ──────────     ───────────
1M rows         287 MB           3 MB
10M rows       ~2.8 GB           3 MB
100M rows    out of memory        3 MB
```

### Database — 1,000,000 rows (Windows)

| Approach | Peak memory | Rows read | Notes |
|---|---|---|---|
| Eager (load all) | 247 MB | 1,000,000 | loads all rows before processing |
| Lazy (vortex) | 397 KB | 10 | stops the moment it has what it needs |

**636x less memory** with lazy processing.
```
without vortex:
  memory after loading all rows:  134 MB
  memory after filtering:         204 MB
  memory after extracting names:  247 MB
  rows loaded: 1,000,000

with vortex:
  memory after creating source:   393 KB
  memory after defining filter:   393 KB
  memory after defining map:      393 KB
  memory after defining take:     393 KB
  peak memory: 397 KB
  rows read: 10 out of 1,000,000
```

The lazy approach stops reading from the database the moment
it has enough results — it never touches the remaining 999,990 rows.

## Examples

### Lazy filtering
```go
import (
    "slices"
    "github.com/MostafaMagdSalama/vortex/iterx"
)

numbers := slices.Values([]int{1, 2, 3, 4, 5})

for v := range iterx.Filter(context.Background(), numbers, func(n int) bool { return n > 2 }) {
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

`sources.CSVRows` accepts any `io.Reader` and returns a lazy sequence of rows.
The source is always streamed - never fully loaded into memory.

### Local file

```go
file, err := os.Open("users.csv")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

for row, err := range sources.CSVRowsWithError(ctx, file) {
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

    for row, err := range sources.CSVRowsWithError(r.Context(), file) {
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

for row, err := range sources.CSVRowsWithError(ctx, resp.Body) {
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

All three sources satisfy `io.Reader`. `CSVRowsWithError` reads one record at a
time regardless of whether the source is a file, an HTTP upload, or a network
stream.

```
multipart upload  -> io.Reader -> CSVRowsWithError -> one row at a time
presigned URL     -> io.Reader -> CSVRowsWithError -> one row at a time
local file        -> io.Reader -> CSVRowsWithError -> one row at a time
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
