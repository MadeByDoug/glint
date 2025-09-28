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
	once      sync.Once
	devLogger zerolog.Logger
)

// init runs automatically, setting up a disabled logger by default.
// This ensures that if Initialize() is never called, logging is off.
func init() {
	once.Do(func() {
		devLogger = zerolog.Nop()
	})
}

// Initialize configures the developer logger based on the level provided.
// This should be called once at application startup.
func Initialize(devLogLevel string, format string) error {
	zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
		return strings.ToUpper(l.String())
	}

	// If the level is empty or "disabled", the logger remains a Nop logger.
	if devLogLevel == "" || devLogLevel == "disabled" {
		devLogger = zerolog.Nop()
		return nil
	}

	level, err := zerolog.ParseLevel(devLogLevel)
	if err != nil {
		devLogger = zerolog.Nop() // Disable if the level is invalid.
		return fmt.Errorf("parse dev log level %q: %w", devLogLevel, err)
	}

	// All dev logs go to std error
	var writer io.Writer

	if format == "json" {
		writer = os.Stderr
    } else {
        writer = zerolog.ConsoleWriter{
            Out: os.Stderr,
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

	devLogger = zerolog.New(writer).
		Level(level).
		With().
		Caller().
		Logger()

	devLogger.Info().Msgf("developer logger initialized with level '%s'", level)

	return nil
}

// Get returns the configured logger instance.
func Get() zerolog.Logger {
	return devLogger
}
