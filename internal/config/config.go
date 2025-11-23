package config

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultHTTPPort = "8080"
	defaultEnv      = "development"
	defaultShutdown = 10 * time.Second
)

// Config captures the runtime configuration for the API server.
type Config struct {
	Env             string
	HTTPPort        string
	DatabaseURL     string
	JWTSecret       string
	RateLimitPerMin int
	APIKeyRateLimit int
	RateLimitWindow time.Duration
	APIKey          string
	APIKeyHeader    string
	ShutdownTimeout time.Duration
}

var (
	cfg  Config
	once sync.Once
)

// Load builds a Config value using environment variables and sensible defaults.
func Load() Config {
	once.Do(func() {
		_ = godotenv.Load()
		port := getEnv("PORT", defaultHTTPPort)

		cfg = Config{
			Env:             getEnv("APP_ENV", defaultEnv),
			HTTPPort:        port,
			DatabaseURL:     os.Getenv("DATABASE_URL"),
			JWTSecret:       os.Getenv("JWT_SECRET"),
			RateLimitPerMin: getInt("RATE_LIMIT_PER_MIN", 120),
			APIKeyRateLimit: getInt("API_KEY_RATE_LIMIT_PER_MIN", 300),
			RateLimitWindow: getDuration("RATE_LIMIT_WINDOW", time.Minute),
			APIKey:          os.Getenv("API_KEY"),
			APIKeyHeader:    getEnv("API_KEY_HEADER", "X-API-Key"),
			ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", defaultShutdown),
		}

		if cfg.DatabaseURL == "" {
			log.Println("config: DATABASE_URL not set; DB calls will fail")
		}
		if cfg.JWTSecret == "" {
			log.Println("config: JWT_SECRET not set; auth will fail")
		}
	})

	return cfg
}

// Addr returns the listen address for the HTTP server.
func (c Config) Addr() string {
	return fmt.Sprintf(":%s", c.HTTPPort)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
		log.Printf("config: invalid duration for %s, using fallback %v", key, fallback)
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		var parsed int
		if _, err := fmt.Sscanf(value, "%d", &parsed); err == nil {
			return parsed
		}
		log.Printf("config: invalid int for %s, using fallback %d", key, fallback)
	}
	return fallback
}
