package main

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"slices"
	"time"

	"github.com/MostafaMagdSalama/vortex/parallel"
)

const (
	TOTAL_USERS = 2_000
	WORKERS     = 50
)

var countries = []string{"US", "UK", "DE", "FR", "JP", "BR", "IN", "CA", "AU", "MX"}

// EnrichedUser carries both the result and any per-item error.
// This wrapper pattern lets ParallelMapSeq (which has no error channel) surface
// per-item failures to the consumer without panicking.
type EnrichedUser struct {
	ID      int
	Name    string
	Country string
	Score   float64
	Err     error
}

// mockAPICall simulates an external HTTP enrichment call with variable latency
// and a small random failure rate.
func mockAPICall(id int) EnrichedUser {
	rng := rand.New(rand.NewSource(int64(id * 31337)))
	time.Sleep(time.Duration(3+rng.Intn(4)) * time.Millisecond) // 3–6 ms
	if rng.Float64() < 0.02 {
		return EnrichedUser{ID: id, Err: fmt.Errorf("API timeout for user %d", id)}
	}
	return EnrichedUser{
		ID:      id,
		Name:    fmt.Sprintf("User-%d", id),
		Country: countries[id%len(countries)],
		Score:   rng.Float64() * 100,
	}
}

func memMB() float64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}

func makeIDs(n int) []int {
	ids := make([]int, n)
	for i := range ids {
		ids[i] = i + 1
	}
	return ids
}

func main() {
	ctx := context.Background()
	ids := makeIDs(TOTAL_USERS)

	fmt.Println("=== Parallel API Enrichment — vortex example ===")
	fmt.Printf("users: %d   workers: %d\n", TOTAL_USERS, WORKERS)
	fmt.Printf("mock latency per call: 3–6 ms   failure rate: ~2%%\n\n")

	// ── Sequential ───────────────────────────────────────────────────────────
	fmt.Println("--- Sequential (single goroutine) ---")
	memBefore := memMB()
	seqStart := time.Now()

	var seqOK, seqFailed int
	for _, id := range ids {
		r := mockAPICall(id)
		if r.Err != nil {
			seqFailed++
		} else {
			seqOK++
		}
	}

	seqDuration := time.Since(seqStart)
	fmt.Printf("duration  : %v\n", seqDuration)
	fmt.Printf("ok/failed : %d / %d\n", seqOK, seqFailed)
	fmt.Printf("mem delta : %+.1f MB\n\n", memMB()-memBefore)

	// ── Parallel ─────────────────────────────────────────────────────────────
	fmt.Printf("--- Parallel (%d workers, parallel.ParallelMapSeq) ---\n", WORKERS)
	memBefore = memMB()
	parStart := time.Now()

	// ParallelMapSeq fans out to WORKERS goroutines. Results arrive unordered
	// (fastest worker wins), which is ideal for I/O-bound tasks where order
	// doesn't matter.
	enriched := parallel.ParallelMapSeq(ctx, slices.Values(ids), mockAPICall, WORKERS)

	var parOK, parFailed int
	for r := range enriched {
		if r.Err != nil {
			parFailed++
		} else {
			parOK++
		}
	}

	parDuration := time.Since(parStart)
	fmt.Printf("duration  : %v\n", parDuration)
	fmt.Printf("ok/failed : %d / %d\n", parOK, parFailed)
	fmt.Printf("mem delta : %+.1f MB\n\n", memMB()-memBefore)

	// ── Summary ───────────────────────────────────────────────────────────────
	speedup := float64(seqDuration) / float64(parDuration)
	fmt.Printf("=== Summary ===\n")
	fmt.Printf("sequential : %v\n", seqDuration)
	fmt.Printf("parallel   : %v\n", parDuration)
	fmt.Printf("speedup    : %.1fx\n", speedup)
	fmt.Printf("\nNote: scale to 50 000 users — same speedup, flat memory throughout.\n")
}
