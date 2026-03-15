package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MostafaMagdSalama/vortex/resilience"
	"github.com/MostafaMagdSalama/vortex/sources"
)

type UserRequest struct {
	ID string `json:"id"`
}

func main() {
	ctx := context.Background()

	// Simulate a JSONLines input source of IDs to scrape
	input := strings.NewReader(`{"id": "user_1"}
{"id": "user_2"}
{"id": "user_3"}
{"id": "user_4"}
{"id": "user_5"}
{"id": "user_6"}`)

	requestStream := sources.JSONLines[UserRequest](ctx, input)

	// CircuitBreaker: trips after 2 failures, attempts half-open after 500ms
	breaker := resilience.NewCircuitBreaker(2, 500*time.Millisecond)

	// Retry config: 3 attempts with exponential backoff
	retryCfg := resilience.RetryConfig{
		MaxAttempts: 3,
		Backoff:     resilience.DefaultBackoff,
	}

	// Simulated unstable API — fails on calls 2 and 4
	apiCalls := 0
	scrapeAPI := func(ctx context.Context, id string) (string, error) {
		apiCalls++
		fmt.Printf("  [API] call #%d for %s\n", apiCalls, id)
		if apiCalls == 2 || apiCalls == 4 {
			return "", fmt.Errorf("HTTP 503 Service Unavailable")
		}
		return fmt.Sprintf("scraped data for %s", id), nil
	}

	// Pipeline: Seq2[string, error] — data and errors stay separate
	results := func(yield func(string, error) bool) {
		for req, err := range requestStream {
			// stream read error — surface and continue
			if err != nil {
				if !yield("", fmt.Errorf("stream error for request: %w", err)) {
					return
				}
				continue
			}

			fmt.Printf("\n[PIPELINE] processing %s | circuit=%s\n", req.ID, breaker.State())

			var scrapedData string

			// Retry wraps CircuitBreaker:
			// — if circuit is open, Execute returns ErrCircuitOpen immediately
			// — ErrCircuitOpen is not Retryable so Retry stops immediately
			// — only Retryable errors trigger a retry attempt
			retryErr := resilience.Retry(ctx, retryCfg, func(attempt int) error {
				if attempt > 0 {
					fmt.Printf("  [RETRY] attempt %d for %s | circuit=%s\n", attempt+1, req.ID, breaker.State())
				}

				return breaker.Execute(ctx, func(ctx context.Context) error {
					data, apiErr := scrapeAPI(ctx, req.ID)
					if apiErr != nil {
						// wrap as retryable so Retry will try again
						return resilience.Retryable(apiErr)
					}
					scrapedData = data
					return nil
				})
			})

			if !yield(scrapedData, retryErr) {
				return
			}
		}
	}

	// Terminal step — consume results
	fmt.Println("=== scrape pipeline ===")
	successCount := 0
	failCount := 0

	for result, err := range results {
		if err != nil {
			failCount++
			fmt.Printf("[FAIL] %v\n", err)
			continue
		}
		successCount++
		fmt.Printf("[OK]   %s\n", result)
	}

	// Circuit breaker stats
	stats := breaker.Stats()
	fmt.Println()
	fmt.Println("=== summary ===")
	fmt.Printf("success  : %d\n", successCount)
	fmt.Printf("failed   : %d\n", failCount)
	fmt.Printf("api calls: %d\n", apiCalls)
	fmt.Println()
	fmt.Println("=== circuit breaker stats ===")
	fmt.Printf("state    : %s\n", stats.State)
	fmt.Printf("requests : %d\n", stats.Requests)
	fmt.Printf("failures : %d\n", stats.Failures)
	fmt.Printf("successes: %d\n", stats.Successes)
	fmt.Printf("rejected : %d\n", stats.Rejected)
}
