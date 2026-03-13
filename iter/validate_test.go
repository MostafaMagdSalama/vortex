package iter_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex/iter"
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

// all users are valid — errList should be empty
func TestValidate_AllValid(t *testing.T) {
	users := slices.Values([]User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", Status: "active"},
		{ID: "2", Name: "Bob", Email: "bob@example.com", Status: "inactive"},
	})

	var errList []iter.ValidationError[User]

	valid := iter.Validate(users, validateUser, func(ve iter.ValidationError[User]) {
		errList = append(errList, ve)
	})

	var validList []User
	for u := range valid {
		validList = append(validList, u)
	}

	if len(validList) != 2 {
		t.Fatalf("expected 2 valid users, got %d", len(validList))
	}
	if len(errList) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errList))
	}
}

// all users are invalid — validList should be empty
func TestValidate_AllInvalid(t *testing.T) {
	users := slices.Values([]User{
		{ID: "", Name: "Alice", Email: "alice@example.com", Status: "active"},
		{ID: "2", Name: "", Email: "bob@example.com", Status: "inactive"},
		{ID: "3", Name: "Charlie", Email: "", Status: "active"},
	})

	var errList []iter.ValidationError[User]

	valid := iter.Validate(users, validateUser, func(ve iter.ValidationError[User]) {
		errList = append(errList, ve)
	})

	var validList []User
	for u := range valid {
		validList = append(validList, u)
	}

	if len(validList) != 0 {
		t.Fatalf("expected 0 valid users, got %d", len(validList))
	}
	if len(errList) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(errList))
	}
}

// mix of valid and invalid — check counts
func TestValidate_Mixed(t *testing.T) {
	users := slices.Values([]User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", Status: "active"},    // valid
		{ID: "", Name: "Bob", Email: "bob@example.com", Status: "inactive"},       // invalid — missing ID
		{ID: "3", Name: "Charlie", Email: "charlie@example.com", Status: "active"}, // valid
		{ID: "4", Name: "Diana", Email: "", Status: "active"},                     // invalid — missing email
		{ID: "5", Name: "Eve", Email: "eve@example.com", Status: "unknown"},       // invalid — bad status
	})

	var errList []iter.ValidationError[User]

	valid := iter.Validate(users, validateUser, func(ve iter.ValidationError[User]) {
		errList = append(errList, ve)
	})

	var validList []User
	for u := range valid {
		validList = append(validList, u)
	}

	if len(validList) != 2 {
		t.Fatalf("expected 2 valid users, got %d", len(validList))
	}
	if len(errList) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(errList))
	}
}

// check that error reasons are correct
func TestValidate_ErrorReasons(t *testing.T) {
	users := slices.Values([]User{
		{ID: "", Name: "Alice", Email: "alice@example.com", Status: "active"},
		{ID: "2", Name: "", Email: "bob@example.com", Status: "inactive"},
		{ID: "3", Name: "Charlie", Email: "", Status: "active"},
		{ID: "4", Name: "Diana", Email: "diana@example.com", Status: "unknown"},
	})

	expected := []string{
		"missing ID",
		"missing name",
		"missing email",
		"invalid status: unknown",
	}

	i := 0
	valid := iter.Validate(users, validateUser, func(ve iter.ValidationError[User]) {
		if i >= len(expected) {
			t.Errorf("unexpected extra error: %s", ve.Reason)
			return
		}
		if ve.Reason != expected[i] {
			t.Errorf("error %d: expected reason %q, got %q", i, expected[i], ve.Reason)
		}
		i++
	})

	// drain valid to trigger the pipeline
	for range valid {}

	if i != len(expected) {
		t.Fatalf("expected %d errors, got %d", len(expected), i)
	}
}

// empty sequence — both validList and errList should be empty
func TestValidate_EmptySequence(t *testing.T) {
	users := slices.Values([]User{})

	var errList []iter.ValidationError[User]

	valid := iter.Validate(users, validateUser, func(ve iter.ValidationError[User]) {
		errList = append(errList, ve)
	})

	var validList []User
	for u := range valid {
		validList = append(validList, u)
	}

	if len(validList) != 0 {
		t.Fatalf("expected 0 valid, got %d", len(validList))
	}
	if len(errList) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errList))
	}
}

// correct item is attached to the error
func TestValidate_CorrectItemInError(t *testing.T) {
	badUser := User{ID: "", Name: "Bob", Email: "bob@example.com", Status: "inactive"}

	users := slices.Values([]User{badUser})

	var got iter.ValidationError[User]

	valid := iter.Validate(users, validateUser, func(ve iter.ValidationError[User]) {
		got = ve
	})

	// drain valid to trigger the pipeline
	for range valid {}

	if got.Item != badUser {
		t.Fatalf("expected item %v in error, got %v", badUser, got.Item)
	}
	if got.Reason != "missing ID" {
		t.Fatalf("expected reason 'missing ID', got %q", got.Reason)
	}
}

// nil callback — invalid items silently dropped, no panic
func TestValidate_NilCallback(t *testing.T) {
	users := slices.Values([]User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", Status: "active"},
		{ID: "", Name: "Bob", Email: "bob@example.com", Status: "inactive"},
	})

	valid := iter.Validate(users, validateUser, nil)

	var validList []User
	for u := range valid {
		validList = append(validList, u)
	}

	if len(validList) != 1 {
		t.Fatalf("expected 1 valid user, got %d", len(validList))
	}
}