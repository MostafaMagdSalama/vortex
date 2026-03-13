package sources

import (
	"encoding/csv"
	"io"
	"iter"
)

// CSVRows returns a lazy sequence of rows from a CSV reader.
// Reads one row at a time — never loads the whole file into memory.
//
// example:
//
//	file, _ := os.Open("huge.csv")
//	defer file.Close()
//
//	for row := range sources.CSVRows(file) {
//	    fmt.Println(row)
//	}
func CSVRows(r io.Reader) iter.Seq[[]string] {
	return func(yield func([]string) bool) {
		reader := csv.NewReader(r)
		for {
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
// Use this when you need to handle malformed CSV gracefully.
func CSVRowsWithError(r io.Reader) iter.Seq2[[]string, error] {
	return func(yield func([]string, error) bool) {
		reader := csv.NewReader(r)
		for {
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