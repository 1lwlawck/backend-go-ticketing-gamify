package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"backend-go-ticketing-gamify/internal/config"
	"backend-go-ticketing-gamify/internal/database"
	"backend-go-ticketing-gamify/internal/server"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to create db pool: %v", err)
	}
	defer pool.Close()

	srv := server.New(cfg, pool)

	if err := srv.Start(ctx); err != nil {
		log.Fatalf("fatal server error: %v", err)
	}
}
