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
| `vortex/iter` | Lazy sequences — Filter, Map, Take, FlatMap |
| `vortex/parallel` | Parallel processing — ParallelMap, BatchMap, WorkerPoolMap |
| `vortex/reduce` | Aggregation — AsyncReduce, Fold, Collect |
| `vortex/resilience` | Fault tolerance — Retry, Backoff, CircuitBreaker |
| `vortex/sources` | Data sources — CSVRows, DBRows, Lines, FileLines |

## Benchmarks

Tested on 1,000,000 rows CSV file on Windows.

| Approach | Peak memory | Notes |
|---|---|---|
| Eager (load all) | 287 MB | loads entire file into RAM |
| Lazy (vortex) | 3 MB | one row at a time |

**95x less memory** with lazy processing.

Memory stays flat at ~3 MB regardless of file size:
```
file size     eager peak     vortex peak
──────────    ──────────     ───────────
1M rows         287 MB           3 MB
10M rows       ~2.8 GB           3 MB
100M rows    out of memory        3 MB
```

The lazy approach processes one row at a time — memory is determined
by the size of one row, not the size of the file.

## Examples

### Lazy filtering
```go
import (
    "slices"
    "github.com/MostafaMagdSalama/vortex/iter"
)

numbers := slices.Values([]int{1, 2, 3, 4, 5})

for v := range iter.Filter(numbers, func(n int) bool { return n > 2 }) {
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

for v := range parallel.ParallelMap(numbers, func(n int) int {
    return n * 2
}, 4) {
    fmt.Println(v) // 2, 4, 6, 8, 10 (unordered)
}
```

### Batch processing
```go
for v := range parallel.BatchMap(numbers, func(batch []int) []int {
    // process whole batch at once — good for API/DB calls
    results := make([]int, len(batch))
    for i, v := range batch {
        results[i] = v * 2
    }
    return results
}, 3) {
    fmt.Println(v)
}
```

### CSV pipeline
```go
import (
    "github.com/MostafaMagdSalama/vortex/iter"
    "github.com/MostafaMagdSalama/vortex/sources"
)

file, _ := os.Open("users.csv")
defer file.Close()

// reads one row at a time — works on files of any size
for user := range iter.Filter(
    sources.CSVRows(file),
    func(row []string) bool { return row[3] == "active" },
) {
    fmt.Println(user)
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