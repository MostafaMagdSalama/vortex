package vortex_test

import (
	"errors"
	"testing"

	"github.com/MostafaMagdSalama/vortex"
)

func TestWrap_NilError(t *testing.T) {
	if got := vortex.Wrap("op", nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestWrap_RegularError(t *testing.T) {
	underlying := errors.New("something failed")
	err := vortex.Wrap("myOp", underlying)

	var vErr *vortex.Error
	if !errors.As(err, &vErr) {
		t.Fatalf("expected *vortex.Error, got %T", err)
	}
	if vErr.Op != "myOp" {
		t.Fatalf("expected Op %q, got %q", "myOp", vErr.Op)
	}
	if !errors.Is(err, underlying) {
		t.Fatalf("expected error chain to contain underlying error")
	}
}

func TestWrap_NestedVortexError(t *testing.T) {
	inner := errors.New("root cause")
	first := vortex.Wrap("firstOp", inner)
	second := vortex.Wrap("secondOp", first)

	var vErr *vortex.Error
	if !errors.As(second, &vErr) {
		t.Fatalf("expected *vortex.Error")
	}
	if vErr.Op != "secondOp" {
		t.Fatalf("expected Op %q, got %q", "secondOp", vErr.Op)
	}
	// Wrap must flatten chains: inner .Err must be the root cause, not another *vortex.Error
	var nested *vortex.Error
	if errors.As(vErr.Err, &nested) {
		t.Fatalf("Wrap should flatten nested vortex.Error chains, got nested %T", vErr.Err)
	}
	if !errors.Is(second, inner) {
		t.Fatalf("expected error chain to still reach root cause")
	}
}

func TestWrapCancelled_SentinelCheck(t *testing.T) {
	err := vortex.WrapCancelled("someOp")
	if !errors.Is(err, vortex.ErrCancelled) {
		t.Fatalf("expected errors.Is(err, ErrCancelled) to be true")
	}
}

func TestWrapCancelled_Op(t *testing.T) {
	err := vortex.WrapCancelled("someOp")

	var vErr *vortex.Error
	if !errors.As(err, &vErr) {
		t.Fatalf("expected *vortex.Error, got %T", err)
	}
	if vErr.Op != "someOp" {
		t.Fatalf("expected Op %q, got %q", "someOp", vErr.Op)
	}
	if vErr.Err != vortex.ErrCancelled {
		t.Fatalf("expected Err to be ErrCancelled, got %v", vErr.Err)
	}
}

func TestError_ErrorString_WithErr(t *testing.T) {
	e := &vortex.Error{Op: "myOp", Err: errors.New("bad thing")}
	got := e.Error()
	want := "vortex: myOp: bad thing"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestError_ErrorString_NilErr(t *testing.T) {
	e := &vortex.Error{Op: "myOp"}
	got := e.Error()
	want := "vortex: myOp failed"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestError_Unwrap(t *testing.T) {
	underlying := errors.New("original")
	e := &vortex.Error{Op: "op", Err: underlying}
	if e.Unwrap() != underlying {
		t.Fatalf("expected Unwrap to return underlying error")
	}
	if !errors.Is(e, underlying) {
		t.Fatalf("expected errors.Is chain to work via Unwrap")
	}
}
