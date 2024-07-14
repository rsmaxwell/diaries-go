package loggerlevel

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func SetLoggerLevel() error {

	value, exists := os.LookupEnv("LOGGER_LEVEL")

	if exists {

		level := slog.LevelInfo

		switch strings.ToLower(value) {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			text := fmt.Sprintf("Unexpected logging level: env LOGGER_LEVEL = %s", value)
			slog.Info(text)
			return fmt.Errorf(text)
		}

		slog.SetLogLoggerLevel(level)
	}

	return nil
}
