package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/MostafaMagdSalama/vortex/iterx"
	"github.com/MostafaMagdSalama/vortex/sources"
	"github.com/xuri/excelize/v2"
	_ "modernc.org/sqlite"
)

const TOTAL_ROWS = 1_000_000

type Product struct {
	ID       int
	Name     string
	Category string
	Price    float64
	Stock    int
}

func scanProduct(rows *sql.Rows) (Product, error) {
	var p Product
	return p, rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.Stock)
}

func memMB() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}

func seedDB(db *sql.DB) {
	fmt.Printf("Seeding %d rows...\n", TOTAL_ROWS)
	db.Exec(`CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY, name TEXT, category TEXT, price REAL, stock INTEGER
	)`)
	categories := []string{"Electronics", "Furniture", "Stationery", "Clothing", "Food"}
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare(`INSERT INTO products VALUES (?,?,?,?,?)`)
	for i := 1; i <= TOTAL_ROWS; i++ {
		stmt.Exec(i, fmt.Sprintf("Product-%d", i), categories[i%len(categories)], float64(i%500)+0.99, i%300)
		if i%100_000 == 0 {
			tx.Commit()
			tx, _ = db.Begin()
			stmt, _ = tx.Prepare(`INSERT INTO products VALUES (?,?,?,?,?)`)
			fmt.Printf("  inserted %d rows...\n", i)
		}
	}
	tx.Commit()
	stmt.Close()
	fmt.Println("Seeding done.")
}

// APPROACH 1: vortex lazy pipeline + StreamWriter (truly constant memory)
func runVortex(db *sql.DB) {
	fmt.Println("==============================================")
	fmt.Println("  APPROACH 1: vortex lazy + StreamWriter")
	fmt.Println("  (rows streamed from DB, flushed to disk)")
	fmt.Println("==============================================")

	runtime.GC()
	memBefore := memMB()
	start := time.Now()

	ctx := context.Background()

	seq := sources.DBRows(ctx, db,
		`SELECT id, name, category, price, stock FROM products`,
		scanProduct,
	)
	filtered := iterx.Filter(ctx, seq, func(p Product) bool {
		return p.Stock > 0
	})

	f := excelize.NewFile()
	sheet := "Products"
	f.SetSheetName("Sheet1", sheet)

	// StreamWriter flushes rows directly to disk — never keeps them in RAM
	sw, err := f.NewStreamWriter(sheet)
	if err != nil {
		log.Fatal(err)
	}

	sw.SetRow("A1", []any{
		excelize.Cell{Value: "ID"},
		excelize.Cell{Value: "Name"},
		excelize.Cell{Value: "Category"},
		excelize.Cell{Value: "Price ($)"},
		excelize.Cell{Value: "Stock"},
	})

	rowNum := 2
	for p, err := range filtered {
		if err != nil {
			log.Fatal(err)
		}
		cell, _ := excelize.CoordinatesToCellName(1, rowNum)
		sw.SetRow(cell, []any{p.ID, p.Name, p.Category, p.Price, p.Stock})
		rowNum++
	}

	sw.Flush()
	f.SaveAs("vortex_output.xlsx")

	elapsed := time.Since(start)
	memAfter := memMB()

	fmt.Printf("  rows written : %d\n", rowNum-2)
	fmt.Printf("  time         : %s\n", elapsed)
	fmt.Printf("  mem before   : %.2f MB\n", memBefore)
	fmt.Printf("  mem after    : %.2f MB\n", memAfter)
	fmt.Printf("  mem delta    : +%.2f MB\n\n", memAfter-memBefore)
}

// APPROACH 2: eager — loads everything into RAM first
func runEager(db *sql.DB) {
	fmt.Println("==============================================")
	fmt.Println("  APPROACH 2: eager (load all rows into RAM)")
	fmt.Println("==============================================")

	runtime.GC()
	memBefore := memMB()
	start := time.Now()

	rows, err := db.Query(`SELECT id, name, category, price, stock FROM products`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var all []Product
	for rows.Next() {
		var p Product
		rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.Stock)
		all = append(all, p)
	}
	fmt.Printf("  mem after loading all rows : %.2f MB\n", memMB())

	var filtered []Product
	for _, p := range all {
		if p.Stock > 0 {
			filtered = append(filtered, p)
		}
	}
	fmt.Printf("  mem after filtering        : %.2f MB\n", memMB())

	f := excelize.NewFile()
	sheet := "Products"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"ID", "Name", "Category", "Price ($)", "Stock"}
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, p := range filtered {
		rowNum := i + 2
		values := []any{p.ID, p.Name, p.Category, p.Price, p.Stock}
		for col, v := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, rowNum)
			f.SetCellValue(sheet, cell, v)
		}
	}
	f.SaveAs("eager_output.xlsx")

	elapsed := time.Since(start)
	memAfter := memMB()

	fmt.Printf("  rows written : %d\n", len(filtered))
	fmt.Printf("  time         : %s\n", elapsed)
	fmt.Printf("  mem before   : %.2f MB\n", memBefore)
	fmt.Printf("  mem after    : %.2f MB\n", memAfter)
	fmt.Printf("  mem delta    : +%.2f MB\n\n", memAfter-memBefore)
}

func main() {
	db, err := sql.Open("sqlite", "file:bench.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	seedDB(db)
	runVortex(db)
	runEager(db)

	fmt.Println("Done. Check vortex_output.xlsx and eager_output.xlsx")
}