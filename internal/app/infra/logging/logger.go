// internal/app/infra/logging/dev_logger.go
package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

var (
	once   sync.Once
	logger zerolog.Logger
)

// init runs automatically, setting up a disabled logger by default.
// This ensures that if Initialize() is never called, logging is off.
func init() {
	once.Do(func() {
		logger = zerolog.Nop()
	})
}

// Initialize configures the logger based on the level provided.
// This should be called once at application startup.
func Initialize(logLevel string, format string, sink string) error {

	zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
		return strings.ToUpper(l.String())
	}

	// If the level "disabled", the logger remains a Nop logger.
	if logLevel == "disabled" {
		return nil
	}

	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		logger = zerolog.Nop() // Disable if the level is invalid.
		return fmt.Errorf("parse runtime log level %q: %w", logLevel, err)
	}

	var writer io.Writer

	switch sink {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		writer = os.Stderr
	}

	if format == "json" {
		writer = os.Stderr
	} else {
		writer = zerolog.ConsoleWriter{
			Out: writer,
			PartsOrder: []string{
				zerolog.LevelFieldName,
				zerolog.CallerFieldName,
				zerolog.MessageFieldName,
			},
			// Ensure full, uppercased level words in console output (e.g., INFO, WARN),
			// and pad to a fixed width so following columns align.
			// ConsoleWriter defaults to 3-letter abbreviations; override to match LevelFieldMarshalFunc.
			FormatLevel: func(i interface{}) string {
				return fmt.Sprintf("%-5s", strings.ToUpper(fmt.Sprint(i)))
			},
		}
	}

	logger = zerolog.New(writer).
		Level(level).
		With().
		Caller().
		Logger()

	logger.Info().Msgf("runtime logger initialized with level '%s'", level)

	return nil
}

// Get returns the configured logger instance.
func Get() zerolog.Logger {
	return logger
}
