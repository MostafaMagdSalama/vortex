package sources

import (
	"database/sql"
	"iter"
)

// querier is anything that can run a SQL query.
// both *sql.DB and *sql.Tx satisfy this interface.
type querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// DBRows returns a lazy sequence of scanned rows from a SQL query.
// Works with both *sql.DB and *sql.Tx.
//
// example:
//
//	users := sources.DBRows(db, "SELECT id, name FROM users",
//	    func(rows *sql.Rows) (User, error) {
//	        var u User
//	        return u, rows.Scan(&u.ID, &u.Name)
//	    },
//	)
//
//	for user := range users {
//	    fmt.Println(user)
//	}
func DBRows[T any](db querier, query string, scan func(*sql.Rows) (T, error)) iter.Seq[T] {
	return func(yield func(T) bool) {
		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
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
// Works with both *sql.DB and *sql.Tx.
//
// example:
//
//	users := sources.DBRowsWithArgs(db,
//	    "SELECT id, name FROM users WHERE status = ?",
//	    []any{"active"},
//	    func(rows *sql.Rows) (User, error) {
//	        var u User
//	        return u, rows.Scan(&u.ID, &u.Name)
//	    },
//	)
//
//	for user := range users {
//	    fmt.Println(user)
//	}
func DBRowsWithArgs[T any](db querier, query string, args []any, scan func(*sql.Rows) (T, error)) iter.Seq[T] {
	return func(yield func(T) bool) {
		rows, err := db.Query(query, args...)
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
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