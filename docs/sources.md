# Sources and Lazy Generators

Vortex treats all incoming data — a 100GB CSV, a 10-row database query, or lines streamed from an HTTP endpoint — as generic `iter.Seq2`.

This drastically reduces the API surface of Vortex. Every source yields identically, so they are entirely interchangeable.

## `io.Reader` sources

Functions like `Lines` and `CSVRows` accept any `io.Reader`. This allows extreme flexibility:

```go
file, _ := os.Open("huge_file.csv")                 // A file
resp, _ := http.DefaultClient.Do(req)               // An HTTP stream
formFile, _, _ := r.FormFile("csv")                 // A multipart form upload
buf := bytes.NewBufferString("a,b\n1,2")            // A string buffer 
```

Because `vortex/sources` wraps these into an iterator, the downstream `vortex/iterx` package (like `Map` or `Take`) has no idea where the data originated.

## Memory Implications

When using an `io.Reader` source like `JSONLines` or `CSVRows`, the peak memory consumption is precisely bounded to the size of **a single line or record**, not the size of the whole payload. You could process a terabyte of CSV data on a 50 MB RAM container safely using Vortex.

## Graceful Cancellation

If the `context.Context` is cancelled during iteration:
1. `sources` will stop querying `reader.Read()` or `rows.Next()`.
2. The sequence terminates immediately.
3. The underlying operation bubbles up a `vortex.Error` with `vortex.ErrCancelled`.

This allows immediate resource de-allocation instead of hanging until an eager list is fully built.
