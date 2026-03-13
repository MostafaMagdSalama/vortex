package resilience_test

import (
	"context"
	"errors"
	"testing"

	"github.com/MostafaMagdSalama/vortex/resilience"
)

func TestRetry(t *testing.T) {
	attempts := 0

	err := resilience.Retry(context.Background(), resilience.RetryConfig{
		MaxAttempts: 3,
		Backoff:     resilience.Backoff{Base: 0}, // no wait in tests
	}, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("not yet")
		}
		return nil // succeeds on 3rd attempt
	})

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestCircuitBreaker(t *testing.T) {
	cb := resilience.NewCircuitBreaker(3, 0) // opens after 3 failures

	// fail 3 times to open the circuit
	for i := 0; i < 3; i++ {
		cb.Execute(func() error {
			return errors.New("fail")
		})
	}

	// now circuit is open — should reject immediately
	err := cb.Execute(func() error {
		return nil // this should never run
	})

	if !errors.Is(err, resilience.ErrCircuitOpen) {
		t.Fatalf("expected circuit open error, got %v", err)
	}
}