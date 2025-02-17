package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Add this middleware function to main.go:
func AuthMiddleware(c *fiber.Ctx) error {
	// Skip auth for public routes
	if isPublicPath(c.Path()) {
		return c.Next()
	}

	// Get token from header
	auth := c.Get("Authorization")
	if auth == "" {
		return c.Status(401).JSON(fiber.Map{"error": "No authorization header"})
	}

	// Remove "Bearer " prefix if present
	token := auth
	if strings.HasPrefix(auth, "Bearer ") {
		token = auth[7:]
	}

	// TODO: Implement proper JWT validation
	// For now, just check if token exists
	if token == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
	}

	return c.Next()
}

func isPublicPath(path string) bool {
	publicPaths := []string{
		"/api/login",
		"/api/signup",
		"/api/request-reset",
		"/api/reset-password",
		"/login.html",
		"/signup.html",
		"/reset-password.html",
	}

	for _, pp := range publicPaths {
		if strings.HasPrefix(path, pp) {
			return true
		}
	}

	return false
}
