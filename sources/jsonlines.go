package sources

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"iter"
	"os"
)

// JSONLines returns a lazy sequence of decoded JSON objects from any io.Reader.
// Each line must be a valid JSON object — empty lines are skipped.
// Stops immediately if ctx is cancelled or the consumer breaks early.
//
// Works with any io.Reader — files, HTTP responses, buffers:
//
//	// from file
//	file, _ := os.Open("data.jsonl")
//	defer file.Close()
//	for item, err := range sources.JSONLines[User](ctx, file) {
//	    if err != nil { log.Fatal(err) }
//	    fmt.Println(item)
//	}
//
//	// from HTTP response
//	resp, _ := http.DefaultClient.Do(req)
//	defer resp.Body.Close()
//	for item, err := range sources.JSONLines[User](ctx, resp.Body) {
//	    if err != nil { log.Fatal(err) }
//	    fmt.Println(item)
//	}
func JSONLines[T any](ctx context.Context, r io.Reader) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T

		if ctx.Err() != nil {
			yield(zero, ctx.Err())
			return
		}

		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for scanner.Scan() {
			if ctx.Err() != nil {
				yield(zero, ctx.Err())
				return
			}

			line := scanner.Bytes()

			// skip empty lines
			if len(line) == 0 {
				continue
			}

			var item T
			if err := json.Unmarshal(line, &item); err != nil {
				if !yield(zero, err) {
					return
				}
				continue
			}

			if !yield(item, nil) {
				return
			}
		}

		// surface scanner errors — network drop, oversized line, etc
		if err := scanner.Err(); err != nil {
			if ctx.Err() != nil {
				return
			}
			yield(zero, err)
		}
	}
}

// JSONLinesFile opens a file and returns a lazy sequence of decoded JSON objects.
// Each line must be a valid JSON object — empty lines are skipped.
//
// example:
//
//	for item, err := range sources.JSONLinesFile[User](ctx, "users.jsonl") {
//	    if err != nil { log.Fatal(err) }
//	    fmt.Println(item)
//	}
func JSONLinesFile[T any](ctx context.Context, path string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T

		if ctx.Err() != nil {
			yield(zero, ctx.Err())
			return
		}

		file, err := os.Open(path)
		if err != nil {
			yield(zero, err)
			return
		}
		defer file.Close()

		for item, err := range JSONLines[T](ctx, file) {
			if !yield(item, err) {
				return
			}
		}
	}
}