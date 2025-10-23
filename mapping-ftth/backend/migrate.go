package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

func main() {
	fmt.Println("=== Database Migration: Add Profile Column ===")
	
	db, err := sql.Open("sqlite", "./ftth.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	fmt.Println("Adding 'profile' column to pelanggan table...")
	
	_, err = db.Exec(`ALTER TABLE pelanggan ADD COLUMN profile TEXT`)
	if err != nil {
		// Check if error is because column already exists
		if err.Error() == "duplicate column name: profile" {
			fmt.Println("✓ Column 'profile' already exists - skipping")
		} else {
			fmt.Println("✗ Error:", err)
		}
	} else {
		fmt.Println("✓ Column 'profile' added successfully!")
	}

	// Verify column exists
	rows, err := db.Query(`PRAGMA table_info(pelanggan)`)
	if err != nil {
		fmt.Println("Error checking table info:", err)
		return
	}
	defer rows.Close()

	fmt.Println("\nVerifying table structure:")
	hasProfile := false
	for rows.Next() {
		var cid int
		var name, type_ string
		var notnull, dflt_value, pk interface{}
		rows.Scan(&cid, &name, &type_, &notnull, &dflt_value, &pk)
		if name == "profile" {
			hasProfile = true
			fmt.Printf("  ✓ Column: %s (%s)\n", name, type_)
		}
	}

	if hasProfile {
		fmt.Println("\n✓✓✓ Migration successful! Column 'profile' is ready.")
	} else {
		fmt.Println("\n✗✗✗ Migration failed! Column 'profile' not found.")
	}
}
