package sources

import (
	"database/sql"
	"iter"
)

// DBRows returns a lazy sequence of scanned rows from a SQL query.
// The scan function controls how each row is converted to type T.
// Closes the cursor automatically when the sequence is exhausted
// or the caller stops early.
//
// example:
//
//	type User struct { ID int; Name string }
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
func DBRows[T any](db *sql.DB, query string, scan func(*sql.Rows) (T, error)) iter.Seq[T] {
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
				return // caller stopped — cursor closes via defer
			}
		}
	}
}

// DBRowsWithArgs is like DBRows but accepts query arguments.
//
// example:
//
//	users := sources.DBRowsWithArgs(db,
//	    "SELECT id, name FROM users WHERE active = ?",
//	    []any{true},
//	    func(rows *sql.Rows) (User, error) { ... },
//	)
func DBRowsWithArgs[T any](db *sql.DB, query string, args []any, scan func(*sql.Rows) (T, error)) iter.Seq[T] {
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