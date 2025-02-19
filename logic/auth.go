package logic

import (
	"crudracula/dal"
	"crudracula/models"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func GetLoginPage(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{
		"title": "Login - Crudracula",
	})
}

func GetLogoutPage(c *fiber.Ctx) error {
	return c.Render("logout", fiber.Map{
		"title": "Login - Crudracula",
	})
}

func Signup(c *fiber.Ctx) error {
	req := new(models.SignupRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Validate email and password
	if !isValidEmail(req.Email) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid email format"})
	}
	if len(req.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 8 characters"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return c.Status(500).JSON(fiber.Map{"error": "Server error"})
	}

	// Check if user already exists
	var exists bool
	err = dal.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", req.Email).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Msg("Database error checking user existence")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	if exists {
		return c.Status(400).JSON(fiber.Map{"error": "Email already registered"})
	}

	// Create user
	result, err := dal.DB.Exec(
		"INSERT INTO users (email, password, created_at) VALUES (?, ?, ?)",
		req.Email, hashedPassword, time.Now(),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	id, _ := result.LastInsertId()
	return c.Status(201).JSON(fiber.Map{
		"id":    id,
		"email": req.Email,
	})
}

func Login(c *fiber.Ctx) error {
	req := new(models.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	var user models.User
	err := dal.DB.QueryRow(
		"SELECT id, email, password FROM users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Password)

	if err == sql.ErrNoRows {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	} else if err != nil {
		log.Error().Err(err).Msg("Database error during login")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Generate JWT token
	token, err := generateToken(user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token")
		return c.Status(500).JSON(fiber.Map{"error": "Server error"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}

func RequestPasswordReset(c *fiber.Ctx) error {
	req := new(models.ResetPasswordRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Generate reset token
	token, err := generateResetToken()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate reset token")
		return c.Status(500).JSON(fiber.Map{"error": "Server error"})
	}

	// Set token expiration
	expires := time.Now().Add(1 * time.Hour)

	// Update user with reset token
	result, err := dal.DB.Exec(
		"UPDATE users SET reset_token = ?, reset_token_expires = ? WHERE email = ?",
		token, expires, req.Email,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update reset token")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Email not found"})
	}

	// TODO: Send email with reset token
	// For now, just return token in response (development only)
	return c.JSON(fiber.Map{"token": token})
}

func ResetPassword(c *fiber.Ctx) error {
	req := new(models.UpdatePasswordRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if len(req.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 8 characters"})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return c.Status(500).JSON(fiber.Map{"error": "Server error"})
	}

	// Update password and clear reset token
	result, err := dal.DB.Exec(
		`UPDATE users 
		SET password = ?, reset_token = NULL, reset_token_expires = NULL 
		WHERE reset_token = ? AND reset_token_expires > ?`,
		hashedPassword, req.Token, time.Now(),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update password")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid or expired reset token"})
	}

	return c.SendStatus(200)
}

// Helper functions
func generateResetToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func isValidEmail(email string) bool {
	// Basic email validation - could be more sophisticated
	return len(email) > 3 && contains(email, "@") && contains(email, ".")
}

func contains(s string, substr string) bool {
	return strings.Contains(s, substr)
}

func generateToken(userID int) (string, error) {
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
