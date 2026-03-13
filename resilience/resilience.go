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

// Backoff calculates how long to wait before the next retry.
type Backoff struct {
	Base       time.Duration // initial wait time
	Max        time.Duration // maximum wait time
	Multiplier float64       // how fast the wait grows
	Jitter     bool          // add randomness to avoid thundering herd
}

// DefaultBackoff is a sensible starting point for most use cases.
var DefaultBackoff = Backoff{
	Base:       100 * time.Millisecond,
	Max:        30 * time.Second,
	Multiplier: 2.0,
	Jitter:     true,
}

// Duration returns how long to wait before attempt number n.
// n starts at 0 for the first retry.
func (b Backoff) Duration(n int) time.Duration {
	// exponential: base * multiplier^n
	d := float64(b.Base) * math.Pow(b.Multiplier, float64(n))

	// cap at max
	if d > float64(b.Max) {
		d = float64(b.Max)
	}

	// add jitter — random value between 0 and d
	// prevents all retrying clients hitting the server at the same time
	if b.Jitter {
		d = d/2 + rand.Float64()*d/2
	}

	return time.Duration(d)
}

// RetryConfig controls how retries behave.
type RetryConfig struct {
	MaxAttempts int     // total attempts including the first one
	Backoff     Backoff // how long to wait between attempts
}

// DefaultRetry is a sensible starting point.
var DefaultRetry = RetryConfig{
	MaxAttempts: 3,
	Backoff:     DefaultBackoff,
}

// Retry calls fn repeatedly until it succeeds or runs out of attempts.
// ctx lets the caller cancel retrying at any time.
func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		// check if caller cancelled before trying
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// call the function
		lastErr = fn()
		if lastErr == nil {
			return nil // success — stop retrying
		}

		// last attempt — don't wait, just return the error
		if attempt == cfg.MaxAttempts-1 {
			break
		}

		// wait before next attempt
		wait := cfg.Backoff.Duration(attempt)
		select {
		case <-time.After(wait): // wait is over, try again
		case <-ctx.Done():       // caller cancelled, stop
			return ctx.Err()
		}
	}

	return fmt.Errorf("all %d attempts failed: %w", cfg.MaxAttempts, lastErr)
}

type circuitState int

const (
	stateClosed   circuitState = iota // normal — requests go through
	stateOpen                         // tripped — requests blocked
	stateHalfOpen                     // testing — one request allowed
)

// CircuitBreaker stops calling a failing service to give it time to recover.
type CircuitBreaker struct {
	maxFailures int           // failures before opening
	timeout     time.Duration // how long to wait before trying again

	state       circuitState
	failures    int
	lastFailure time.Time
	mu          sync.Mutex
}

// NewCircuitBreaker creates a breaker that opens after maxFailures
// consecutive failures and tries again after timeout.
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       stateClosed,
	}
}

var ErrCircuitOpen = errors.New("circuit breaker is open")

// Execute runs fn if the circuit allows it.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	switch cb.state {
	case stateOpen:
		// check if timeout has passed
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = stateHalfOpen // allow one test request
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen // still open — reject immediately
		}
	}

	cb.mu.Unlock()

	// call the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.failures >= cb.maxFailures || cb.state == stateHalfOpen {
			cb.state = stateOpen // trip the breaker
		}
		return err
	}

	// success — reset everything
	cb.failures = 0
	cb.state = stateClosed
	return nil
}