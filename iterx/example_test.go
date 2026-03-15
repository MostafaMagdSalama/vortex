package iterx_test

import (
	"context"
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/MostafaMagdSalama/vortex/iterx"
)

// ExampleChunk demonstrates how to process data in batches, such as for
// bulk database inserts or writing to chunked APIs.
func ExampleChunk() {
	ctx := context.Background()
	logs := slices.Values([]string{"log1", "log2", "log3", "log4", "log5"})

	for batch := range iterx.ChunkSeq(ctx, logs, 2) {
		fmt.Printf("Batch size: %d, items: %v\n", len(batch), batch)
	}
	// Output:
	// Batch size: 2, items: [log1 log2]
	// Batch size: 2, items: [log3 log4]
	// Batch size: 1, items: [log5]
}

// ExampleContains demonstrates checking if an item exists efficiently,
// as the iterator stops processing as soon as it finds a match.
func ExampleContains() {
	ctx := context.Background()
	users := slices.Values([]string{"alice", "bob", "charlie", "admin"})

	hasAdmin := iterx.ContainsSeq(ctx, users, "admin")
	fmt.Println("Has admin:", hasAdmin)
	// Output: Has admin: true
}

// ExampleDistinct demonstrates removing duplicates from an incoming stream,
// such as a sequence of IP addresses.
func ExampleDistinct() {
	ctx := context.Background()
	ipAddresses := slices.Values([]string{"192.168.1.1", "10.0.0.1", "192.168.1.1", "10.0.0.2"})

	for ip := range iterx.DistinctSeq(ctx, ipAddresses) {
		fmt.Println(ip)
	}
	// Output:
	// 192.168.1.1
	// 10.0.0.1
	// 10.0.0.2
}

// ExampleDrain demonstrates exhausting a sequence when the terminal
// operation can fail, like writing to an io.Writer.
func ExampleDrain() {
	ctx := context.Background()
	lines := slices.Values([]string{"header", "data 1", "data 2"})

	// Simulate writing to something that could fail
	var out strings.Builder

	err := iterx.DrainSeq(ctx, lines, func(line string) error {
		_, err := out.WriteString(line + "\n")
		return err // stops early if err != nil
	})

	fmt.Printf("Error: %v, Output:\n%s", err, out.String())
	// Output:
	// Error: <nil>, Output:
	// header
	// data 1
	// data 2
}

// ExampleFilter demonstrates removing unwanted items from a stream.
func ExampleFilter() {
	ctx := context.Background()
	numbers := slices.Values([]int{1, 2, 3, 4, 5, 6})

	evens := iterx.FilterSeq(ctx, numbers, func(n int) bool {
		return n%2 == 0
	})

	for v := range evens {
		fmt.Println(v)
	}
	// Output:
	// 2
	// 4
	// 6
}

type DemoUser struct {
	Name  string
	Roles []string
}

// ExampleFlatMap demonstrates expanding a single item into multiple items.
// In this case: expanding a user into a stream of their roles.
func ExampleFlatMap() {
	ctx := context.Background()
	users := slices.Values([]DemoUser{
		{Name: "Alice", Roles: []string{"admin", "editor"}},
		{Name: "Bob", Roles: []string{"viewer"}},
	})

	roles := iterx.FlatMapSeq(ctx, users, func(u DemoUser) iter.Seq[string] {
		return slices.Values(u.Roles)
	})

	for role := range roles {
		fmt.Println(role)
	}
	// Output:
	// admin
	// editor
	// viewer
}

// ExampleFlatten demonstrates unrolling an iterator of slices into a flat iterator.
func ExampleFlatten() {
	ctx := context.Background()
	batches := slices.Values([][]int{
		{1, 2},
		{3},
		{4, 5, 6},
	})

	for v := range iterx.FlattenSeq(ctx, batches) {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
}

// ExampleForEach demonstrates running a non-failing side effect on every item.
func ExampleForEach() {
	ctx := context.Background()
	messages := slices.Values([]string{"hello", "world"})

	iterx.ForEachSeq(ctx, messages, func(msg string) {
		fmt.Println("Processed:", msg)
	})
	// Output:
	// Processed: hello
	// Processed: world
}

// ExampleMap demonstrates transforming items from one type to another.
func ExampleMap() {
	ctx := context.Background()
	users := slices.Values([]DemoUser{
		{Name: "Alice"},
		{Name: "Bob"},
	})

	names := iterx.MapSeq(ctx, users, func(u DemoUser) string {
		return strings.ToUpper(u.Name) // U is string
	})

	for name := range names {
		fmt.Println(name)
	}
	// Output:
	// ALICE
	// BOB
}

// ExampleReverse demonstrates yielding the sequence in reverse order.
func ExampleReverse() {
	ctx := context.Background()
	steps := slices.Values([]string{"step 1", "step 2", "step 3"})

	for step := range iterx.ReverseSeq(ctx, steps) {
		fmt.Println(step)
	}
	// Output:
	// step 3
	// step 2
	// step 1
}

// ExampleTake demonstrates grabbing just the first N elements.
// Particularly useful with endless streams or large files.
func ExampleTake() {
	ctx := context.Background()
	items := slices.Values([]int{10, 20, 30, 40, 50, 60})

	top3 := iterx.TakeSeq(ctx, items, 3)

	for v := range top3 {
		fmt.Println(v)
	}
	// Output:
	// 10
	// 20
	// 30
}

// ExampleTakeWhile demonstrates reading a stream until a condition stops being met.
// For instance: reading lines from a log until a marker is reached.
func ExampleTakeWhile() {
	ctx := context.Background()
	logs := slices.Values([]string{"ok", "ok", "ok", "error", "ok"})

	for v := range iterx.TakeWhileSeq(ctx, logs, func(s string) bool {
		return s != "error"
	}) {
		fmt.Println(v)
	}
	// Output:
	// ok
	// ok
	// ok
}

// ExampleValidate demonstrates checking structures and extracting validation metadata
// on failing entries without stopping the processing pipeline.
func ExampleValidate() {
	ctx := context.Background()
	emails := slices.Values([]string{"test@example.com", "invalid-email", "hello@world.com"})

	valid := iterx.ValidateSeq(ctx, emails,
		func(email string) (bool, string) {
			if !strings.Contains(email, "@") {
				return false, "missing @"
			}
			return true, ""
		},
		func(err iterx.ValidationError[string]) {
			fmt.Printf("Validation error: %s - %s\n", err.Item, err.Reason)
		},
	)

	for email := range valid {
		fmt.Println("Processed valid email:", email)
	}
	// Output:
	// Processed valid email: test@example.com
	// Validation error: invalid-email - missing @
	// Processed valid email: hello@world.com
}

// ExampleZip demonstrates combining two synchronised streams into pairs.
func ExampleZip() {
	ctx := context.Background()
	keys := slices.Values([]string{"A", "B", "C"})
	values := slices.Values([]int{100, 200}) // Notice: shorter sequence!

	for pair := range iterx.ZipSeq(ctx, keys, values) {
		fmt.Printf("%v: %v\n", pair[0], pair[1])
	}
	// Output:
	// A: 100
	// B: 200
}
