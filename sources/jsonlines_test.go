package sources_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MostafaMagdSalama/vortex/sources"
)

type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Service string `json:"service"`
}

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price float64 `json:"price"`
}

func TestJSONLines_Basic(t *testing.T) {
	input := strings.NewReader(`{"level":"info","message":"started","service":"api"}
{"level":"error","message":"failed","service":"api"}
{"level":"info","message":"done","service":"api"}
`)

	var result []LogEntry
	for item, err := range sources.JSONLines[LogEntry](context.Background(), input) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, item)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}
	if result[0].Level != "info" {
		t.Fatalf("expected info, got %s", result[0].Level)
	}
	if result[1].Level != "error" {
		t.Fatalf("expected error, got %s", result[1].Level)
	}
}

// correct data is decoded
func TestJSONLines_CorrectData(t *testing.T) {
	input := strings.NewReader(`{"id":1,"name":"Alice","price":9.99}
{"id":2,"name":"Bob","price":19.99}
`)

	var result []Product
	for item, err := range sources.JSONLines[Product](context.Background(), input) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, item)
	}

	if result[0].ID != 1 || result[0].Name != "Alice" || result[0].Price != 9.99 {
		t.Fatalf("unexpected first item: %+v", result[0])
	}
	if result[1].ID != 2 || result[1].Name != "Bob" {
		t.Fatalf("unexpected second item: %+v", result[1])
	}
}

// empty lines are skipped
func TestJSONLines_SkipsEmptyLines(t *testing.T) {
	input := strings.NewReader(`{"level":"info","message":"a","service":"x"}

{"level":"error","message":"b","service":"x"}

{"level":"info","message":"c","service":"x"}
`)

	var result []LogEntry
	for item, err := range sources.JSONLines[LogEntry](context.Background(), input) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, item)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}
}

// invalid JSON line surfaces error and continues
func TestJSONLines_InvalidJSON(t *testing.T) {
	input := strings.NewReader(`{"level":"info","message":"ok","service":"x"}
not valid json
{"level":"info","message":"also ok","service":"x"}
`)

	var items []LogEntry
	var errs []error

	for item, err := range sources.JSONLines[LogEntry](context.Background(), input) {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		items = append(items, item)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 valid items, got %d", len(items))
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

// empty reader — no items no error
func TestJSONLines_Empty(t *testing.T) {
	input := strings.NewReader("")

	var result []LogEntry
	for item, err := range sources.JSONLines[LogEntry](context.Background(), input) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, item)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 items, got %d", len(result))
	}
}

// early stop — stops cleanly
func TestJSONLines_EarlyStop(t *testing.T) {
	input := strings.NewReader(`{"level":"info","message":"a","service":"x"}
{"level":"info","message":"b","service":"x"}
{"level":"info","message":"c","service":"x"}
{"level":"info","message":"d","service":"x"}
{"level":"info","message":"e","service":"x"}
`)

	count := 0
	for _, err := range sources.JSONLines[LogEntry](context.Background(), input) {
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

// cancelled context — stops immediately
func TestJSONLines_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	input := strings.NewReader(`{"level":"info","message":"a","service":"x"}
{"level":"info","message":"b","service":"x"}
`)

	var gotErr error
	for _, err := range sources.JSONLines[LogEntry](ctx, input) {
		if err != nil {
			gotErr = err
			break
		}
	}

	if !errors.Is(gotErr, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", gotErr)
	}
}

// JSONLinesFile — reads from file correctly
func TestJSONLinesFile_Basic(t *testing.T) {
	content := `{"level":"info","message":"started","service":"api"}
{"level":"error","message":"failed","service":"api"}
`
	tmp := filepath.Join(t.TempDir(), "test.jsonl")
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var result []LogEntry
	for item, err := range sources.JSONLinesFile[LogEntry](context.Background(), tmp) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result = append(result, item)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}
}

// JSONLinesFile — missing file surfaces error
func TestJSONLinesFile_MissingFile(t *testing.T) {
	var gotErr error
	for _, err := range sources.JSONLinesFile[LogEntry](context.Background(), "nonexistent.jsonl") {
		if err != nil {
			gotErr = err
			break
		}
	}

	if gotErr == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// Example functions for pkg.go.dev
func ExampleJSONLines() {
	input := strings.NewReader(`{"level":"info","message":"started","service":"api"}
{"level":"error","message":"failed","service":"api"}
`)

	type Entry struct {
		Level   string `json:"level"`
		Message string `json:"message"`
		Service string `json:"service"`
	}

	for item, err := range sources.JSONLines[Entry](context.Background(), input) {
		if err != nil {
			break
		}
		fmt.Printf("%s: %s\n", item.Level, item.Message)
	}
	// Output:
	// info: started
	// error: failed
}

func ExampleJSONLinesFile() {
	// write a temp file for the example
	tmp, _ := os.CreateTemp("", "*.jsonl")
	defer os.Remove(tmp.Name())
	tmp.WriteString(`{"id":1,"name":"Alice","price":9.99}` + "\n")
	tmp.WriteString(`{"id":2,"name":"Bob","price":19.99}` + "\n")
	tmp.Close()

	type Product struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}

	for item, err := range sources.JSONLinesFile[Product](context.Background(), tmp.Name()) {
		if err != nil {
			break
		}
		fmt.Printf("%d: %s $%.2f\n", item.ID, item.Name, item.Price)
	}
	// Output:
	// 1: Alice $9.99
	// 2: Bob $19.99
}