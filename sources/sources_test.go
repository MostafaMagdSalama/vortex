package sources_test

import (
	"strings"
	"testing"

	"github.com/MostafaMagdSalama/vortex/sources"
)

func TestCSVRows(t *testing.T) {
	input := "name,age,status\nAlice,30,active\nBob,25,inactive\nCharlie,35,active"

	var rows [][]string
	for row := range sources.CSVRows(strings.NewReader(input)) {
		rows = append(rows, row)
	}

	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(rows))
	}
	if rows[1][0] != "Alice" {
		t.Fatalf("expected Alice, got %s", rows[1][0])
	}
}

func TestLines(t *testing.T) {
	input := "line1\nline2\nline3"

	var lines []string
	for line := range sources.Lines(strings.NewReader(input)) {
		lines = append(lines, line)
	}

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line1" {
		t.Fatalf("expected line1, got %s", lines[0])
	}
}

func TestFileLines(t *testing.T) {
	for line := range sources.FileLines("testdata/sample.txt") {
		if line == "" {
			t.Fatal("got empty line")
		}
	}
}