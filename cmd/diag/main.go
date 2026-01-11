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

	var count int
	db.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&count)
	fmt.Printf("Total users: %d\n", count)

	rows, _ := db.Query(context.Background(), "SELECT id, name, username, role FROM users")
	defer rows.Close()
	fmt.Println("--- User List ---")
	for rows.Next() {
		var id, name, username, role string
		rows.Scan(&id, &name, &username, &role)
		fmt.Printf("%s | %s (%s) [%s]\n", name, username, role, id)
	}
}
