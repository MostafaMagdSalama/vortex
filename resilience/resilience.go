package resilience

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// -----------------------------
// Errors
// -----------------------------

var ErrCircuitOpen = errors.New("circuit breaker is open")

// RetryableError wraps an error and signals that the operation can be retried.
// Use Retryable() to wrap errors that should be retried and return unwrapped
// errors for failures that should stop retrying immediately.
type RetryableError struct {
	Err error
}

func (e RetryableError) Error() string { return e.Err.Error() }
func (e RetryableError) Unwrap() error { return e.Err }

// Retryable wraps err to signal that the operation should be retried.
func Retryable(err error) error {
	return RetryableError{Err: err}
}

// IsRetryable reports whether err is a RetryableError.
func IsRetryable(err error) bool {
	var r RetryableError
	return errors.As(err, &r)
}

// -----------------------------
// Backoff
// -----------------------------

// Backoff calculates how long to wait before the next retry.
type Backoff struct {
	Base       time.Duration
	Max        time.Duration
	Multiplier float64
	Jitter     bool
}

// DefaultBackoff is a sensible starting point for most use cases.
var DefaultBackoff = Backoff{
	Base:       100 * time.Millisecond,
	Max:        30 * time.Second,
	Multiplier: 2.0,
	Jitter:     true,
}

// Duration returns how long to wait before attempt number n.
func (b Backoff) Duration(attempt int) time.Duration {
	d := float64(b.Base) * math.Pow(b.Multiplier, float64(attempt))

	if d > float64(b.Max) {
		d = float64(b.Max)
	}

	// jitter between 50% and 100% of d
	// prevents all retrying clients hitting the server at the same time
	// using 50-100% range avoids near-zero wait times
	if b.Jitter {
		d = d/2 + rand.Float64()*d/2
	}

	return time.Duration(d)
}

// -----------------------------
// Retry
// -----------------------------

// RetryConfig controls how retries behave.
type RetryConfig struct {
	MaxAttempts int
	Backoff     Backoff
}

// DefaultRetry is a sensible starting point.
var DefaultRetry = RetryConfig{
	MaxAttempts: 3,
	Backoff:     DefaultBackoff,
}

// Retry calls fn repeatedly until it succeeds or runs out of attempts.
// fn receives the current attempt number starting at 0.
// Only errors wrapped with Retryable() are retried — all other errors
// stop retrying immediately and are returned as-is.
//
// example:
//
//	err := resilience.Retry(ctx, resilience.DefaultRetry, func(attempt int) error {
//	    if err := callAPI(); err != nil {
//	        if isTransient(err) {
//	            return resilience.Retryable(err) // retry this
//	        }
//	        return err // stop immediately — not retryable
//	    }
//	    return nil
//	})
func Retry(ctx context.Context, cfg RetryConfig, fn func(attempt int) error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn(attempt)
		if err == nil {
			return nil
		}

		lastErr = err

		// non-retryable error — stop immediately
		if !IsRetryable(err) {
			return err
		}

		// last attempt — do not wait
		if attempt == cfg.MaxAttempts-1 {
			break
		}

		wait := cfg.Backoff.Duration(attempt)
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
}

// -----------------------------
// Circuit Breaker
// -----------------------------

// CircuitState represents the current state of a CircuitBreaker.
type CircuitState int

const (
	StateClosed   CircuitState = iota // normal — requests go through
	StateOpen                         // tripped — requests blocked
	StateHalfOpen                     // testing — one request allowed
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Stats holds runtime metrics for a CircuitBreaker.
type Stats struct {
	Requests  int64        // total calls to Execute
	Failures  int64        // calls where fn returned an error
	Successes int64        // calls where fn returned nil
	Rejected  int64        // calls rejected because circuit was open
	State     CircuitState // current state of the breaker
}

// CircuitBreaker stops calling a failing service to give it time to recover.
type CircuitBreaker struct {
	maxFailures      int
	timeout          time.Duration
	state            CircuitState
	failures         int
	lastFailure      time.Time
	halfOpenInFlight bool
	stats            Stats
	mu               sync.Mutex
}

// NewCircuitBreaker creates a breaker that opens after maxFailures consecutive
// failures and tries again after timeout.
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       StateClosed,
	}
}

// Execute runs fn if the circuit allows it.
// ctx is passed to fn so the call can be cancelled.
// In half-open state only one trial request is allowed at a time.
//
// example:
//
//	err := cb.Execute(ctx, func(ctx context.Context) error {
//	    return callAPI(ctx)
//	})
//	if errors.Is(err, resilience.ErrCircuitOpen) {
//	    // serve from cache instead
//	}
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	cb.mu.Lock()
	cb.stats.Requests++

	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.timeout {
			// timeout passed — allow one trial request
			cb.state = StateHalfOpen
			cb.halfOpenInFlight = true
		} else {
			// still open — reject immediately
			cb.stats.Rejected++
			cb.mu.Unlock()
			return ErrCircuitOpen
		}

	case StateHalfOpen:
		// trial request already in flight — reject everyone else
		if cb.halfOpenInFlight {
			cb.stats.Rejected++
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
		cb.halfOpenInFlight = true
	}

	cb.mu.Unlock()

	// call fn outside the lock — does not block other goroutines
	err := fn(ctx)

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.stats.Failures++
		cb.failures++
		cb.lastFailure = time.Now()

		// trip to open if threshold reached or trial request failed
		if cb.state == StateHalfOpen || cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}

		cb.halfOpenInFlight = false
		return err
	}

	// success
	cb.stats.Successes++

	if cb.state == StateHalfOpen {
		// recovering from half-open — full reset
		cb.failures = 0
		cb.state = StateClosed
	} else if cb.state == StateClosed {
		// reset failures on success to ensure we only trip on consecutive failures
		cb.failures = 0
	}

	cb.halfOpenInFlight = false
	return nil
}

// Stats returns a snapshot of the breaker's runtime metrics.
func (cb *CircuitBreaker) Stats() Stats {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	s := cb.stats
	s.State = cb.state
	return s
}

// State returns the current state of the breaker.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
