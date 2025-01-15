package log

import (
	"log/slog"
	"os"
)

func GetLogger() func(handle slog.Handler) *slog.Logger {
	var logger *slog.Logger

	return func(handler slog.Handler) *slog.Logger {
		if logger != nil {
			return logger
		}

		logger = slog.New(handler)

		return logger
	}
}

var DefaultLogger *slog.Logger = GetLogger()(slog.NewJSONHandler(
	os.Stdout,
	&slog.HandlerOptions{
		AddSource: true,
		Level: slog.LevelInfo,
	},
))
