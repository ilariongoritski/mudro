package logger

import (
	"log"
	"log/slog"
	"os"
	"strings"
)

// Init sets up structured logging as the process-wide default.
// All existing log.Printf calls are automatically routed through slog.
//
// MUDRO_LOG_FORMAT controls the output format:
//
//	"json" → JSON lines (recommended for production)
//	"text" → human-readable (default for development)
//
// MUDRO_LOG_LEVEL controls the minimum level:
//
//	"debug", "info" (default), "warn", "error"
func Init(service string) {
	level := parseLevel(os.Getenv("MUDRO_LOG_LEVEL"))

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}

	switch strings.ToLower(strings.TrimSpace(os.Getenv("MUDRO_LOG_FORMAT"))) {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, opts)
	default:
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	handler = handler.WithAttrs([]slog.Attr{
		slog.String("service", service),
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Redirect standard log package through slog so existing log.Printf calls
	// produce structured output.
	log.SetOutput(&slogWriter{logger: logger})
	log.SetFlags(0) // slog handles timestamps
}

func parseLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
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

// slogWriter adapts slog to the io.Writer interface so standard log calls
// are emitted as structured log entries.
type slogWriter struct {
	logger *slog.Logger
}

func (w *slogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\n")
	w.logger.Info(msg)
	return len(p), nil
}
