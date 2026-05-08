package sources_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/MostafaMagdSalama/vortex"
	"github.com/MostafaMagdSalama/vortex/sources"
)

// ── Minimal fake SQL driver ──────────────────────────────────────────────────
//
// Registers "fakedb" once and lets each test configure independent rows,
// query errors, and rows.Err() values without any external dependency.

var (
	fakeDBRegistry sync.Map
	fakeDBCounter  atomic.Int64
)

func init() {
	sql.Register("fakedb", &fakeDriver{})
}

type fakeConfig struct {
	cols     []string
	rows     [][]driver.Value
	queryErr error // returned by QueryContext
	rowsErr  error // surfaced via rows.Err() after all rows are exhausted
}

func openFakeDB(t *testing.T, cfg *fakeConfig) *sql.DB {
	t.Helper()
	key := fmt.Sprintf("fakedb-%d", fakeDBCounter.Add(1))
	fakeDBRegistry.Store(key, cfg)
	db, err := sql.Open("fakedb", key)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		fakeDBRegistry.Delete(key)
	})
	return db
}

type fakeDriver struct{}

func (d *fakeDriver) Open(name string) (driver.Conn, error) {
	v, ok := fakeDBRegistry.Load(name)
	if !ok {
		return nil, fmt.Errorf("fakedb: unknown key %q", name)
	}
	return &fakeConn{cfg: v.(*fakeConfig)}, nil
}

type fakeConn struct{ cfg *fakeConfig }

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeStmt{cfg: c.cfg}, nil
}
func (c *fakeConn) Close() error              { return nil }
func (c *fakeConn) Begin() (driver.Tx, error) { return &fakeTx{}, nil }

type fakeTx struct{}

func (t *fakeTx) Commit() error   { return nil }
func (t *fakeTx) Rollback() error { return nil }

type fakeStmt struct{ cfg *fakeConfig }

func (s *fakeStmt) Close() error  { return nil }
func (s *fakeStmt) NumInput() int { return -1 }
func (s *fakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, errors.New("fakedb: exec not supported")
}
func (s *fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	if s.cfg.queryErr != nil {
		return nil, s.cfg.queryErr
	}
	return &fakeRows{cfg: s.cfg}, nil
}

type fakeRows struct {
	cfg *fakeConfig
	idx int
}

func (r *fakeRows) Columns() []string { return r.cfg.cols }
func (r *fakeRows) Close() error      { return nil }
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.cfg.rows) {
		if r.cfg.rowsErr != nil {
			err := r.cfg.rowsErr
			r.cfg.rowsErr = nil
			return err
		}
		return io.EOF
	}
	copy(dest, r.cfg.rows[r.idx])
	r.idx++
	return nil
}

// ── Tests ────────────────────────────────────────────────────────────────────

func scanInt64(rows *sql.Rows) (int64, error) {
	var n int64
	return n, rows.Scan(&n)
}

func TestDBRows_Basic(t *testing.T) {
	db := openFakeDB(t, &fakeConfig{
		cols: []string{"n"},
		rows: [][]driver.Value{{int64(1)}, {int64(2)}, {int64(3)}},
	})

	var got []int64
	for v, err := range sources.DBRows(context.Background(), db, "SELECT n FROM t", scanInt64) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}

	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("expected [1 2 3], got %v", got)
	}
}

func TestDBRows_Empty(t *testing.T) {
	db := openFakeDB(t, &fakeConfig{
		cols: []string{"n"},
		rows: nil,
	})

	var count int
	for _, err := range sources.DBRows(context.Background(), db, "SELECT n FROM t", scanInt64) {
		if err != nil {
			t.Fatal(err)
		}
		count++
	}

	if count != 0 {
		t.Fatalf("expected 0 rows, got %d", count)
	}
}

func TestDBRows_WithArgs(t *testing.T) {
	db := openFakeDB(t, &fakeConfig{
		cols: []string{"n"},
		rows: [][]driver.Value{{int64(42)}},
	})

	var got []int64
	for v, err := range sources.DBRows(context.Background(), db, "SELECT n FROM t WHERE n = ?", scanInt64, int64(42)) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}

	if len(got) != 1 || got[0] != 42 {
		t.Fatalf("expected [42], got %v", got)
	}
}

