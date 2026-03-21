# Example: ETL Pipeline (1,000,000 rows)

This example demonstrates a full **Extract → Transform → Load** pipeline over 1,000,000 rows using [vortex](https://github.com/MostafaMagdSalama/vortex) — with lazy evaluation and parallel transformation, all while keeping memory usage nearly flat throughout execution.

---

## What it does

1. **Extract** — streams rows one at a time from an in-memory SQLite database using `sources.DBRows`
2. **Transform** — uppercases each user's name and extracts the domain from their email, in parallel across 4 workers using `parallel.OrderedParallelMap`
3. **Load** — writes the transformed rows to a CSV file line by line

The pipeline is defined lazily — **nothing runs until the `for range` loop drives it**. At no point are all rows held in memory.

---

## How to run

```bash
go run main.go
```

This will:
1. Insert 1,000,000 rows into an in-memory SQLite database
2. Define the lazy pipeline (no work happens here)
3. Execute the pipeline, writing results to `users.csv`
4. Print memory stats and a summary, then clean up

---

## Benchmark

> Measured on Linux with 4 parallel workers.

### Results

| Metric | Value |
|---|---|
| Rows processed | 1,000,000 |
| Rows skipped | 0 |
| Duration | ~4.75s |
| Memory at pipeline start | ~361 KB alloc |
| Memory at pipeline end | ~388 KB alloc |
| Peak heap in-use (during execution) | ~1.1 MB |

Memory stays essentially **flat across all 1,000,000 rows** — the pipeline never accumulates more than a handful of rows in memory at any point.

<details>
<summary><b>View detailed benchmark output</b></summary>
<br>

```text
=== ETL Pipeline — 1,000,000 rows ===

[MEM] start                                         alloc=   305 KB   heapInUse=   728 KB   heapObjects=   469   sys=  6740 KB
inserting 1,000,000 rows...
done inserting
[MEM] after db setup (1,000,000 rows)               alloc=   360 KB   heapInUse=   912 KB   heapObjects=   536   sys= 11924 KB
[MEM] after file create                             alloc=   360 KB   heapInUse=   904 KB   heapObjects=   536   sys= 11924 KB

--- pipeline definition (all lazy, nothing runs yet) ---

[MEM] after DBRows defined     (lazy)               alloc=   360 KB   heapInUse=   904 KB   heapObjects=   537   sys= 11924 KB
[MEM] after OrderedParallelMap (lazy)               alloc=   361 KB   heapInUse=   912 KB   heapObjects=   538   sys= 11924 KB

--- pipeline execution (for range drives everything) ---

[MEM] progress:  100000 rows written                alloc=   388 KB   heapInUse=  1096 KB   heapObjects=   653   sys= 12628 KB
[MEM] progress:  200000 rows written                alloc=   390 KB   heapInUse=  1104 KB   heapObjects=   672   sys= 12628 KB
[MEM] progress:  300000 rows written                alloc=   400 KB   heapInUse=  1120 KB   heapObjects=   681   sys= 12628 KB
[MEM] progress:  400000 rows written                alloc=   400 KB   heapInUse=  1144 KB   heapObjects=   688   sys= 12628 KB
[MEM] progress:  500000 rows written                alloc=   401 KB   heapInUse=  1096 KB   heapObjects=   694   sys= 12628 KB
[MEM] progress:  600000 rows written                alloc=   403 KB   heapInUse=  1104 KB   heapObjects=   710   sys= 12628 KB
[MEM] progress:  700000 rows written                alloc=   403 KB   heapInUse=  1112 KB   heapObjects=   721   sys= 12628 KB
[MEM] progress:  800000 rows written                alloc=   409 KB   heapInUse=  1088 KB   heapObjects=   729   sys= 16724 KB
[MEM] progress:  900000 rows written                alloc=   409 KB   heapInUse=  1096 KB   heapObjects=   739   sys= 16724 KB
[MEM] progress: 1000000 rows written                alloc=   408 KB   heapInUse=  1088 KB   heapObjects=   681   sys= 16724 KB

[MEM] after pipeline completed                      alloc=   388 KB   heapInUse=   936 KB   heapObjects=   663   sys= 16724 KB

--- cleaning up ---

[MEM] after cleanup                                 alloc=   388 KB   heapInUse=   936 KB   heapObjects=   666   sys= 16724 KB

=== summary ===
rows written : 1000000
rows skipped : 0
duration     : 4.7542156s
```

</details>

### Why memory stays flat

Defining `sources.DBRows` and `parallel.OrderedParallelMap` allocates **less than 2 KB** combined — they only describe what will happen, not do it. When the `for range` loop starts, the pipeline pulls one row at a time: the source fetches it, the transform worker processes it, and the loader writes it to disk. At any moment only a small window of in-flight rows exists across the 4 workers — **no accumulation, no slices, no buffering of the full dataset**.

---

## Key vortex APIs used

| API | Purpose |
|---|---|
| `sources.DBRows` | Turns a `*sql.Rows` query into a lazy iterator |
| `parallel.OrderedParallelMap` | Transforms each element in parallel across N workers while preserving output order |

---

## Files

| File | Description |
|---|---|
| `main.go` | Full ETL pipeline: seed, define, execute, summarize |
| `users.csv` | Output CSV (generated at runtime, deleted after summary) |