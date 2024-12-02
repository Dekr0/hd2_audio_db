package main

import (
	"log/slog"
	"os"
)

func getLogger() func(handle slog.Handler) *slog.Logger {
	var logger *slog.Logger

	return func(handler slog.Handler) *slog.Logger {
		if logger != nil {
			return logger
		}

		logger = slog.New(handler)

		return logger
	}
}

var DefaultLogger *slog.Logger = getLogger()(slog.NewJSONHandler(
	os.Stdout,
	&slog.HandlerOptions{
		AddSource: true,
		Level: slog.LevelInfo,
	},
))
