package main

import (
	"crudracula/dal"
	"crudracula/logger"
	"crudracula/logic"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	logger.InitLogger()
	log.Info().Msg("Starting application...")

	// Initialize database
	dal.InitDB()
	defer dal.DB.Close()
	log.Info().Msg("Database initialized successfully")

	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Enable CORS
	app.Use(cors.New())

	// Add request logging middleware
	app.Use(requestLogger)

	// Serve static files
	app.Static("/", "./public")

	// CRUD endpoints
	app.Get("/api/items", logic.GetItems)
	app.Get("/api/items/:id", logic.GetItem)
	app.Post("/api/items", logic.CreateItem)
	app.Put("/api/items/:id", logic.UpdateItem)
	app.Delete("/api/items/:id", logic.DeleteItem)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Info().Str("port", port).Msg("Server starting...")
	if err := app.Listen(":" + port); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

func requestLogger(c *fiber.Ctx) error {
	start := time.Now()

	// Continue stack
	err := c.Next()

	// Log after request is processed
	duration := time.Since(start)
	logger.LogRequest(
		c.Method(),
		c.Path(),
		duration,
		c.Response().StatusCode(),
	)

	return err
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	// Log error with context
	logger.LogError(err, "Request error occurred", map[string]interface{}{
		"path":   c.Path(),
		"method": c.Method(),
		"ip":     c.IP(),
	})

	// Default 500 status code
	code := fiber.StatusInternalServerError

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Return error response
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
