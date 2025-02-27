package logic

import (
	"crudracula/models"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var jwtSecret []byte

func init() {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist in production
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

// GetUserIDFromToken extracts the user ID from the JWT token in the Authorization header
func GetUserIDFromToken(c *fiber.Ctx) (int, error) {
	auth := c.Get("Authorization")
	if auth == "" {
		return 0, errors.New("no authorization header")
	}

	// Remove Bearer prefix if present
	token := auth
	if strings.HasPrefix(auth, "Bearer ") {
		token = auth[7:]
	}

	// Verify and parse the token
	userID, err := VerifyToken(token)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// VerifyToken validates a JWT token and returns the user ID
func VerifyToken(tokenString string) (int, error) {
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

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID int) (string, error) {
	// Create the Claims
	claims := &models.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
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
