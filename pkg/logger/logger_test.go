package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		raw  string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"  warn ", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"", slog.LevelInfo},
		{"info", slog.LevelInfo},
		{"unknown", slog.LevelInfo},
	}

	for _, c := range cases {
		if got := parseLevel(c.raw); got != c.want {
			t.Errorf("parseLevel(%q) = %v, want %v", c.raw, got, c.want)
		}
	}
}

// captureHandler is a minimal slog.Handler that records emitted messages.
type captureHandler struct {
	buf    *bytes.Buffer
	level  slog.Level
	attrs  []slog.Attr
	groups []string
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.buf.WriteString(r.Message)
	return nil
}

func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := *h
	nh.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &nh
}

func (h *captureHandler) WithGroup(name string) slog.Handler {
	nh := *h
	nh.groups = append(append([]string{}, h.groups...), name)
	return &nh
}

func TestSlogWriter(t *testing.T) {
	var buf bytes.Buffer
	handler := &captureHandler{buf: &buf}
	l := slog.New(handler)

	w := &slogWriter{logger: l}
	msg := "hello world\n"

	n, err := w.Write([]byte(msg))
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != len(msg) {
		t.Errorf("Write returned n=%d, want %d", n, len(msg))
	}
	if got := buf.String(); !strings.Contains(got, "hello world") {
		t.Errorf("captured message %q does not contain %q", got, "hello world")
	}
}
