package panylexpr

import (
	"context"
	"fmt"
	"log/slog"
)

func parseLevel(s string) (slog.Level, error) {
	var level slog.Level
	var err = level.UnmarshalText([]byte(s))
	return level, err
}

func log(logger *slog.Logger, level string, message string) (bool, error) {
	if logger == nil {
		return true, nil
	}
	llevel, err := parseLevel(level)
	if err != nil {
		return false, fmt.Errorf("error parsing log level '%s': %w", level, err)
	}

	logger.Log(context.Background(), llevel, message)
	return true, nil
}
