// Package logger wires zerolog with sane defaults.
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New returns a JSON logger bound to stdout with the given level.
func New(level string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	lvl, err := zerolog.ParseLevel(level)
	if err != nil || lvl == zerolog.NoLevel {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}
