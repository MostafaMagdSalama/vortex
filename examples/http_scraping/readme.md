# Example: Resilient Scraping Pipeline (Circuit Breaker + Retry)

This example demonstrates how to build a fault-tolerant data scraping pipeline using [vortex](https://github.com/MostafaMagdSalama/vortex)'s resilience primitives — combining a **circuit breaker** and **exponential backoff retry** around an unstable API, all driven by a lazy JSONLines input stream.

---

## What it does

1. **Extract** — reads a stream of user IDs from a JSONLines source using `sources.JSONLines`
2. **Transform** — for each ID, calls a simulated unstable API wrapped in:
   - A **retry** layer (up to 3 attempts with exponential backoff)
   - A **circuit breaker** (trips after 2 consecutive failures, attempts recovery after 500ms)
3. **Load** — collects results and errors separately, printing a summary with circuit breaker stats

The API is rigged to fail on calls #2 and #4 to demonstrate retry behavior. Errors are surfaced as values in the stream — **successes and failures flow through the same pipeline without panicking or short-circuiting**.

---

## How resilience layers interact

```
Retry
  └── CircuitBreaker
        └── API call
```

- If the **circuit is open**, `CircuitBreaker.Execute` returns `ErrCircuitOpen` immediately — no API call is made
- `ErrCircuitOpen` is **not retryable**, so `Retry` stops immediately and surfaces the error
- If the API returns a transient error wrapped with `resilience.Retryable(...)`, `Retry` will attempt again up to `MaxAttempts`
- After `MaxAttempts` are exhausted, the error is yielded downstream as a regular value

This composition means the retry layer never fights the circuit breaker — once the circuit trips, all pending retries for that request stop instantly.

---

## How to run

```bash
go run main.go
```

---

## Output

> The simulated API fails on calls #2 and #4, triggering retries. Both requests eventually succeed after one retry each.

```text
=== scrape pipeline ===

[PIPELINE] processing user_1 | circuit=closed
  [API] call #1 for user_1
[OK]   scraped data for user_1

[PIPELINE] processing user_2 | circuit=closed
  [API] call #2 for user_2
  [RETRY] attempt 2 for user_2 | circuit=closed
  [API] call #3 for user_2
[OK]   scraped data for user_2

[PIPELINE] processing user_3 | circuit=closed
  [API] call #4 for user_3
  [RETRY] attempt 2 for user_3 | circuit=closed
  [API] call #5 for user_3
[OK]   scraped data for user_3

[PIPELINE] processing user_4 | circuit=closed
  [API] call #6 for user_4
[OK]   scraped data for user_4

[PIPELINE] processing user_5 | circuit=closed
  [API] call #7 for user_5
[OK]   scraped data for user_5

[PIPELINE] processing user_6 | circuit=closed
  [API] call #8 for user_6
[OK]   scraped data for user_6

=== summary ===
success  : 6
failed   : 0
api calls: 8

=== circuit breaker stats ===
state    : closed
requests : 8
failures : 2
successes: 6
rejected : 0
```

### Reading the output

| Stat | Value | Meaning |
|---|---|---|
| success | 6 | All 6 IDs eventually produced data |
| failed | 0 | No request exhausted all retries |
| api calls | 8 | 6 clean calls + 2 retries from transient failures |
| rejected | 0 | Circuit never tripped (failures were not consecutive enough) |
| state | closed | Breaker remained healthy throughout |

The circuit stayed closed because the two failures were spread apart and did not hit the threshold of 2 back-to-back failures. In a real scenario with a sustained outage, the breaker would trip and subsequent requests would be rejected immediately — saving downstream resources and failing fast instead of waiting on timeouts.

---

## Key vortex APIs used

| API | Purpose |
|---|---|
| `sources.JSONLines` | Parses a JSONLines stream into a lazy typed iterator |
| `resilience.NewCircuitBreaker` | Creates a circuit breaker with a failure threshold and recovery window |
| `resilience.Retry` | Retries a function with configurable attempts and backoff strategy |
| `resilience.Retryable` | Wraps an error to signal that it is safe to retry |
| `resilience.DefaultBackoff` | Exponential backoff strategy used between retry attempts |

---

## Files

| File | Description |
|---|---|
| `main.go` | Full scraping pipeline with retry, circuit breaker, and JSONLines source |