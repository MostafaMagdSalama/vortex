package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/MostafaMagdSalama/vortex/parallel"
	"github.com/MostafaMagdSalama/vortex/sources"

	_ "modernc.org/sqlite"
)

type User struct {
	ID    int
	Name  string
	Email string
}

type TransformedUser struct {
	ID        int
	NameUpper string
	Domain    string
}

func memStats(label string) {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf(
		"[MEM] %-45s alloc=%6d KB   heapInUse=%6d KB   heapObjects=%6d   sys=%6d KB\n",
		label,
		m.Alloc/1024,
		m.HeapInuse/1024,
		m.HeapObjects,
		m.Sys/1024,
	)
}

func main() {
	start := time.Now()

	fmt.Println("=== ETL Pipeline — 1,000,000 rows ===")
	fmt.Println()

	ctx := context.Background()

	memStats("start")

	// 1. Setup database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal("failed to open db:", err)
	}
	defer db.Close()

	fmt.Println("inserting 1,000,000 rows...")
	setupDB(db, 1_000_000)
	fmt.Println("done inserting")
	fmt.Println()

	memStats("after db setup (1,000,000 rows)")

	// 2. Open output CSV
	out, err := os.Create("users.csv")
	if err != nil {
		log.Fatal("failed to create output file:", err)
	}
	defer out.Close()
	fmt.Fprintln(out, "id,name_upper,domain")

	memStats("after file create")
	fmt.Println()
	fmt.Println("--- pipeline definition (all lazy, nothing runs yet) ---")
	fmt.Println()

	// SOURCE
	users := sources.DBRows(ctx, db, "SELECT id, name, email FROM users", func(rows *sql.Rows) (User, error) {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Email)
		return u, err
	})

	memStats("after DBRows defined     (lazy)")

	// TRANSFORM
	transformed := parallel.OrderedParallelMap(ctx, users, func(u User) TransformedUser {
		domain := ""
		for i, c := range u.Email {
			if c == '@' {
				domain = u.Email[i+1:]
				break
			}
		}

		upperName := ""
		for _, c := range u.Name {
			if c >= 'a' && c <= 'z' {
				upperName += string(c - 32)
			} else {
				upperName += string(c)
			}
		}

		return TransformedUser{
			ID:        u.ID,
			NameUpper: upperName,
			Domain:    domain,
		}
	}, 4)

	memStats("after OrderedParallelMap (lazy)")

	fmt.Println()
	fmt.Println("--- pipeline execution (for range drives everything) ---")
	fmt.Println()

	// LOAD
	rowCount := 0
	errCount := 0
	logEvery := 100_000

	for tu, err := range transformed {
		if err != nil {
			errCount++
			log.Printf("[WARN] skipping bad row: %v\n", err)
			continue
		}
		line := fmt.Sprintf("%d,%s,%s\n", tu.ID, tu.NameUpper, tu.Domain)
		if _, err := out.WriteString(line); err != nil {
			log.Fatal("pipeline failed:", err)
		}
		rowCount++

		// progress + memory every 100,000 rows
		if rowCount%logEvery == 0 {
			memStats(fmt.Sprintf("progress: %7d rows written", rowCount))
		}
	}

	fmt.Println()
	memStats("after pipeline completed")

	fmt.Println()
	fmt.Println("--- cleaning up ---")
	fmt.Println()
	os.Remove("users.csv")

	memStats("after cleanup")

	fmt.Println()
	fmt.Println("=== summary ===")
	fmt.Printf("rows written : %d\n", rowCount)
	fmt.Printf("rows skipped : %d\n", errCount)
	fmt.Printf("duration     : %s\n", time.Since(start))
	fmt.Println()
}

func setupDB(db *sql.DB, count int) {
	_, _ = db.Exec(`CREATE TABLE users (id INTEGER, name TEXT, email TEXT)`)

	// batch insert using transactions — much faster than one insert per row
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("failed to begin transaction:", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO users VALUES (?, ?, ?)`)
	if err != nil {
		log.Fatal("failed to prepare statement:", err)
	}
	defer stmt.Close()

	domains := []string{"example.com", "vortex.dev", "gmail.com", "yahoo.com", "outlook.com"}
	names := []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry"}

	for i := 1; i <= count; i++ {
		name := names[i%len(names)]
		domain := domains[i%len(domains)]
		email := fmt.Sprintf("user%d@%s", i, domain)
		if _, err := stmt.Exec(i, name, email); err != nil {
			log.Fatal("failed to insert row:", err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal("failed to commit:", err)
	}
}