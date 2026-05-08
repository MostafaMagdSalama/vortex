package sources

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"iter"

	"github.com/MostafaMagdSalama/vortex"
)

// CSVRows returns a lazy sequence of rows from a CSV reader.
//
// Recoverable parse errors (wrong field count, bad quoting) are yielded inline
// and iteration continues. Non-recoverable I/O errors are yielded once and
// iteration stops.
//
// Cancellation: ctx is checked between rows. A blocked read on the underlying
// io.Reader (e.g. a slow HTTP body) will not be interrupted by ctx alone —
// close the reader to unblock it. For HTTP, cancel the request context or call
// resp.Body.Close from another goroutine.
func CSVRows(ctx context.Context, r io.Reader) iter.Seq2[[]string, error] {
	return func(yield func([]string, error) bool) {
		if ctx.Err() != nil {
			yield(nil, vortex.WrapCancelled("sources.CSVRows"))
			return
		}

		reader := csv.NewReader(r)
		// FieldsPerRecord = 0 means the field count is set by the first record.
		// Subsequent records with a different number of fields will return an error.
		reader.FieldsPerRecord = 0

		for {
			if ctx.Err() != nil {
				yield(nil, vortex.WrapCancelled("sources.CSVRows"))
				return
			}

			row, err := reader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				// csv.ParseError is per-record and recoverable; the next Read
				// continues at the next record. Any other error (I/O failure,
				// closed reader) is terminal — csv.Reader will keep returning
				// it, so surface it once and stop.
				var perr *csv.ParseError
				if errors.As(err, &perr) {
					if !yield(nil, vortex.Wrap("sources.CSVRows", err)) {
						return
					}
					continue
				}
				yield(nil, vortex.Wrap("sources.CSVRows", err))
				return
			}

			if !yield(row, nil) {
				return
			}
		}
	}
}