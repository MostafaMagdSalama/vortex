package sources_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
	"testing"

	"github.com/MostafaMagdSalama/vortex/sources"
)

type stubDriver struct{}

type stubConn struct{}

type stubRows struct {
	columns []string
	data    [][]driver.Value
	index   int
}

var (
	registerStubDriver sync.Once
	queryCount         int
	queryArgsCount     int
)

func (stubDriver) Open(name string) (driver.Conn, error) {
	return stubConn{}, nil
}

func (stubConn) Prepare(query string) (driver.Stmt, error) { return nil, nil }
func (stubConn) Close() error                              { return nil }
func (stubConn) Begin() (driver.Tx, error)                 { return nil, nil }

func (stubConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	switch query {
	case "SELECT name FROM users":
		queryCount++
		return &stubRows{
			columns: []string{"name"},
			data: [][]driver.Value{
				{"alice"},
				{"bob"},
			},
		}, nil
	case "SELECT value FROM numbers WHERE kind = ?":
		queryArgsCount++
		return &stubRows{
			columns: []string{"value"},
			data: [][]driver.Value{
				{1},
				{2},
				{3},
			},
		}, nil
	default:
		return &stubRows{columns: []string{"value"}}, nil
	}
}

func (r *stubRows) Columns() []string {
	return r.columns
}

func (r *stubRows) Close() error {
	return nil
}

func (r *stubRows) Next(dest []driver.Value) error {
	if r.index >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.index])
	r.index++
	return nil
}

func openStubDB(t *testing.T) *sql.DB {
	t.Helper()
	registerStubDriver.Do(func() {
		sql.Register("vortex_stub", stubDriver{})
	})

	db, err := sql.Open("vortex_stub", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestDBRows(t *testing.T) {
	queryCount = 0
	db := openStubDB(t)

	var names []string
	for name := range sources.DBRows(context.Background(), db, "SELECT name FROM users", func(rows *sql.Rows) (string, error) {
		var name string
		return name, rows.Scan(&name)
	}) {
		names = append(names, name)
	}

	if len(names) != 2 || names[0] != "alice" || queryCount != 1 {
		t.Fatalf("names=%v queryCount=%d", names, queryCount)
	}
}

func TestDBRows_Cancelled(t *testing.T) {
	queryCount = 0
	db := openStubDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var names []string
	for name := range sources.DBRows(ctx, db, "SELECT name FROM users", func(rows *sql.Rows) (string, error) {
		var name string
		return name, rows.Scan(&name)
	}) {
		names = append(names, name)
	}

	if len(names) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(names))
	}
	if queryCount != 0 {
		t.Fatalf("expected query not to run on cancelled context, got %d calls", queryCount)
	}
}

func TestDBRowsWithArgs(t *testing.T) {
	queryArgsCount = 0
	db := openStubDB(t)

	var values []int
	for value := range sources.DBRowsWithArgs(context.Background(), db, "SELECT value FROM numbers WHERE kind = ?", []any{"even"}, func(rows *sql.Rows) (int, error) {
		var value int
		return value, rows.Scan(&value)
	}) {
		values = append(values, value)
	}

	if len(values) != 3 || values[0] != 1 || queryArgsCount != 1 {
		t.Fatalf("values=%v queryArgsCount=%d", values, queryArgsCount)
	}
}

func TestDBRowsWithArgs_Cancelled(t *testing.T) {
	queryArgsCount = 0
	db := openStubDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var values []int
	for value := range sources.DBRowsWithArgs(ctx, db, "SELECT value FROM numbers WHERE kind = ?", []any{"even"}, func(rows *sql.Rows) (int, error) {
		var value int
		return value, rows.Scan(&value)
	}) {
		values = append(values, value)
	}

	if len(values) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(values))
	}
	if queryArgsCount != 0 {
		t.Fatalf("expected query not to run on cancelled context, got %d calls", queryArgsCount)
	}
}
