package middlewares

import (
	"crudracula/logic" // Import the logic package to use the token verification
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware validates JWT tokens and extracts user information
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
	tokenString := auth
	if strings.HasPrefix(auth, "Bearer ") {
		tokenString = auth[7:]
	}

	// Use the VerifyToken function from logic package
	userID, err := logic.VerifyToken(tokenString)
	if err != nil {
		log.Error().Err(err).Str("token", tokenString).Msg("Invalid token")
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	// Store user ID in context
	c.Locals("userID", userID)
	return c.Next()
}

func isPublicPath(path string) bool {
	// Public paths same as before
	publicPaths := []string{
		"/api/login",
		"/api/signup",
		"/api/request-reset",
		"/api/reset-password",
		"/login",
		"/signup",
		"/reset-password",
	}

	for _, pp := range publicPaths {
		if strings.HasPrefix(path, pp) {
			return true
		}
	}

	return false
}
