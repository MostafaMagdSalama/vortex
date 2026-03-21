# Example: Database → Excel Export (1,000,000 rows)

This example benchmarks two approaches to exporting a large SQLite table to an Excel file using [vortex](https://github.com/MostafaMagdSalama/vortex).

It seeds a `products` table with **1,000,000 rows**, applies a filter (`stock > 0`), and writes the results to `.xlsx` — once lazily with vortex, once eagerly by loading everything into RAM.

---

## What it demonstrates

- **Lazy pipeline with vortex** — rows are streamed one at a time from the database directly into the Excel file via `excelize.StreamWriter`. Memory usage stays flat regardless of table size.
- **Eager approach** — all rows are loaded into a slice, filtered into a second slice, then written to Excel. This keeps up to three copies of the data in RAM at the same time.

---

## How to run

```bash
go run main.go
```

This will:
1. Seed `bench.db` with 1,000,000 rows (skipped if already seeded)
2. Run the vortex (lazy) approach → writes `vortex_output.xlsx`
3. Run the eager approach → writes `eager_output.xlsx`
4. Print memory and timing stats for both

---

## Benchmark

> Measured on Linux. Numbers may vary by machine.
> 
> ⚠️ The lazy approach memory figure could be even lower — the current Excel streaming library (`excelize`) uses an internal write buffer, so actual row-level memory pressure from vortex is smaller than the reported delta suggests.

### Results

| Approach | Peak memory | Time | Notes |
|---|---|---|---|
| Eager (load all) | 440 MB | ~12.6s | loads all rows into RAM, then filters, then writes |
| Lazy (vortex) | 77 MB | ~3.7s | streams one row at a time directly into the Excel file |

**5.7× less memory** and **3.4× faster** with lazy processing.

<details>
<summary><b>View detailed benchmark output</b></summary>
<br>

```text
╔══════════════════════════════════════════╗
║         with vortex (lazy)               ║
╚══════════════════════════════════════════╝
rows written : 996667
time         : 3.7112285s
mem before   : 0.73 MB
mem after    : 77.65 MB
mem delta    : +76.92 MB

╔══════════════════════════════════════════╗
║         without vortex (eager)           ║
╚══════════════════════════════════════════╝
mem after loading all rows : 213.58 MB
mem after filtering        : 315.11 MB
rows written : 996667
time         : 12.595819s
mem before   : 53.61 MB
mem after    : 493.71 MB
mem delta    : +440.10 MB
```

</details>

### Why the eager approach uses so much memory

The eager approach accumulates **three copies of the data in RAM simultaneously**:

1. The raw slice of all scanned rows
2. The filtered slice
3. The Excel in-memory XML tree built by `excelize`

The lazy approach holds **one row at a time** in memory, regardless of how large the table grows.

---

## Key vortex APIs used

| API | Purpose |
|---|---|
| `sources.DBRows` | Turns a `*sql.Rows` query into a lazy iterator |
| `iterx.Filter` | Filters elements from a sequence without materializing it |

---

## Files

| File | Description |
|---|---|
| `main.go` | Full benchmark: seed, lazy run, eager run |
| `vortex_output.xlsx` | Output from the lazy approach (generated at runtime) |
| `eager_output.xlsx` | Output from the eager approach (generated at runtime) |
| `bench.db` | SQLite database (generated at runtime) |