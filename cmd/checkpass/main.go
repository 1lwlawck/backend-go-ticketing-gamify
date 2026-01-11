package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL empty")
	}
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var id, name, username, role string
	// Search for 'Deny'
	err = db.QueryRow(context.Background(), "SELECT id, name, username, role FROM users WHERE name ILIKE '%Deny%'").Scan(&id, &name, &username, &role)
	if err != nil {
		fmt.Printf("User 'Deny' not found in DB: %v\n", err)
	} else {
		fmt.Printf("User Found: %s | Role: %s | ID: %s\n", name, role, id)

		// Check project memberships
		rows, _ := db.Query(context.Background(), "SELECT project_id, member_role FROM project_members WHERE user_id = $1", id)
		defer rows.Close()
		fmt.Println("--- Project Memberships ---")
		count := 0
		for rows.Next() {
			var pid, mrole string
			rows.Scan(&pid, &mrole)
			fmt.Printf("Project: %s | Role: %s\n", pid, mrole)
			count++
		}
		if count == 0 {
			fmt.Println("No project memberships found.")
		}
	}
}
