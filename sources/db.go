package sources

import (
	"context"
	"database/sql"
	"iter"
)

// querier is anything that can run a SQL query.
// both *sql.DB and *sql.Tx satisfy this interface.
type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// DBRows returns a lazy sequence of scanned rows from a SQL query.
// Accepts optional query arguments via variadic args.
// Always check the error — a non-nil error means iteration stopped early
// due to a scan failure, driver error, or context cancellation.
//
// example without args:
//
//	for u, err := range sources.DBRows(ctx, db, "SELECT * FROM users", scan) {
//	    if err != nil {
//	        log.Println(err)
//	        return
//	    }
//	    process(u)
//	}
//
// example with args:
//
//	for u, err := range sources.DBRows(ctx, db,
//	    "SELECT * FROM users WHERE status = ?",
//	    scan,
//	    "active",
//	) {
//	    if err != nil {
//	        log.Println(err)
//	        return
//	    }
//	    process(u)
//	}
func DBRows[T any](ctx context.Context, db querier, query string, scan func(*sql.Rows) (T, error), args ...any) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T

		// check context before opening connection
		if ctx.Err() != nil {
			yield(zero, ctx.Err())
			return
		}

		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			yield(zero, err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			// check context before processing each row
			if ctx.Err() != nil {
				yield(zero, ctx.Err())
				return
			}

			val, err := scan(rows)
			if err != nil {
				// scan failed — surface the error and stop
				yield(zero, err)
				return
			}

			if !yield(val, nil) {
				// caller broke out of the loop — stop cleanly
				return
			}
		}

		// check for driver-side iteration errors
		// rows.Next() returns false on both clean completion AND error
		// without this check a dropped connection looks like clean completion
		if err := rows.Err(); err != nil {
			yield(zero, err)
		}
	}
}