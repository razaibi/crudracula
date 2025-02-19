package main

import (
	"crudracula/dal"
	"crudracula/logger"
	"crudracula/logic"
	"crudracula/middlewares"
	"encoding/json"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html/v2"
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

	// Set Views Engine with proper configuration
	engine := html.New("./views", ".html")
	engine.Reload(true) // Enable reloading in development
	engine.Debug(true)  // Enable debug mode in development

	// Configure Fiber with proper settings
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
		Views:        engine,
		// Add proper JSON settings
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// Enable CORS with specific configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Add request logging middleware
	app.Use(requestLogger)

	// Enable Gzip Compression with proper settings
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Serve static files with specific configuration
	app.Static("/", "./public", fiber.Static{
		Compress:      true,
		Browse:        false,
		Index:         "index.html",
		CacheDuration: 24 * time.Hour,
		MaxAge:        24 * 60 * 60,
	})

	// Serve CSS files specifically
	app.Static("/css", "./public/css", fiber.Static{
		Compress:      true,
		Browse:        false,
		CacheDuration: 24 * time.Hour,
		MaxAge:        24 * 60 * 60,
	})

	// Route for serving index page
	app.Get("/", logic.GetItemsPage)
	app.Get("/login", logic.GetLoginPage)
	app.Get("/logout", logic.GetLogoutPage)
	app.Get("/signup", logic.GetSignUpPage)
	app.Get("/reset-password", logic.GetResetPasswordPage)

	// Public auth endpoints
	app.Post("/api/signup", logic.Signup)
	app.Post("/api/login", logic.Login)
	app.Post("/api/request-reset", logic.RequestPasswordReset)
	app.Post("/api/reset-password", logic.ResetPassword)

	// Protected routes group
	api := app.Group("/api")
	api.Use(middlewares.AuthMiddleware)

	// CRUD endpoints (protected)
	api.Get("/items", logic.GetItems)
	api.Get("/items/:id", logic.GetItem)
	api.Post("/items", logic.CreateItem)
	api.Put("/items/:id", logic.UpdateItem)
	api.Delete("/items/:id", logic.DeleteItem)

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
