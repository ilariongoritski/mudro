package config

import (
	"fmt"
	"net"
	"net/url"
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
	DefaultAPIRateRPS    = 20
	DefaultAPIRateBurst  = 40
	DefaultCodexLogsDir  = ".codex/logs"
	DefaultReportMinutes = 30
	DefaultRedisAddr     = "localhost:6379"
	DefaultKafkaClientID = "mudro"
	DefaultKafkaTopic    = "mudro.agent.tasks.v1"
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

func APIRateLimitRPS() int {
	if v := strings.TrimSpace(os.Getenv("API_RATE_LIMIT_RPS")); v != "" {
		if n, ok := parseNonNegativeInt(v); ok {
			return n
		}
	}
	return DefaultAPIRateRPS
}

func APIRateLimitBurst() int {
	if v := strings.TrimSpace(os.Getenv("API_RATE_LIMIT_BURST")); v != "" {
		if n, ok := parseNonNegativeInt(v); ok {
			return n
		}
	}
	return DefaultAPIRateBurst
}

func CodexLogsDir() string {
	return envOr("CODEX_LOGS_DIR", DefaultCodexLogsDir)
}

func RedisAddr() string {
	return envOr("REDIS_ADDR", DefaultRedisAddr)
}

func RedisPassword() string {
	return strings.TrimSpace(os.Getenv("REDIS_PASSWORD"))
}

func RedisDB() int {
	if v := strings.TrimSpace(os.Getenv("REDIS_DB")); v != "" {
		if n, ok := parseNonNegativeInt(v); ok {
			return n
		}
	}
	return 0
}

func RedisRateLimitEnabled() bool {
	return parseBoolEnv("REDIS_RATE_LIMIT_ENABLED")
}

func KafkaEnabled() bool {
	return parseBoolEnv("KAFKA_ENABLED")
}

func KafkaBrokers() []string {
	raw := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func KafkaClientID() string {
	return envOr("KAFKA_CLIENT_ID", DefaultKafkaClientID)
}

func KafkaTopicTasks() string {
	return envOr("KAFKA_TOPIC_TASKS", DefaultKafkaTopic)
}

func MudroEnv() string {
	return strings.ToLower(strings.TrimSpace(os.Getenv("MUDRO_ENV")))
}

func ValidateRequiredEnv(required ...string) error {
	var missing []string
	for _, key := range required {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("missing required env: %s", strings.Join(missing, ", "))
}

func ValidateRuntimeDSN(service, dsn string) error {
	parsed, err := url.Parse(strings.TrimSpace(dsn))
	if err != nil {
		return fmt.Errorf("%s DSN parse: %w", service, err)
	}
	if parsed.Scheme == "" {
		return fmt.Errorf("%s DSN parse: missing scheme", service)
	}

	user := ""
	if parsed.User != nil {
		user = strings.ToLower(strings.TrimSpace(parsed.User.Username()))
	}
	if user != "postgres" {
		return nil
	}

	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if isProductionLikeEnv(MudroEnv()) {
		return fmt.Errorf("%s runtime refuses superuser DSN for MUDRO_ENV=%q", service, MudroEnv())
	}
	if isLocalDevDBHost(host) {
		return nil
	}
	return fmt.Errorf("%s runtime refuses superuser DSN for non-local host %q", service, host)
}

func ValidateRuntime(service string, requiredEnv ...string) error {
	if err := ValidateRequiredEnv(requiredEnv...); err != nil {
		return fmt.Errorf("%s config invalid: %w", service, err)
	}
	if err := ValidateRuntimeDSN(service, DSN()); err != nil {
		return err
	}
	return nil
}

func isProductionLikeEnv(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "prod", "production", "stage", "staging":
		return true
	default:
		return false
	}
}

func isLocalDevDBHost(host string) bool {
	if host == "" {
		return false
	}
	switch host {
	case "localhost", "127.0.0.1", "::1", "db":
		return true
	}
	return net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback()
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

func parseNonNegativeInt(v string) (int, bool) {
	if v == "" {
		return 0, false
	}
	n := 0
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
		if n < 0 {
			return 0, false
		}
	}
	return n, true
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

func parseBoolEnv(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
