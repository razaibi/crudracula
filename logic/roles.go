package logic

import (
	"crudracula/dal"
	"crudracula/models"
	"database/sql"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func GetRolesPage(c *fiber.Ctx) error {
	return c.Render("role_manager", fiber.Map{
		"title": "Role Management - Crudracula",
	})
}

func GetRoles(c *fiber.Ctx) error {
	rows, err := dal.DB.Query(`
		SELECT r.id, r.name, r.description
		FROM roles r
		ORDER BY r.name`)
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Msg("Failed to fetch roles")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			fmt.Println(err)
			log.Error().Err(err).Msg("Failed to scan role")
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}

		// Fetch permissions for this role
		permRows, err := dal.DB.Query(`
			SELECT p.id, p.name, p.description
			FROM permissions p
			JOIN role_permissions rp ON rp.permission_id = p.id
			WHERE rp.role_id = ?
			ORDER BY p.name`, role.ID)
		if err != nil {
			fmt.Println(err)
			log.Error().Err(err).Msg("Failed to fetch role permissions")
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}
		defer permRows.Close()

		for permRows.Next() {
			var perm models.Permission
			if err := permRows.Scan(&perm.ID, &perm.Name, &perm.Description); err != nil {
				fmt.Println(err)
				log.Error().Err(err).Msg("Failed to scan permission")
				return c.Status(500).JSON(fiber.Map{"error": "Database error"})
			}
			role.Permissions = append(role.Permissions, perm)
		}

		roles = append(roles, role)
	}

	return c.JSON(roles)
}

func GetRole(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role ID"})
	}

	var role models.Role
	err = dal.DB.QueryRow(`
		SELECT id, name, description
		FROM roles
		WHERE id = ?`, id).Scan(&role.ID, &role.Name, &role.Description)
	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"error": "Role not found"})
	}
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Int("id", id).Msg("Failed to fetch role")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	// Fetch permissions for this role
	rows, err := dal.DB.Query(`
		SELECT p.id, p.name, p.description
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = ?
		ORDER BY p.name`, id)
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Int("id", id).Msg("Failed to fetch role permissions")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	for rows.Next() {
		var perm models.Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Description); err != nil {
			fmt.Println(err)
			log.Error().Err(err).Msg("Failed to scan permission")
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}
		role.Permissions = append(role.Permissions, perm)
	}

	return c.JSON(role)
}

func CreateRole(c *fiber.Ctx) error {
	var roleRequest models.RoleRequest
	if err := c.BodyParser(&roleRequest); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if roleRequest.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Role name is required"})
	}

	// Start transaction
	tx, err := dal.DB.Begin()
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Msg("Failed to start transaction")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback()

	// Insert role
	result, err := tx.Exec(`
		INSERT INTO roles (name, description)
		VALUES (?, ?)`, roleRequest.Name, roleRequest.Description)
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Str("name", roleRequest.Name).Msg("Failed to create role")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	roleID, err := result.LastInsertId()
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Msg("Failed to get last insert ID")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	// Insert role permissions
	for _, permID := range roleRequest.Permissions {
		_, err = tx.Exec(`
			INSERT INTO role_permissions (role_id, permission_id)
			VALUES (?, ?)`, roleID, permID)
		if err != nil {
			fmt.Println(err)
			log.Error().Err(err).Int64("roleId", roleID).Int("permId", permID).
				Msg("Failed to insert role permission")
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
		log.Error().Err(err).Msg("Failed to commit transaction")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	return c.Status(201).JSON(fiber.Map{
		"id":   roleID,
		"name": roleRequest.Name,
	})
}

func UpdateRole(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role ID"})
	}

	// Don't allow updating the admin role
	if id == 1 {
		return c.Status(403).JSON(fiber.Map{"error": "Cannot modify admin role"})
	}

	var roleRequest models.RoleRequest
	if err := c.BodyParser(&roleRequest); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if roleRequest.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Role name is required"})
	}

	// Start transaction
	tx, err := dal.DB.Begin()
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Msg("Failed to start transaction")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback()

	// Update role
	result, err := tx.Exec(`
		UPDATE roles
		SET name = ?, description = ?
		WHERE id = ?`, roleRequest.Name, roleRequest.Description, id)
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Int("id", id).Msg("Failed to update role")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Msg("Failed to get rows affected")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Role not found"})
	}

	// Delete existing permissions
	_, err = tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", id)
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Int("id", id).Msg("Failed to delete role permissions")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	// Insert new permissions
	for _, permID := range roleRequest.Permissions {
		_, err = tx.Exec(`
			INSERT INTO role_permissions (role_id, permission_id)
			VALUES (?, ?)`, id, permID)
		if err != nil {
			fmt.Println(err)
			log.Error().Err(err).Int("roleId", id).Int("permId", permID).
				Msg("Failed to insert role permission")
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
		log.Error().Err(err).Msg("Failed to commit transaction")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	return c.JSON(fiber.Map{
		"id":   id,
		"name": roleRequest.Name,
	})
}

func DeleteRole(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role ID"})
	}

	// Don't allow deleting the admin role
	if id == 1 {
		return c.Status(403).JSON(fiber.Map{"error": "Cannot delete admin role"})
	}

	// Check if any users are using this role
	var userCount int
	err = dal.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role_id = ?", id).Scan(&userCount)
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Int("id", id).Msg("Failed to check role users")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	if userCount > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot delete role while it is assigned to users"})
	}

	// Start transaction
	tx, err := dal.DB.Begin()
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Msg("Failed to start transaction")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback()

	// Delete role permissions
	_, err = tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", id)
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Int("id", id).Msg("Failed to delete role permissions")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	// Delete role
	result, err := tx.Exec("DELETE FROM roles WHERE id = ?", id)
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Int("id", id).Msg("Failed to delete role")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Msg("Failed to get rows affected")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Role not found"})
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
		log.Error().Err(err).Msg("Failed to commit transaction")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	return c.SendStatus(204)
}

func GetPermissions(c *fiber.Ctx) error {
	rows, err := dal.DB.Query("SELECT id, name, description FROM permissions ORDER BY name")
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Msg("Failed to fetch permissions")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Description); err != nil {
			fmt.Println(err)
			log.Error().Err(err).Msg("Failed to scan permission")
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}
		permissions = append(permissions, perm)
	}

	return c.JSON(permissions)
}

func CheckPermission(c *fiber.Ctx) error {
	userID, err := GetUserIDFromToken(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	permName := c.Params("permission")
	if permName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Permission name is required"})
	}

	hasAccess, err := hasPermission(userID, permName)
	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Int("userId", userID).Str("permission", permName).
			Msg("Failed to check permission")
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	if !hasAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Permission denied"})
	}

	return c.SendStatus(200)
}

// Helper function to check if a user has a specific permission
func hasPermission(userID int, permissionName string) (bool, error) {
	var exists bool
	err := dal.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM users u
			JOIN roles r ON r.id = u.role_id
			JOIN role_permissions rp ON rp.role_id = r.id
			JOIN permissions p ON p.id = rp.permission_id
			WHERE u.id = ? AND p.name = ?
		)`, userID, permissionName).Scan(&exists)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return exists, nil
}
