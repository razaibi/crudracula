package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func InitLogger() {
	// Set up pretty console logging for development
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	// Create multi-writer for both console and file
	_, err := os.OpenFile(
		"app.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not open log file")
	}

	// Set global logger
	Logger = zerolog.New(consoleWriter).
		With().
		Timestamp().
		Caller().
		Logger()

	// Set log level based on environment
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("ENV") == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Set default logger
	log.Logger = Logger
}

// LogRequest logs incoming HTTP requests
func LogRequest(method, path string, duration time.Duration, status int) {
	Logger.Info().
		Str("method", method).
		Str("path", path).
		Dur("duration", duration).
		Int("status", status).
		Msg("Request processed")
}

// LogError logs error messages with context
func LogError(err error, message string, fields map[string]interface{}) {
	event := Logger.Error().Err(err)
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	event.Msg(message)
}
