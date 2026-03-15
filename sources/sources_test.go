package sources_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/MostafaMagdSalama/vortex/sources"
)

func TestCSVRows(t *testing.T) {
	input := "name,age,status\nAlice,30,active\nBob,25,inactive\nCharlie,35,active"

	var rows [][]string
	for row := range sources.CSVRows(context.Background(), strings.NewReader(input)) {
		rows = append(rows, row)
	}

	if len(rows) != 4 || rows[1][0] != "Alice" {
		t.Fatalf("got %v", rows)
	}
}

func TestCSVRows_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var rows [][]string
	for row := range sources.CSVRows(ctx, strings.NewReader("a,b\n1,2")) {
		rows = append(rows, row)
	}

	if len(rows) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(rows))
	}
}

func TestCSVRowsWithError(t *testing.T) {
	input := "\"unterminated"
	var gotErr error

	for _, err := range sources.CSVRowsWithError(context.Background(), strings.NewReader(input)) {
		gotErr = err
	}

	if gotErr == nil {
		t.Fatal("expected csv read error")
	}
}

func TestCSVRowsWithError_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count := 0
	for range sources.CSVRowsWithError(ctx, strings.NewReader("a,b\n1,2")) {
		count++
	}

	if count != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", count)
	}
}

func TestLines(t *testing.T) {
	var lines []string
	for line, err := range sources.Lines(context.Background(), strings.NewReader("line1\nline2\nline3")) {
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, line)
	}

	if !slices.Equal(lines, []string{"line1", "line2", "line3"}) {
		t.Fatalf("got %v", lines)
	}
}

func TestLines_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var lines []string
	for line, err := range sources.Lines(ctx, strings.NewReader("line1\nline2")) {
		if err == nil {
			lines = append(lines, line)
		}
	}

	if len(lines) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(lines))
	}
}

func TestLinesWithError(t *testing.T) {
	longLine := strings.Repeat("a", 1024*1024+1)
	var gotErr error

	for _, err := range sources.LinesWithError(context.Background(), strings.NewReader(longLine)) {
		if err != nil {
			gotErr = err
		}
	}

	if gotErr == nil {
		t.Fatal("expected scanner error")
	}
}

func TestLinesWithError_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count := 0
	for range sources.LinesWithError(ctx, strings.NewReader("line1\nline2")) {
		count++
	}

	if count != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", count)
	}
}

func TestFileLines(t *testing.T) {
	var lines []string
	for line, err := range sources.FileLines(context.Background(), filepath.Join("testdata", "sample.txt")) {
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, line)
	}

	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("got %v", lines)
	}
}

func TestFileLines_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var lines []string
	for line, err := range sources.FileLines(ctx, filepath.Join("testdata", "sample.txt")) {
		if err == nil {
			lines = append(lines, line)
		}
	}

	if len(lines) != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", len(lines))
	}
}

func TestFileLinesWithError(t *testing.T) {
	var gotErr error

	for _, err := range sources.FileLinesWithError(context.Background(), filepath.Join("testdata", "missing.txt")) {
		gotErr = err
	}

	if gotErr == nil || !errors.Is(gotErr, os.ErrNotExist) {
		t.Fatalf("expected file open error, got %v", gotErr)
	}
}

func TestFileLinesWithError_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count := 0
	for range sources.FileLinesWithError(ctx, filepath.Join("testdata", "sample.txt")) {
		count++
	}

	if count != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", count)
	}
}

func TestStdin(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "stdin-*")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString("one\ntwo\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	oldStdin := os.Stdin
	os.Stdin = tmpFile
	defer func() { os.Stdin = oldStdin }()

	var lines []string
	for line, err := range sources.Stdin(context.Background()) {
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, line)
	}

	if !slices.Equal(lines, []string{"one", "two"}) {
		t.Fatalf("got %v", lines)
	}
}

func TestStdin_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count := 0
	for _, err := range sources.Stdin(ctx) {
		if err == nil {
			count++
		}
	}

	if count != 0 {
		t.Fatalf("expected 0 results on cancelled context, got %d", count)
	}
}

func ExampleCSVRows() {
	input := strings.NewReader("name,age\nAlice,30\nBob,25\n")

	for row := range sources.CSVRows(context.Background(), input) {
		fmt.Println(row[0], row[1])
	}
	// Output:
	// name age
	// Alice 30
	// Bob 25
}

func ExampleCSVRowsWithError() {
	input := strings.NewReader("name,age\nAlice,30\nBob,25\n")

	for row, err := range sources.CSVRowsWithError(context.Background(), input) {
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println(row[0], row[1])
	}
	// Output:
	// name age
	// Alice 30
	// Bob 25
}

func ExampleLines() {
	input := strings.NewReader("line1\nline2\nline3\n")

	for line, err := range sources.Lines(context.Background(), input) {
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println(line)
	}
	// Output:
	// line1
	// line2
	// line3
}

func ExampleLinesWithError() {
	input := strings.NewReader("line1\nline2\n")

	for line, err := range sources.LinesWithError(context.Background(), input) {
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println(line)
	}
	// Output:
	// line1
	// line2
}

func ExampleFileLines() {
	for line, err := range sources.FileLines(context.Background(), filepath.Join("testdata", "sample.txt")) {
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println(line)
	}
	// Output:
	// hello\nworld\nfoo
}

func ExampleFileLinesWithError() {
	for line, err := range sources.FileLinesWithError(context.Background(), filepath.Join("testdata", "sample.txt")) {
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println(line)
	}
	// Output:
	// hello\nworld\nfoo
}

func ExampleStdin() {
	tmpFile, err := os.CreateTemp("", "vortex-stdin-example-*")
	if err != nil {
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString("one\ntwo\n"); err != nil {
		return
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return
	}

	oldStdin := os.Stdin
	os.Stdin = tmpFile
	defer func() { os.Stdin = oldStdin }()

	for line, err := range sources.Stdin(context.Background()) {
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println(line)
	}
	// Output:
	// one
	// two
}
