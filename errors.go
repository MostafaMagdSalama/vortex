// Package vortex provides lazy evaluation and structured concurrency
// for data pipeline development.
//
// # Error Handling
//
// Vortex exposes a unified error architecture to make error handling transparent and robust.
// All operations that encounter underlying failures (like I/O, network, or decoding errors)
// wrap those errors in a [vortex.Error], which provides context about what operation failed
// while preserving the original error for inspection via [errors.As].
//
//	err := iterx.Drain(ctx, seq, processFn)
//	var vErr *vortex.Error
//	if errors.As(err, &vErr) {
//	    fmt.Printf("Operation: %s, Cause: %v\n", vErr.Op, vErr.Err)
//	}
//
// Vortex also provides Sentinel Errors for predictable failure states, like [ErrCancelled],
// which can be checked using [errors.Is].
package vortex

import (
	"errors"
	"fmt"
)

// Sentinel errors tailored for standard Vortex operations.
var (
	ErrCancelled  = errors.New("vortex: operation cancelled")
	ErrValidation = errors.New("vortex: validation failed")
)

// Error represents a Vortex library error.
type Error struct {
	Op  string // e.g., "jsonlines.Read", "parallel.Map"
	Err error  // The underlying error (e.g., io.EOF, sql.ErrNoRows)
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("vortex: %s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("vortex: %s failed", e.Op)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Wrap is a helper to easily construct these errors.
// It returns nil if the provided err is nil.
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{Op: op, Err: err}
}
