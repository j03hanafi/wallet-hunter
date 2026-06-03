package util

import (
	"log/slog"
	"os"
)

func InitLogger() {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, opts)))
}
