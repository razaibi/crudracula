package logger

import (
	"context"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

type contextKey string

const RequestIDKey contextKey = "requestID"

func InitLogger() {
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	file, err := os.OpenFile(
		"app.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not open log file")
	}

	multi := zerolog.MultiLevelWriter(consoleWriter, file)

	Logger = zerolog.New(multi).
		With().
		Timestamp().
		Caller().
		Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("ENV") == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Logger = Logger
}

// LogRequest now includes error information and request ID
func LogRequest(method, path string, duration time.Duration, status int, ctx context.Context, err error) {
	requestID := ctx.Value(RequestIDKey).(string)

	event := Logger.Info()
	if status >= 500 {
		event = Logger.Error()
	}

	event.
		Str("requestId", requestID).
		Str("method", method).
		Str("path", path).
		Dur("duration", duration).
		Int("status", status)

	if err != nil {
		event.Err(err)
	}

	event.Msg("Request processed")
}

// LogError now includes request ID from context
func LogError(ctx context.Context, err error, msg string, fields map[string]interface{}) {
	requestID := ctx.Value(RequestIDKey).(string)

	event := log.Error().
		Str("requestId", requestID).
		Err(err)

	if fields != nil {
		event.Fields(fields)
	}

	event.Msg(msg)
}

// Middleware to add request ID to context
func RequestIDMiddleware(c *fiber.Ctx) error {
	requestID := uuid.New().String()
	ctx := context.WithValue(c.Context(), RequestIDKey, requestID)
	c.SetUserContext(ctx)
	c.Set("X-Request-ID", requestID)
	return c.Next()
}
