package sources

import (
	"context"
	"database/sql"
	"iter"
)

type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// DBRows returns a lazy sequence of scanned rows from a SQL query.
func DBRows[T any](ctx context.Context, db querier, query string, scan func(*sql.Rows) (T, error)) iter.Seq[T] {
	return func(yield func(T) bool) {
		if ctx.Err() != nil {
			return
		}

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			if ctx.Err() != nil {
				return
			}
			val, err := scan(rows)
			if err != nil {
				return
			}
			if !yield(val) {
				return
			}
		}
	}
}

// DBRowsWithArgs is like DBRows but accepts query arguments.
func DBRowsWithArgs[T any](ctx context.Context, db querier, query string, args []any, scan func(*sql.Rows) (T, error)) iter.Seq[T] {
	return func(yield func(T) bool) {
		if ctx.Err() != nil {
			return
		}

		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			if ctx.Err() != nil {
				return
			}
			val, err := scan(rows)
			if err != nil {
				return
			}
			if !yield(val) {
				return
			}
		}
	}
}