func TestDBRows_QueryError(t *testing.T) {
	queryErr := errors.New("connection refused")
	db := openFakeDB(t, &fakeConfig{
		cols:     []string{"n"},
		queryErr: queryErr,
	})

	var gotErr error
	for _, err := range sources.DBRows(context.Background(), db, "SELECT n FROM t", scanInt64) {
		gotErr = err
	}

	if !errors.Is(gotErr, queryErr) {
		t.Fatalf("expected query error, got %v", gotErr)
	}
}

func TestDBRows_ScanError(t *testing.T) {
	scanErr := errors.New("scan failed")
	db := openFakeDB(t, &fakeConfig{
		cols: []string{"n"},
		rows: [][]driver.Value{{int64(1)}, {int64(2)}, {int64(3)}},
	})

	call := 0
	scan := func(rows *sql.Rows) (int64, error) {
		var n int64
		if err := rows.Scan(&n); err != nil {
			return 0, err
		}
		call++
		if call == 2 {
			return 0, scanErr
		}
		return n, nil
	}

	var got []int64
	var gotErr error
	for v, err := range sources.DBRows(context.Background(), db, "SELECT n FROM t", scan) {
		if err != nil {
			gotErr = err
			break
		}
		got = append(got, v)
	}

	if !errors.Is(gotErr, scanErr) {
		t.Fatalf("expected scan error, got %v", gotErr)
	}
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("expected [1] before scan error, got %v", got)
	}
}

func TestDBRows_CancelledBeforeQuery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	db := openFakeDB(t, &fakeConfig{
		cols: []string{"n"},
		rows: [][]driver.Value{{int64(1)}, {int64(2)}},
	})

	var gotErr error
	for _, err := range sources.DBRows(ctx, db, "SELECT n FROM t", scanInt64) {
		gotErr = err
	}

	if !errors.Is(gotErr, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled before query, got %v", gotErr)
	}
}

func TestDBRows_CancelledMidIteration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	db := openFakeDB(t, &fakeConfig{
		cols: []string{"n"},
		rows: [][]driver.Value{{int64(1)}, {int64(2)}, {int64(3)}},
	})

	call := 0
	scan := func(rows *sql.Rows) (int64, error) {
		var n int64
		err := rows.Scan(&n)
		call++
		if call == 1 {
			cancel() // cancel after scanning first row
		}
		return n, err
	}

	var got []int64
	var gotErr error
	for v, err := range sources.DBRows(ctx, db, "SELECT n FROM t", scan) {
		if err != nil {
			gotErr = err
			break
		}
		got = append(got, v)
	}

	if !errors.Is(gotErr, vortex.ErrCancelled) {
		t.Fatalf("expected ErrCancelled mid-iteration, got %v", gotErr)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 row before cancellation, got %v", got)
	}
}

func TestDBRows_RowsErr(t *testing.T) {
	driverErr := errors.New("driver network error")
	db := openFakeDB(t, &fakeConfig{
		cols:    []string{"n"},
		rows:    [][]driver.Value{{int64(1)}, {int64(2)}},
		rowsErr: driverErr,
	})

	var got []int64
	var gotErr error
	for v, err := range sources.DBRows(context.Background(), db, "SELECT n FROM t", scanInt64) {
		if err != nil {
			gotErr = err
			break
		}
		got = append(got, v)
	}

	if !errors.Is(gotErr, driverErr) {
		t.Fatalf("expected driver error from rows.Err(), got %v", gotErr)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 rows before error, got %v", got)
	}
}

func TestDBRows_EarlyStop(t *testing.T) {
	db := openFakeDB(t, &fakeConfig{
		cols: []string{"n"},
		rows: [][]driver.Value{{int64(1)}, {int64(2)}, {int64(3)}, {int64(4)}, {int64(5)}},
	})

	var got []int64
	for v, err := range sources.DBRows(context.Background(), db, "SELECT n FROM t", scanInt64) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}

	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("expected [1 2], got %v", got)
	}
}
