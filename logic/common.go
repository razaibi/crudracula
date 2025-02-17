package logic

import (
	"crudracula/models"
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var jwtSecret []byte

func init() {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist in production
		// as env vars might be set through other means
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// Get the JWT secret from environment variable
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// In a production environment, you might want to
		// handle this differently, e.g., exit the program
		fmt.Println("Warning: JWT_SECRET not set, using default value")
		secret = "your-secret-key"
	}
	jwtSecret = []byte(secret)
}

func verifyToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*models.Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, errors.New("invalid token")
}
