package sources

import (
	"context"
	"fmt"
	"strings"

	"github.com/MostafaMagdSalama/vortex/interx"
)

func ExampleCSVRows_filter() {
	input := "name,email,status\nAlice,alice@example.com,active\nBob,bob@example.com,inactive\nCharlie,charlie@example.com,active\n"
	r := strings.NewReader(input)

	// skip header and filter active users
	first := true
	for row := range CSVRows(context.Background(), r) {
		if first {
			first = false
			continue
		}
		if row[2] == "active" {
			fmt.Println(row[0])
		}
	}
	// Output:
	// Alice
	// Charlie
}

func ExampleCSVRows_pipeline() {
	input := "name,email,status\nAlice,alice@example.com,active\nBob,bob@example.com,inactive\nCharlie,charlie@example.com,active\n"
	r := strings.NewReader(input)

	ctx := context.Background()

	rows := CSVRows(ctx, r)

	first := true
	dataRows := interx.Filter(ctx, rows, func(row []string) bool {
		if first {
			first = false
			return false
		}
		return true
	})

	names := interx.Map(ctx,
		interx.Filter(ctx, dataRows, func(row []string) bool {
			return row[2] == "active"
		}),
		func(row []string) string {
			return row[0]
		},
	)

	for name := range names {
		fmt.Println(name)
	}
	// Output:
	// Alice
	// Charlie
}
