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
{"id": "user_4"}`)

	requestStream := sources.JSONLines[UserRequest](ctx, input)

	// Create a CircuitBreaker for the scraping endpoint
	// Trips after 2 failures, attempts half-open after 500ms
	breaker := resilience.NewCircuitBreaker(2, 500*time.Millisecond)

	// A simulated unstable HTTP API
	apiCalls := 0
	scrapeAPI := func(ctx context.Context, id string) (string, error) {
		apiCalls++
		if apiCalls == 2 {
			return "", fmt.Errorf("HTTP 503 Service Unavailable")
		}
		return fmt.Sprintf("Scraped Data for %s", id), nil
	}

	// For each ID, attempt to scrape it utilizing Retry + CircuitBreaker
	results := func(yield func(string) bool) {
		for req, err := range requestStream {
			if err != nil {
				if !yield(fmt.Sprintf("Stream err: %v", err)) {
					return
				}
				continue
			}

			var scrapedData string
			// Attempt the scrape with Retries
			err = resilience.Retry(ctx, resilience.DefaultRetry, func(attempt int) error {
				// Pass the execution through the Circuit Breaker
				return breaker.Execute(ctx, func(ctx context.Context) error {
					data, err := scrapeAPI(ctx, req.ID)
					if err != nil {
						return resilience.Retryable(err)
					}
					scrapedData = data
					return nil
				})
			})

			if err != nil {
				if !yield(fmt.Sprintf("Failed %s: %v", req.ID, err)) {
					return
				}
				continue
			}

			if !yield(scrapedData) {
				return
			}
		}
	}

	// Process outcomes
	for result := range results {
		fmt.Println("Result:", result)
	}
}
