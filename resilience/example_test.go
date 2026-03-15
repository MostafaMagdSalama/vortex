package resilience_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MostafaMagdSalama/vortex"
	"github.com/MostafaMagdSalama/vortex/resilience"
)

func ExampleRetry() {
	// A simulated API call that fails once before succeeding.
	failures := 0
	callAPI := func() error {
		if failures < 1 {
			failures++
			return errors.New("temporary network error")
		}
		return nil
	}

	// Retry executes the function until it succeeds or MaxAttempts is reached.
	// Only errors wrapped in resilience.Retryable() are retried.
	err := resilience.Retry(context.Background(), resilience.DefaultRetry, func(attempt int) error {
		fmt.Printf("Attempt %d...\n", attempt)
		
		if err := callAPI(); err != nil {
			// Wrap the error to tell Retry that this is a transient failure.
			return resilience.Retryable(err)
		}
		
		fmt.Println("Success!")
		return nil
	})

	if err != nil {
		fmt.Println("Final Error:", err)
	}

	// Output:
	// Attempt 0...
	// Attempt 1...
	// Success!
}

func ExampleCircuitBreaker() {
	// Create a CircuitBreaker that opens after 2 consecutive failures
	// and waits 100 milliseconds before allowing a trial request (Half-Open).
	cb := resilience.NewCircuitBreaker(2, 100*time.Millisecond)

	// A simulated failing API
	failingAPI := func(ctx context.Context) error {
		return errors.New("500 Internal Server Error")
	}

	// 1. First failure
	_ = cb.Execute(context.Background(), failingAPI)

	// 2. Second failure - The circuit now trips to "Open"
	_ = cb.Execute(context.Background(), failingAPI)

	// 3. Third request - Rejected immediately without calling the API
	err := cb.Execute(context.Background(), failingAPI)
	if errors.Is(err, vortex.ErrCircuitOpen) {
		fmt.Println("Circuit is open! Falling back to cache.")
	}

	// Output:
	// Circuit is open! Falling back to cache.
}
