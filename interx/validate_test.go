package interx_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	vinterx "github.com/MostafaMagdSalama/vortex/interx"
)

type User struct {
	ID     string
	Name   string
	Email  string
	Status string
}

func validateUser(u User) (bool, string) {
	if u.ID == "" {
		return false, "missing ID"
	}
	if u.Name == "" {
		return false, "missing name"
	}
	if u.Email == "" {
		return false, "missing email"
	}
	if u.Status != "active" && u.Status != "inactive" {
		return false, fmt.Sprintf("invalid status: %s", u.Status)
	}
	return true, ""
}

func TestValidate_AllValid(t *testing.T) {
	users := slices.Values([]User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", Status: "active"},
		{ID: "2", Name: "Bob", Email: "bob@example.com", Status: "inactive"},
	})

	var errList []vinterx.ValidationError[User]
	valid := vinterx.Validate(context.Background(), users, validateUser, func(ve vinterx.ValidationError[User]) {
		errList = append(errList, ve)
	})

	var validList []User
	for u := range valid {
		validList = append(validList, u)
	}

	if len(validList) != 2 || len(errList) != 0 {
		t.Fatalf("valid=%d errors=%d", len(validList), len(errList))
	}
}

func TestValidate_AllInvalid(t *testing.T) {
	users := slices.Values([]User{
		{ID: "", Name: "Alice", Email: "alice@example.com", Status: "active"},
		{ID: "2", Name: "", Email: "bob@example.com", Status: "inactive"},
		{ID: "3", Name: "Charlie", Email: "", Status: "active"},
	})

	var errList []vinterx.ValidationError[User]
	valid := vinterx.Validate(context.Background(), users, validateUser, func(ve vinterx.ValidationError[User]) {
		errList = append(errList, ve)
	})

	var validList []User
	for u := range valid {
		validList = append(validList, u)
	}

	if len(validList) != 0 || len(errList) != 3 {
		t.Fatalf("valid=%d errors=%d", len(validList), len(errList))
	}
}

func TestValidate_ErrorReasons(t *testing.T) {
	users := slices.Values([]User{
		{ID: "", Name: "Alice", Email: "alice@example.com", Status: "active"},
		{ID: "2", Name: "", Email: "bob@example.com", Status: "inactive"},
		{ID: "3", Name: "Charlie", Email: "", Status: "active"},
		{ID: "4", Name: "Diana", Email: "diana@example.com", Status: "unknown"},
	})

	expected := []string{"missing ID", "missing name", "missing email", "invalid status: unknown"}
	var got []string
	valid := vinterx.Validate(context.Background(), users, validateUser, func(ve vinterx.ValidationError[User]) {
		got = append(got, ve.Reason)
	})

	for range valid {
	}

	if !slices.Equal(got, expected) {
		t.Fatalf("got %v", got)
	}
}

func TestValidate_NilCallback(t *testing.T) {
	users := slices.Values([]User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", Status: "active"},
		{ID: "", Name: "Bob", Email: "bob@example.com", Status: "inactive"},
	})

	var validList []User
	for u := range vinterx.Validate(context.Background(), users, validateUser, nil) {
		validList = append(validList, u)
	}

	if len(validList) != 1 {
		t.Fatalf("expected 1 valid user, got %d", len(validList))
	}
}

func TestValidate_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	users := slices.Values([]User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", Status: "active"},
		{ID: "2", Name: "Bob", Email: "bob@example.com", Status: "inactive"},
	})

	var validList []User
	var errList []vinterx.ValidationError[User]
	for u := range vinterx.Validate(ctx, users, validateUser, func(ve vinterx.ValidationError[User]) {
		errList = append(errList, ve)
	}) {
		validList = append(validList, u)
	}

	if len(validList) != 0 {
		t.Fatalf("expected 0 valid results on cancelled context, got %d", len(validList))
	}
	if len(errList) != 0 {
		t.Fatalf("expected 0 errors on cancelled context, got %d", len(errList))
	}
}

func ExampleValidate() {
	users := slices.Values([]User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", Status: "active"},
		{ID: "", Name: "Bob", Email: "bob@example.com", Status: "inactive"},
		{ID: "3", Name: "Carol", Email: "carol@example.com", Status: "active"},
	})

	for user := range vinterx.Validate(context.Background(), users, validateUser, func(ve vinterx.ValidationError[User]) {
		fmt.Println("invalid:", ve.Reason)
	}) {
		fmt.Println("valid:", user.Name)
	}
	// Output:
	// valid: Alice
	// invalid: missing ID
	// valid: Carol
}
