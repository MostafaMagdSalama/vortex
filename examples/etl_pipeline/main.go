package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

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

func main() {
	ctx := context.Background()

	// 1. Setup a dummy SQLite database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal("failed to open db:", err)
	}
	defer db.Close()
	setupDB(db)

	// 2. Open an output CSV file
	out, err := os.Create("users.csv")
	if err != nil {
		log.Fatal("failed to create output file:", err)
	}
	defer out.Close()
	fmt.Fprintln(out, "id,name_upper,domain")

	// --- ETL PIPELINE ---

	// SOURCE: Lazy scan from database
	users := sources.DBRows(ctx, db, "SELECT id, name, email FROM users", func(rows *sql.Rows) (User, error) {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Email)
		return u, err
	})

	// Validate filters and drops the errors out of iter.Seq2 into iter.Seq
	validUsers := func(yield func(User) bool) {
		for u, err := range users {
			if err != nil {
				log.Println("Skipping bad row:", err)
				continue
			}
			if !yield(u) {
				return
			}
		}
	}

	// TRANSFORM: Convert to uppercase and extract domain (safely)
	// We use parallel map to simulate a heavy transformation (like an external API enrichment)
	transformed := parallel.OrderedParallelMap(ctx, validUsers, func(u User) TransformedUser {
		domain := ""
		for i, c := range u.Email {
			if c == '@' {
				domain = u.Email[i+1:]
				break
			}
		}

		// Convert string to uppercase without strings.ToUpper for simplicity here
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
	}, 4) // 4 workers

	// LOAD: Write to CSV manually instead of Drain to avoid import cycles / iterx issues right here
	for tu := range transformed {
		line := fmt.Sprintf("%d,%s,%s\n", tu.ID, tu.NameUpper, tu.Domain)
		if _, err := out.WriteString(line); err != nil {
			log.Fatal("pipeline failed:", err)
		}
	}

	fmt.Println("ETL Pipeline completed successfully! Wrote outputs to users.csv")
	
	// Print file contents
	b, _ := os.ReadFile("users.csv")
	fmt.Println(string(b))
	os.Remove("users.csv")
}

func setupDB(db *sql.DB) {
	_, _ = db.Exec(`CREATE TABLE users (id INTEGER, name TEXT, email TEXT)`)
	_, _ = db.Exec(`INSERT INTO users VALUES (1, 'Alice', 'alice@example.com')`)
	_, _ = db.Exec(`INSERT INTO users VALUES (2, 'Bob', 'bob@vortex.dev')`)
	_, _ = db.Exec(`INSERT INTO users VALUES (3, 'Charlie', 'charlie@vortex.dev')`)
}
