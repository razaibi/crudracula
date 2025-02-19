package logic

import (
	"crudracula/dal"
	"crudracula/logger"
	"crudracula/models"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// getUserIDFromToken extracts the user ID from the JWT token
func getUserIDFromToken(c *fiber.Ctx) (int, error) {
	auth := c.Get("Authorization")
	if auth == "" {
		return 0, errors.New("no authorization header")
	}

	// Remove Bearer prefix
	token := auth
	if strings.HasPrefix(auth, "Bearer ") {
		token = auth[7:]
	}

	// Verify and parse the token
	userID, err := verifyToken(token)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func GetItemsPage(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"title": "Crudracula",
	})
}

func GetResetPasswordPage(c *fiber.Ctx) error {
	return c.Render("reset-password", fiber.Map{
		"title": "Crudracula",
	})
}

func GetSignUpPage(c *fiber.Ctx) error {
	return c.Render("signup", fiber.Map{
		"title": "Crudracula",
	})
}

func GetItems(c *fiber.Ctx) error {
	// Verify authentication
	userID, err := getUserIDFromToken(c)
	if err != nil {
		log.Debug().Err(err).Msg("Authentication failed")
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	perPage := 3
	offset := (page - 1) * perPage
	search := c.Query("search", "")

	log.Debug().
		Int("userId", userID).
		Int("page", page).
		Int("perPage", perPage).
		Str("search", search).
		Msg("Fetching items")

	var totalItems int
	var rows *sql.Rows

	if search != "" {
		err = dal.DB.QueryRow(`
			SELECT COUNT(*) FROM items 
			WHERE (name LIKE ? OR description LIKE ?) AND user_id = ?`,
			"%"+search+"%", "%"+search+"%", userID).Scan(&totalItems)
		if err != nil {
			logger.LogError(err, "Failed to count items with search", map[string]interface{}{
				"userId": userID,
				"search": search,
			})
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}

		rows, err = dal.DB.Query(`
            SELECT id, name, description 
            FROM items 
            WHERE (name LIKE ? OR description LIKE ?) AND user_id = ?
            ORDER BY id DESC
            LIMIT ? OFFSET ?`,
			"%"+search+"%", "%"+search+"%", userID, perPage, offset)
	} else {
		err = dal.DB.QueryRow("SELECT COUNT(*) FROM items WHERE user_id = ?",
			userID).Scan(&totalItems)
		if err != nil {
			logger.LogError(err, "Failed to count items", map[string]interface{}{
				"userId": userID,
			})
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}

		rows, err = dal.DB.Query(`
			SELECT id, name, description 
			FROM items 
			WHERE user_id = ? 
			ORDER BY id DESC 
			LIMIT ? OFFSET ?`,
			userID, perPage, offset)
	}

	if err != nil {
		logger.LogError(err, "Failed to fetch items", map[string]interface{}{
			"userId": userID,
			"page":   page,
			"search": search,
		})
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	totalPages := (totalItems + perPage - 1) / perPage

	var items []models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description); err != nil {
			logger.LogError(err, "Failed to scan item", map[string]interface{}{
				"userId": userID,
			})
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}
		items = append(items, item)
	}

	log.Info().
		Int("userId", userID).
		Int("totalItems", totalItems).
		Int("returnedItems", len(items)).
		Int("page", page).
		Int("totalPages", totalPages).
		Msg("Items retrieved successfully")

	return c.JSON(models.PaginatedResponse{
		Items:       items,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: page,
	})
}

func GetItem(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		log.Debug().Err(err).Msg("Authentication failed")
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		log.Debug().Err(err).Msg("Invalid ID parameter")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	log.Debug().Int("userId", userID).Int("id", id).Msg("Fetching single item")

	var item models.Item
	err = dal.DB.QueryRow(`
		SELECT id, name, description 
		FROM items 
		WHERE id = ? AND user_id = ?`,
		id, userID).Scan(&item.ID, &item.Name, &item.Description)

	if err == sql.ErrNoRows {
		log.Debug().Int("userId", userID).Int("id", id).Msg("Item not found")
		return c.Status(404).JSON(fiber.Map{"error": "Item not found"})
	} else if err != nil {
		logger.LogError(err, "Failed to fetch item", map[string]interface{}{
			"userId": userID,
			"id":     id,
		})
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	log.Info().Int("userId", userID).Int("id", id).Msg("Item retrieved successfully")
	return c.JSON(item)
}

func CreateItem(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		log.Debug().Err(err).Msg("Authentication failed")
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	item := new(models.Item)
	if err := c.BodyParser(item); err != nil {
		log.Debug().Err(err).Msg("Invalid item input")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	log.Debug().
		Int("userId", userID).
		Str("name", item.Name).
		Str("description", item.Description).
		Msg("Creating new item")

	result, err := dal.DB.Exec(`
		INSERT INTO items (name, description, user_id) 
		VALUES (?, ?, ?)`,
		item.Name, item.Description, userID)
	if err != nil {
		logger.LogError(err, "Failed to create item", map[string]interface{}{
			"userId":      userID,
			"name":        item.Name,
			"description": item.Description,
		})
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	id, _ := result.LastInsertId()
	item.ID = int(id)

	log.Info().
		Int("userId", userID).
		Int("id", item.ID).
		Str("name", item.Name).
		Msg("Item created successfully")

	return c.JSON(item)
}

func UpdateItem(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		log.Debug().Err(err).Msg("Authentication failed")
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		log.Debug().Err(err).Msg("Invalid ID parameter")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	item := new(models.Item)
	if err := c.BodyParser(item); err != nil {
		log.Debug().Err(err).Msg("Invalid item input")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	log.Debug().
		Int("userId", userID).
		Int("id", id).
		Str("name", item.Name).
		Str("description", item.Description).
		Msg("Updating item")

	result, err := dal.DB.Exec(`
		UPDATE items 
		SET name = ?, description = ? 
		WHERE id = ? AND user_id = ?`,
		item.Name, item.Description, id, userID)
	if err != nil {
		logger.LogError(err, "Failed to update item", map[string]interface{}{
			"userId":      userID,
			"id":          id,
			"name":        item.Name,
			"description": item.Description,
		})
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Debug().Int("userId", userID).Int("id", id).Msg("Item not found for update")
		return c.Status(404).JSON(fiber.Map{"error": "Item not found"})
	}

	item.ID = id
	log.Info().
		Int("userId", userID).
		Int("id", id).
		Str("name", item.Name).
		Msg("Item updated successfully")

	return c.JSON(item)
}

func DeleteItem(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		log.Debug().Err(err).Msg("Authentication failed")
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		log.Debug().Err(err).Msg("Invalid ID parameter")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	log.Debug().Int("userId", userID).Int("id", id).Msg("Deleting item")

	result, err := dal.DB.Exec("DELETE FROM items WHERE id = ? AND user_id = ?",
		id, userID)
	if err != nil {
		logger.LogError(err, "Failed to delete item", map[string]interface{}{
			"userId": userID,
			"id":     id,
		})
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Debug().Int("userId", userID).Int("id", id).Msg("Item not found for deletion")
		return c.Status(404).JSON(fiber.Map{"error": "Item not found"})
	}

	log.Info().Int("userId", userID).Int("id", id).Msg("Item deleted successfully")
	return c.SendStatus(204)
}
