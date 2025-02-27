package middlewares

import (
	"crudracula/dal"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// RequirePermission creates a middleware that checks if the user has the required permission
func RequirePermission(permissionName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context (set by AuthMiddleware)
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
		}

		// Check if user has the permission through their role
		var roleID *int
		err := dal.DB.QueryRow("SELECT role_id FROM users WHERE id = ?", userID).Scan(&roleID)
		if err != nil {
			log.Error().Err(err).Int("userID", userID).Msg("Failed to get user role")
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user permissions")
		}

		if roleID == nil {
			return fiber.NewError(fiber.StatusForbidden, "No role assigned to user")
		}

		var hasPermission bool
		query := `
			SELECT EXISTS(
				SELECT 1 FROM role_permissions rp
				JOIN permissions p ON rp.permission_id = p.id
				WHERE rp.role_id = ? AND p.name = ?
			)
		`
		err = dal.DB.QueryRow(query, roleID, permissionName).Scan(&hasPermission)
		if err != nil {
			log.Error().Err(err).Int("userID", userID).Str("permission", permissionName).Msg("Failed to check permission")
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user permissions")
		}

		if !hasPermission {
			return fiber.NewError(fiber.StatusForbidden, "Permission denied")
		}

		// Continue to next middleware or handler
		return c.Next()
	}
}

// HasPermission checks if a user has a specific permission (utility function for other parts of the application)
func HasPermission(userID int, permission string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM role_permissions rp
			JOIN permissions p ON p.id = rp.permission_id
			JOIN users u ON u.role_id = rp.role_id
			WHERE u.id = ? AND p.name = ?
		)`
	err := dal.DB.QueryRow(query, userID, permission).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetUserPermissions returns all permissions for a user (utility function)
func GetUserPermissions(userID int) ([]string, error) {
	rows, err := dal.DB.Query(`
		SELECT DISTINCT p.name 
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN users u ON u.role_id = rp.role_id
		WHERE u.id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}
