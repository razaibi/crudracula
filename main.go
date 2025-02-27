package main

import (
	"crudracula/dal"
	"crudracula/encoders"
	"crudracula/logger"
	"crudracula/logic"
	"crudracula/middlewares"
	"errors"
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
		JSONEncoder: encoders.Marshal,
		JSONDecoder: encoders.Unmarshal,
	})

	// Enable CORS with specific configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Add request ID middleware first
	app.Use(logger.RequestIDMiddleware)

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

	// Public page routes
	app.Get("/", logic.GetItemsPage)
	app.Get("/login", logic.GetLoginPage)
	app.Get("/logout", logic.GetLogoutPage)
	app.Get("/signup", logic.GetSignUpPage)
	app.Get("/reset-password", logic.GetResetPasswordPage)

	// Admin pages (protected by middleware)
	adminPages := app.Group("/admin")
	adminPages.Use(middlewares.AuthMiddleware)
	adminPages.Use(middlewares.RequirePermission("manage_roles"))
	adminPages.Get("/roles", logic.GetRolesPage)

	// Public auth endpoints
	app.Post("/api/signup", logic.Signup)
	app.Post("/api/login", logic.Login)
	app.Post("/api/request-reset", logic.RequestPasswordReset)
	app.Post("/api/reset-password", logic.ResetPassword)

	// Protected API routes group
	api := app.Group("/api")
	api.Use(middlewares.AuthMiddleware)

	// CRUD endpoints (protected by middleware)
	items := api.Group("/items")
	items.Get("/", middlewares.RequirePermission("read_item"), logic.GetItems)
	items.Get("/:id", middlewares.RequirePermission("read_item"), logic.GetItem)
	items.Post("/", middlewares.RequirePermission("create_item"), logic.CreateItem)
	items.Put("/:id", middlewares.RequirePermission("update_item"), logic.UpdateItem)
	items.Delete("/:id", middlewares.RequirePermission("delete_item"), logic.DeleteItem)

	// Role management endpoints (protected + require manage_roles permission)
	roles := api.Group("/roles")
	roles.Use(middlewares.RequirePermission("manage_roles"))
	roles.Get("/", logic.GetRoles)
	roles.Get("/:id", logic.GetRole)
	roles.Post("/", logic.CreateRole)
	roles.Put("/:id", logic.UpdateRole)
	roles.Delete("/:id", logic.DeleteRole)

	// Permission endpoints
	permissions := api.Group("/permissions")
	permissions.Use(middlewares.RequirePermission("manage_roles"))
	permissions.Get("/", logic.GetPermissions)
	permissions.Get("/check/:permission", logic.CheckPermission)

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

	// Get error from response if status is 500
	var responseErr error
	if c.Response().StatusCode() == fiber.StatusInternalServerError {
		if errBody := c.Response().Body(); len(errBody) > 0 {
			var errResp struct {
				Error string `json:"error"`
			}
			if unmarshalErr := encoders.Unmarshal(errBody, &errResp); unmarshalErr == nil {
				responseErr = errors.New(errResp.Error)
			}
		}
	}

	// Use response error if available, otherwise use middleware error
	logErr := responseErr
	if logErr == nil {
		logErr = err
	}

	// Log after request is processed
	duration := time.Since(start)
	logger.LogRequest(
		c.Method(),
		c.Path(),
		duration,
		c.Response().StatusCode(),
		c.UserContext(),
		logErr,
	)

	return err
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	// Use the original error
	originalErr := err

	// Log error with context
	logger.LogError(
		c.UserContext(),
		originalErr,
		"Request error occurred",
		map[string]interface{}{
			"path":    c.Path(),
			"method":  c.Method(),
			"ip":      c.IP(),
			"details": err.Error(),
		},
	)

	// Default 500 status code
	code := fiber.StatusInternalServerError

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Check if it's a permission error
	if err.Error() == "permission denied" {
		code = fiber.StatusForbidden
	}

	// Return error response with detailed message
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
