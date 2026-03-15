package sources

import (
	"context"
	"fmt"
	"strings"

	"github.com/MostafaMagdSalama/vortex/iterx"
)

func ExampleCSVRows_filter() {
	input := "name,email,status\nAlice,alice@example.com,active\nBob,bob@example.com,inactive\nCharlie,charlie@example.com,active\n"
	r := strings.NewReader(input)

	// skip header and filter active users
	first := true
	for row, err := range CSVRows(context.Background(), r) {
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
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

	rows := func(yield func([]string) bool) {
		for row, err := range CSVRows(ctx, r) {
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			if !yield(row) {
				return
			}
		}
	}

	first := true
	dataRows := iterx.FilterSeq(ctx, rows, func(row []string) bool {
		if first {
			first = false
			return false
		}
		return true
	})

	names := iterx.MapSeq(ctx,
		iterx.FilterSeq(ctx, dataRows, func(row []string) bool {
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
