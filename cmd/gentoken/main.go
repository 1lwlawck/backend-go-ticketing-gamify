package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "supersecretkey" // Fallback to default in .env usually
	}

	// Adit Admin ID from sample_data.go
	uid := "6b062d3e-1c3f-4a6f-936e-73f8c4e78a60"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  uid,
		"name": "Adit Santoso",
		"role": "admin",
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})

	t, _ := token.SignedString([]byte(secret))
	fmt.Print(t)
}
