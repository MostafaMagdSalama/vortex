package sources

import (
	"context"
	"encoding/csv"
	"io"
	"iter"

	"github.com/MostafaMagdSalama/vortex"
)

// CSVRows returns a lazy sequence of rows from a CSV reader.
func CSVRows(ctx context.Context, r io.Reader) iter.Seq2[[]string, error] {
	return func(yield func([]string, error) bool) {
		if ctx.Err() != nil {
			yield(nil, vortex.Wrap("sources.CSVRows", ctx.Err()))
			return
		}

		reader := csv.NewReader(r)
		// FieldsPerRecord = 0 means the field count is set by the first record.
		// Subsequent records with a different number of fields will return an error.
		reader.FieldsPerRecord = 0

		for {
			if ctx.Err() != nil {
				yield(nil, vortex.Wrap("sources.CSVRows", ctx.Err()))
				return
			}

			row, err := reader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				if !yield(nil, vortex.Wrap("sources.CSVRows", err)) {
					return
				}
				continue
			}

			if !yield(row, nil) {
				return
			}
		}
	}
}