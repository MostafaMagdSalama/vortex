package iterx

import (
	"context"
	"iter"
)

// Drain consumes a sequence and calls fn for each item.
// Stops immediately if ctx is cancelled or fn returns an error.
// Use Drain when your terminal operation can fail — writing to CSV, DB, files.
// Use ForEach when your terminal operation cannot fail — logging, printing.
//
// example:
//
//	err := interx.Drain(ctx, users, func(u User) error {
//	    return csvWriter.Write([]string{u.Name, u.Email})
//	})
func Drain[T any](ctx context.Context, seq iter.Seq[T], fn func(T) error) error {
	for v := range seq {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := fn(v); err != nil {
			return err
		}
	}
	return nil
}
