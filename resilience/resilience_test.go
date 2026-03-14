package resilience_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
	cb := resilience.NewCircuitBreaker(3, 10*time.Second) // ← real timeout

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

func TestCircuitBreaker_HalfOpenOnlyOneRequest(t *testing.T) {
	// opens immediately after 1 failure, recovers after 0 timeout
	cb := resilience.NewCircuitBreaker(1, 0)

	// trip the breaker
	cb.Execute(func() error {
		return errors.New("fail")
	})

	// wait for timeout to pass
	time.Sleep(10 * time.Millisecond)

	// fire two concurrent requests in half-open window
	var trialCount atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.Execute(func() error {
				trialCount.Add(1)
				time.Sleep(50 * time.Millisecond) // hold the slot open
				return nil
			})
		}()
	}

	wg.Wait()

	// only ONE trial request should have run
	if trialCount.Load() != 1 {
		t.Fatalf("expected 1 trial request, got %d", trialCount.Load())
	}
}

func TestCircuitBreaker_RespectsMaxFailures(t *testing.T) {
	cb := resilience.NewCircuitBreaker(3, 10*time.Second)

	failFn := func() error { return errors.New("fail") }
	successFn := func() error { return nil }

	// 1st failure — circuit should stay closed
	cb.Execute(failFn)
	if err := cb.Execute(successFn); err != nil {
		t.Fatalf("expected closed after 1 failure, got: %v", err)
	}

	// 2nd failure — circuit should stay closed
	cb.Execute(failFn)
	if err := cb.Execute(successFn); err != nil {
		t.Fatalf("expected closed after 2 failures, got: %v", err)
	}

	// 3rd failure — circuit should now open
	cb.Execute(failFn)

	// next call should be rejected — circuit is open
	if err := cb.Execute(successFn); !errors.Is(err, resilience.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen after 3 failures, got: %v", err)
	}
}
