package main

import (
	"log/slog"
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
