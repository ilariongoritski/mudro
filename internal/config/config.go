package config

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultDSN           = "postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable"
	DefaultTelegramLimit = 3800
	DefaultTelegramUser  = "sirilarion"
	DefaultOpenAIModel   = "gpt-4.1-mini"
	DefaultAPIAddr       = ":8080"
	DefaultAPIBaseURL    = "http://127.0.0.1:8080"
	DefaultCodexLogsDir  = ".codex/logs"
	DefaultReportMinutes = 30
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

func TelegramAllowedUsername() string {
	return strings.ToLower(envOr("TELEGRAM_ALLOWED_USERNAME", DefaultTelegramUser))
}

func OpenAIAPIKey() string {
	return strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
}

func OpenAIModel() string {
	return envOr("OPENAI_MODEL", DefaultOpenAIModel)
}

func ReportBotToken() string {
	return strings.TrimSpace(os.Getenv("REPORT_BOT_TOKEN"))
}

func ReportChatID() int64 {
	if v := strings.TrimSpace(os.Getenv("REPORT_CHAT_ID")); v != "" {
		if n, ok := parsePositiveInt64(v); ok {
			return n
		}
	}
	return 0
}

func ReportIntervalMinutes() int {
	if v := strings.TrimSpace(os.Getenv("REPORT_INTERVAL_MIN")); v != "" {
		if n, ok := parsePositiveInt(v); ok {
			return n
		}
	}
	return DefaultReportMinutes
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

func APIAddr() string {
	return envOr("API_ADDR", DefaultAPIAddr)
}

func APIBaseURL() string {
	return strings.TrimRight(envOr("API_BASE_URL", DefaultAPIBaseURL), "/")
}

func CodexLogsDir() string {
	return envOr("CODEX_LOGS_DIR", DefaultCodexLogsDir)
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

func parsePositiveInt64(v string) (int64, bool) {
	var n int64
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int64(c-'0')
		if n <= 0 {
			return 0, false
		}
	}
	return n, n > 0
}
