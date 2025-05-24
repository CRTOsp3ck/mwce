// pkg/logger/logger.go

package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// NewLogger creates a new zerolog logger
func NewLogger() zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	return zerolog.New(output).With().Timestamp().Logger()
}
