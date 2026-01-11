package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"backend-go-ticketing-gamify/internal/seeders"
)

func main() {
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is empty")
	}
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	opt := seeders.Options{
		Users:    parseEnvInt("SEED_USERS", 10),
		Projects: parseEnvInt("SEED_PROJECTS", 3),
		Tickets:  parseEnvInt("SEED_TICKETS", 25),
		Comments: parseEnvInt("SEED_COMMENTS", 40),
		Preset:   "realistic", // Force realistic for now
	}

	// Always clean for this request
	if err := seeders.CleanAll(ctx, pool); err != nil {
		log.Printf("clean warning: %v", err)
	}

	if err := seeders.SeedAll(ctx, pool, opt); err != nil {
		log.Fatalf("seed failed: %v", err)
	}
}

func parseEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
