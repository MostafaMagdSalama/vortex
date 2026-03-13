package sources

import (
	"context"
	"encoding/csv"
	"io"
	"iter"
)

// CSVRows returns a lazy sequence of rows from a CSV reader.
func CSVRows(ctx context.Context, r io.Reader) iter.Seq[[]string] {
	return func(yield func([]string) bool) {
		reader := csv.NewReader(r)
		for {
			if ctx.Err() != nil {
				return
			}
			row, err := reader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				return
			}
			if !yield(row) {
				return
			}
		}
	}
}

// CSVRowsWithError is like CSVRows but surfaces read errors to the caller.
func CSVRowsWithError(ctx context.Context, r io.Reader) iter.Seq2[[]string, error] {
	return func(yield func([]string, error) bool) {
		reader := csv.NewReader(r)
		for {
			if ctx.Err() != nil {
				return
			}
			row, err := reader.Read()
			if err == io.EOF {
				return
			}
			if !yield(row, err) {
				return
			}
		}
	}
}
