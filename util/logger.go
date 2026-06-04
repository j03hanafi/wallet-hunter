package util

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// InitLogger sets up slog to write JSON to both stderr and a file in logs/.
// Returns the open file so caller can defer Close().
func InitLogger() (*os.File, error) {
	const logDir = "logs"

	// Create logs folder if missing. MkdirAll no error if exist.
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	// Log file named by date. One file per day.
	name := fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02"))
	path := filepath.Join(logDir, name)

	// Open/create file, append mode so restart no truncate.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	// Write both: stderr (live view) + file (persist).
	mw := io.MultiWriter(os.Stderr, f)

	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(mw, opts)))

	return f, nil
}
