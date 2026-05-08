package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/MostafaMagdSalama/vortex/iterx"
	"github.com/MostafaMagdSalama/vortex/sources"
)

const TOTAL_ENTRIES = 500_000

var levels = []string{
	"DEBUG", "DEBUG",
	"INFO", "INFO", "INFO", "INFO", "INFO",
	"WARN", "WARN",
	"ERROR",
}

var services = []string{"auth", "billing", "api-gateway", "user-service", "payment", "notifications"}

var messages = map[string][]string{
	"DEBUG": {"cache hit", "query plan computed", "config reloaded", "connection established"},
	"INFO":  {"request handled", "user logged in", "order created", "payment processed"},
	"WARN":  {"retry attempt #2", "slow query detected", "rate limit at 80%", "deprecated API call"},
	"ERROR": {"connection refused", "timeout exceeded", "nil pointer dereference", "database write failed"},
}

type LogEntry struct {
	Timestamp string `json:"ts"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Message   string `json:"msg"`
	ReqID     string `json:"req_id,omitempty"`
}

type Alert struct {
	Timestamp string
	Level     string
	Service   string
	Message   string
}

func memStats(label string) {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("[MEM] %-32s alloc=%6d KB  heapInUse=%6d KB  heapObjects=%6d\n",
		label, m.Alloc/1024, m.HeapInuse/1024, m.HeapObjects)
}

func generateLogFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	rng := rand.New(rand.NewSource(42))
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < TOTAL_ENTRIES; i++ {
		level := levels[rng.Intn(len(levels))]
		svc := services[rng.Intn(len(services))]
		msgs := messages[level]
		entry := LogEntry{
			Timestamp: base.Add(time.Duration(i) * time.Second).Format(time.RFC3339),
			Level:     level,
			Service:   svc,
			Message:   msgs[rng.Intn(len(msgs))],
		}
		if level == "ERROR" || level == "WARN" {
			entry.ReqID = fmt.Sprintf("req-%08d", rng.Intn(100_000_000))
		}
		if err := enc.Encode(entry); err != nil {
			return err
		}
		if i > 0 && i%100_000 == 0 {
			fmt.Printf("  generated %d / %d entries...\n", i, TOTAL_ENTRIES)
		}
	}
	return nil
}

func main() {
	ctx := context.Background()

	fmt.Println("=== JSONL Log Analysis — vortex example ===")
	fmt.Printf("entries: %d\n\n", TOTAL_ENTRIES)

	// ── Generate input file ──────────────────────────────────────────────────
	inputPath := "logs.jsonl"
	fmt.Printf("generating %s...\n", inputPath)
	start := time.Now()
	if err := generateLogFile(inputPath); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("done in %v\n", time.Since(start))
	defer os.Remove(inputPath)
	memStats("after generate")

	// ── Lazy approach (vortex) ───────────────────────────────────────────────
	fmt.Println("\n--- Lazy (vortex): stream → filter → map → drain ---")

	alertsOut, err := os.Create("alerts.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove("alerts.txt")

	lazyStart := time.Now()

	seq := sources.JSONLinesFile[LogEntry](ctx, inputPath)
	filtered := iterx.Filter(ctx, seq, func(e LogEntry) bool {
		return e.Level == "ERROR" || e.Level == "WARN"
	})
	alerts := iterx.Map(ctx, filtered, func(e LogEntry) Alert {
		return Alert{e.Timestamp, e.Level, e.Service, e.Message}
	})

	var lazyCount int
	if err := iterx.Drain(ctx, alerts, func(a Alert) error {
		lazyCount++
		_, err := fmt.Fprintf(alertsOut, "%s [%-5s] %-16s %s\n", a.Timestamp, a.Level, a.Service, a.Message)
		return err
	}); err != nil {
		log.Fatalf("drain error: %v", err)
	}
	alertsOut.Close()

	lazyDuration := time.Since(lazyStart)
	memStats(fmt.Sprintf("after lazy drain (%d alerts)", lazyCount))
	fmt.Printf("duration  : %v\n", lazyDuration)

	// ── Eager approach (load everything first) ───────────────────────────────
	fmt.Println("\n--- Eager: load all → filter → write ---")

	eagerStart := time.Now()

	rawFile, err := os.Open(inputPath)
	if err != nil {
		log.Fatal(err)
	}
	dec := json.NewDecoder(rawFile)
	var all []LogEntry
	for dec.More() {
		var e LogEntry
		if err := dec.Decode(&e); err != nil {
			log.Fatal(err)
		}
		all = append(all, e)
	}
	rawFile.Close()

	var eagerAlerts []Alert
	for _, e := range all {
		if e.Level == "ERROR" || e.Level == "WARN" {
			eagerAlerts = append(eagerAlerts, Alert{e.Timestamp, e.Level, e.Service, e.Message})
		}
	}
	eagerCount := len(eagerAlerts)

	// measure while both slices are still live — this is the peak memory moment
	memStats(fmt.Sprintf("after eager load  (%d alerts)", eagerCount))

	eagerOut, err := os.Create("alerts_eager.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove("alerts_eager.txt")
	for _, a := range eagerAlerts {
		fmt.Fprintf(eagerOut, "%s [%-5s] %-16s %s\n", a.Timestamp, a.Level, a.Service, a.Message)
	}
	eagerOut.Close()

	eagerDuration := time.Since(eagerStart)
	fmt.Printf("duration  : %v\n", eagerDuration)

	// ── Summary ──────────────────────────────────────────────────────────────
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("alerts found : %d (lazy) / %d (eager)\n", lazyCount, eagerCount)
	fmt.Printf("lazy         : %v\n", lazyDuration)
	fmt.Printf("eager        : %v\n", eagerDuration)
}
