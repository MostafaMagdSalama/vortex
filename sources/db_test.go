package sources_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/MostafaMagdSalama/vortex/iterx"
	"github.com/MostafaMagdSalama/vortex/sources"
	_ "modernc.org/sqlite"
)

type User struct {
	ID     int
	Name   string
	Email  string
	Status string
}

func setupDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE users (
		id     INTEGER PRIMARY KEY,
		name   TEXT,
		email  TEXT,
		status TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}

	users := []User{
		{1, "Alice", "alice@example.com", "active"},
		{2, "Bob", "bob@example.com", "inactive"},
		{3, "Charlie", "charlie@example.com", "active"},
		{4, "Diana", "diana@example.com", "active"},
		{5, "Eve", "eve@example.com", "inactive"},
	}

	for _, u := range users {
		_, err = db.Exec(
			"INSERT INTO users VALUES (?, ?, ?, ?)",
			u.ID, u.Name, u.Email, u.Status,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	return db
}

func scanUser(rows *sql.Rows) (User, error) {
	var u User
	err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Status)
	return u, err
}

// reads all rows — no error expected
func TestDBRows_All(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	var result []User
	for u, err := range sources.DBRows(context.Background(), db,
		"SELECT id, name, email, status FROM users",
		scanUser,
	) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, u)
	}

	if len(result) != 5 {
		t.Fatalf("expected 5 users, got %d", len(result))
	}
}

// correct data is returned
func TestDBRows_CorrectData(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	var result []User
	for u, err := range sources.DBRows(context.Background(), db,
		"SELECT id, name, email, status FROM users ORDER BY id",
		scanUser,
	) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, u)
	}

	if result[0].Name != "Alice" {
		t.Fatalf("expected Alice, got %s", result[0].Name)
	}
	if result[0].Email != "alice@example.com" {
		t.Fatalf("expected alice@example.com, got %s", result[0].Email)
	}
}

// filter active users using iterx.Filter
func TestDBRows_Filter(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	// collect without errors first
	users := func(yield func(User) bool) {
		for u, err := range sources.DBRows(context.Background(), db,
			"SELECT id, name, email, status FROM users",
			scanUser,
		) {
			if err != nil {
				return
			}
			if !yield(u) {
				return
			}
		}
	}

	var active []User
	for u := range iterx.Filter(context.Background(), users, func(u User) bool {
		return u.Status == "active"
	}) {
		active = append(active, u)
	}

	if len(active) != 3 {
		t.Fatalf("expected 3 active users, got %d", len(active))
	}
}

// stops early — cursor closes cleanly
func TestDBRows_EarlyStop(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	count := 0
	for _, err := range sources.DBRows(context.Background(), db,
		"SELECT id, name, email, status FROM users",
		scanUser,
	) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		count++
		if count == 2 {
			break
		}
	}

	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

// query with args
func TestDBRows_WithArgs(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	var result []User
	for u, err := range sources.DBRows(context.Background(), db,
		"SELECT id, name, email, status FROM users WHERE status = ?",
		scanUser,
		"active",
	) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, u)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 active users, got %d", len(result))
	}
}

// query with multiple args
func TestDBRows_WithMultipleArgs(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	var result []User
	for u, err := range sources.DBRows(context.Background(), db,
		"SELECT id, name, email, status FROM users WHERE status = ? AND id > ?",
		scanUser,
		"active", 1,
	) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, u)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 users, got %d", len(result))
	}
}

// works with *sql.Tx
func TestDBRows_Transaction(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	var result []User
	for u, err := range sources.DBRows(context.Background(), tx,
		"SELECT id, name, email, status FROM users",
		scanUser,
	) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, u)
	}

	if len(result) != 5 {
		t.Fatalf("expected 5 users, got %d", len(result))
	}
}

// empty table — should return no rows no error
func TestDBRows_EmptyTable(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	db.Exec("DELETE FROM users")

	var result []User
	for u, err := range sources.DBRows(context.Background(), db,
		"SELECT id, name, email, status FROM users",
		scanUser,
	) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, u)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 users, got %d", len(result))
	}
}

// bad query — should surface error not panic
func TestDBRows_BadQuery(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	var gotErr error
	for _, err := range sources.DBRows(context.Background(), db,
		"SELECT * FROM nonexistent_table",
		scanUser,
	) {
		if err != nil {
			gotErr = err
			break
		}
	}

	if gotErr == nil {
		t.Fatal("expected error for bad query, got nil")
	}
}

// cancelled context — should surface context error
func TestDBRows_Cancelled(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var gotErr error
	for _, err := range sources.DBRows(ctx, db,
		"SELECT id, name, email, status FROM users",
		scanUser,
	) {
		if err != nil {
			gotErr = err
			break
		}
	}

	if !errors.Is(gotErr, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", gotErr)
	}
}

// read and write in same transaction
func TestDBRows_TransactionReadWrite(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	for u, err := range sources.DBRows(context.Background(), tx,
		"SELECT id, name, email, status FROM users WHERE status = 'active'",
		scanUser,
	) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = tx.Exec(
			"UPDATE users SET name = ? WHERE id = ?",
			u.Name+"_updated", u.ID,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	tx.Commit()

	var result []User
	for u, err := range sources.DBRows(context.Background(), db,
		"SELECT id, name, email, status FROM users WHERE status = 'active'",
		scanUser,
	) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, u)
	}

	for _, u := range result {
		if u.Name[len(u.Name)-8:] != "_updated" {
			t.Fatalf("expected name to end with _updated, got %s", u.Name)
		}
	}
}
