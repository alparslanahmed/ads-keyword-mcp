package app

import (
	"log/slog"
	"os"
	"strings"
)

// NewLogger builds a slog logger that writes to stderr. Writing to stderr (never
// stdout) is essential for the stdio MCP transport, where stdout carries only
// JSON-RPC messages.
func NewLogger(level string) *slog.Logger {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: parseLevel(level)})
	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
