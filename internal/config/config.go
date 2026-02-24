package config

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultDSN           = "postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable"
	DefaultTelegramLimit = 3800
)

func DSN() string {
	return envOr("DSN", DefaultDSN)
}

func TelegramMessageLimit() int {
	if v := strings.TrimSpace(os.Getenv("TELEGRAM_MESSAGE_LIMIT")); v != "" {
		if n, ok := parsePositiveInt(v); ok {
			return n
		}
	}
	return DefaultTelegramLimit
}

func RepoRoot() string {
	if v := strings.TrimSpace(os.Getenv("MUDRO_ROOT")); v != "" {
		return v
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	dir := cwd
	for {
		if fileExists(filepath.Join(dir, "Makefile")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return cwd
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func parsePositiveInt(v string) (int, bool) {
	n := 0
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
		if n <= 0 {
			return 0, false
		}
	}
	return n, n > 0
}
