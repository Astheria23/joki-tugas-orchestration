package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Log is a pre-configured logger instance.
var Log zerolog.Logger

func init() {
	// Use Unix timestamps for performance in production, but configure format
	zerolog.TimeFieldFormat = time.RFC3339

	// If in development mode (default), output pretty logs to stderr
	env := os.Getenv("APP_ENV")
	if env != "production" {
		Log = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "2006-01-02 15:04:05",
		}).With().Timestamp().Logger()
	} else {
		Log = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

	// Set as global logger too
	log.Logger = Log
}

// Info logs a message at info level.
func Info(msg string, keyvals ...interface{}) {
	Log.Info().Fields(keyvals).Msg(msg)
}

// Error logs a message at error level with an error object.
func Error(err error, msg string, keyvals ...interface{}) {
	Log.Error().Err(err).Fields(keyvals).Msg(msg)
}

// Debug logs a message at debug level.
func Debug(msg string, keyvals ...interface{}) {
	Log.Debug().Fields(keyvals).Msg(msg)
}

// Warn logs a message at warn level.
func Warn(msg string, keyvals ...interface{}) {
	Log.Warn().Fields(keyvals).Msg(msg)
}
