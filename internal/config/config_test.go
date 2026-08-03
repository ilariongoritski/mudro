package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePositiveInt(t *testing.T) {
	if n, ok := parsePositiveInt("123"); !ok || n != 123 {
		t.Fatalf("parsePositiveInt ok=%v n=%d", ok, n)
	}
	if _, ok := parsePositiveInt("0"); ok {
		t.Fatal("expected invalid zero")
	}
	if _, ok := parsePositiveInt("1a"); ok {
		t.Fatal("expected invalid alpha")
	}
}

func TestParsePositiveInt64(t *testing.T) {
	if n, ok := parsePositiveInt64("123"); !ok || n != 123 {
		t.Fatalf("parsePositiveInt64 ok=%v n=%d", ok, n)
	}
	if _, ok := parsePositiveInt64("-1"); ok {
		t.Fatal("expected invalid negative")
	}
}

func TestParseNonNegativeInt(t *testing.T) {
	if n, ok := parseNonNegativeInt("0"); !ok || n != 0 {
		t.Fatalf("parseNonNegativeInt zero ok=%v n=%d", ok, n)
	}
	if n, ok := parseNonNegativeInt("42"); !ok || n != 42 {
		t.Fatalf("parseNonNegativeInt ok=%v n=%d", ok, n)
	}
	if _, ok := parseNonNegativeInt("-1"); ok {
		t.Fatal("expected invalid negative")
	}
	if _, ok := parseNonNegativeInt("abc"); ok {
		t.Fatal("expected invalid alpha")
	}
}

func TestEnvOrAndAPIBaseURL(t *testing.T) {
	t.Setenv("X_TEST_ENV", "  value ")
	if got := envOr("X_TEST_ENV", "def"); got != "value" {
		t.Fatalf("envOr=%q", got)
	}

	t.Setenv("API_BASE_URL", "http://localhost:8080/")
	if got := APIBaseURL(); got != "http://localhost:8080" {
		t.Fatalf("APIBaseURL=%q", got)
	}
}

func TestLLMConfigPrefersGenericEnv(t *testing.T) {
	t.Setenv("LLM_API_KEY", "llm-key")
	t.Setenv("LLM_MODEL", "glm-5.2")
	t.Setenv("LLM_BASE_URL", "https://po.zapro.su/v1/")
	t.Setenv("OPENROUTER_API_KEY", "openrouter-key")
	t.Setenv("OPENROUTER_MODEL", "openrouter/model")
	t.Setenv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1")

	if got := LLMAPIKey(); got != "llm-key" {
		t.Fatalf("LLMAPIKey=%q", got)
	}
	if got := LLMModel(); got != "glm-5.2" {
		t.Fatalf("LLMModel=%q", got)
	}
	if got := LLMBaseURL(); got != "https://po.zapro.su/v1" {
		t.Fatalf("LLMBaseURL=%q", got)
	}
}

func TestLLMConfigKeepsOpenRouterFallback(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "openrouter-key")
	t.Setenv("OPENROUTER_MODEL", "openrouter/model")
	t.Setenv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1/")

	if got := LLMAPIKey(); got != "openrouter-key" {
		t.Fatalf("LLMAPIKey=%q", got)
	}
	if got := LLMModel(); got != "openrouter/model" {
		t.Fatalf("LLMModel=%q", got)
	}
	if got := LLMBaseURL(); got != "https://openrouter.ai/api/v1" {
		t.Fatalf("LLMBaseURL=%q", got)
	}
}

func TestLLMConfigDefaultsToZaproWhenGenericKeyIsSet(t *testing.T) {
	t.Setenv("LLM_API_KEY", "llm-key")

	if got := LLMModel(); got != DefaultLLMModel {
		t.Fatalf("LLMModel=%q", got)
	}
	if got := LLMBaseURL(); got != DefaultLLMBaseURL {
		t.Fatalf("LLMBaseURL=%q", got)
	}
}

func TestTelegramAllowedUsernameRequiresExplicitEnv(t *testing.T) {
	t.Setenv("TELEGRAM_ALLOWED_USERNAME", "")
	if got := TelegramAllowedUsername(); got != "" {
		t.Fatalf("TelegramAllowedUsername=%q, want empty", got)
	}

	t.Setenv("TELEGRAM_ALLOWED_USERNAME", "  MudroAdmin ")
	if got := TelegramAllowedUsername(); got != "mudroadmin" {
		t.Fatalf("TelegramAllowedUsername=%q", got)
	}
}

func TestRepoRootByEnv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MUDRO_ROOT", root)
	if got := RepoRoot(); got != root {
		t.Fatalf("RepoRoot=%q want=%q", got, root)
	}
}

func TestFileExists(t *testing.T) {
	f := filepath.Join(t.TempDir(), "x.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !fileExists(f) {
		t.Fatal("expected fileExists true")
	}
}

func TestAPIRateLimitConfig(t *testing.T) {
	t.Setenv("API_RATE_LIMIT_RPS", "15")
	t.Setenv("API_RATE_LIMIT_BURST", "30")
	if got := APIRateLimitRPS(); got != 15 {
		t.Fatalf("APIRateLimitRPS=%d", got)
	}
	if got := APIRateLimitBurst(); got != 30 {
		t.Fatalf("APIRateLimitBurst=%d", got)
	}
}

func TestParseBoolEnv(t *testing.T) {
	t.Setenv("X_BOOL", "true")
	if !parseBoolEnv("X_BOOL") {
		t.Fatal("expected true")
	}
	t.Setenv("X_BOOL", "0")
	if parseBoolEnv("X_BOOL") {
		t.Fatal("expected false")
	}
}

func TestKafkaBrokers(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "kafka:9092, broker2:9092")
	b := KafkaBrokers()
	if len(b) != 2 {
		t.Fatalf("brokers len=%d", len(b))
	}
	if b[0] != "kafka:9092" || b[1] != "broker2:9092" {
		t.Fatalf("brokers=%v", b)
	}
}

func TestValidateRequiredEnv(t *testing.T) {
	t.Setenv("X_ONE", "ok")
	t.Setenv("X_TWO", "")
	err := ValidateRequiredEnv("X_ONE", "X_TWO")
	if err == nil {
		t.Fatal("expected missing env error")
	}
	if got := err.Error(); got != "missing required env: X_TWO" {
		t.Fatalf("unexpected error: %q", got)
	}

	t.Setenv("X_TWO", "ok")
	if err := ValidateRequiredEnv("X_ONE", "X_TWO"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRuntimeDSNAllowsLocalDevSuperuser(t *testing.T) {
	t.Setenv("MUDRO_ENV", "development")
	err := ValidateRuntimeDSN("api", "postgres://postgres:***@db:5432/gallery?sslmode=disable")
	if err != nil {
		t.Fatalf("expected local dev DSN to pass, got %v", err)
	}
}

func TestValidateRuntimeDSNRejectsSuperuserOutsideLocalDev(t *testing.T) {
	err := ValidateRuntimeDSN("api", "postgres://postgres:***@10.0.0.5:5432/gallery?sslmode=disable")
	if err == nil {
		t.Fatal("expected superuser DSN rejection")
	}
}

func TestValidateRuntimeDSNRejectsProductionSuperuser(t *testing.T) {
	t.Setenv("MUDRO_ENV", "production")
	err := ValidateRuntimeDSN("api", "postgres://postgres:***@127.0.0.1:5432/gallery?sslmode=disable")
	if err == nil {
		t.Fatal("expected production superuser DSN rejection")
	}
}

func TestValidateRuntimeAllowsNonSuperuserProductionDSN(t *testing.T) {
	t.Setenv("MUDRO_ENV", "production")
	t.Setenv("DSN", "postgres://mudro_app:***@127.0.0.1:5432/gallery?sslmode=disable")
	t.Setenv("JWT_SECRET", "test-production-jwt-secret-32chars!!")
	if err := ValidateRuntime("api", "DSN"); err != nil {
		t.Fatalf("expected runtime validation to pass, got %v", err)
	}
}

func TestTelegramMessageLimit(t *testing.T) {
	t.Setenv("TELEGRAM_MESSAGE_LIMIT", "1000")
	if got := TelegramMessageLimit(); got != 1000 {
		t.Fatalf("TelegramMessageLimit=%d", got)
	}

	// Default when unset
	t.Setenv("TELEGRAM_MESSAGE_LIMIT", "")
	if got := TelegramMessageLimit(); got != DefaultTelegramLimit {
		t.Fatalf("TelegramMessageLimit default=%d", got)
	}

	// Invalid value falls back to default
	t.Setenv("TELEGRAM_MESSAGE_LIMIT", "invalid")
	if got := TelegramMessageLimit(); got != DefaultTelegramLimit {
		t.Fatalf("TelegramMessageLimit invalid fallback=%d", got)
	}
}

func TestAPIAddr(t *testing.T) {
	t.Setenv("PORT", "9090")
	if got := APIAddr(); got != ":9090" {
		t.Fatalf("APIAddr=%q", got)
	}

	t.Setenv("PORT", "")
	t.Setenv("API_ADDR", ":3000")
	if got := APIAddr(); got != ":3000" {
		t.Fatalf("APIAddr from API_ADDR=%q", got)
	}

	// Default
	t.Setenv("API_ADDR", "")
	if got := APIAddr(); got != DefaultAPIAddr {
		t.Fatalf("APIAddr default=%q", got)
	}
}

func TestCasinoServiceURL(t *testing.T) {
	t.Setenv("CASINO_SERVICE_URL", "http://casino:8081/")
	if got := CasinoServiceURL(); got != "http://casino:8081" {
		t.Fatalf("CasinoServiceURL=%q", got)
	}

	t.Setenv("CASINO_SERVICE_URL", "")
	if got := CasinoServiceURL(); got != DefaultCasinoServiceURL {
		t.Fatalf("CasinoServiceURL default=%q", got)
	}
}

func TestCodexLogsDir(t *testing.T) {
	t.Setenv("CODEX_LOGS_DIR", "/custom/logs")
	if got := CodexLogsDir(); got != "/custom/logs" {
		t.Fatalf("CodexLogsDir=%q", got)
	}

	t.Setenv("CODEX_LOGS_DIR", "")
	if got := CodexLogsDir(); got != DefaultCodexLogsDir {
		t.Fatalf("CodexLogsDir default=%q", got)
	}
}

func TestMudroEnv(t *testing.T) {
	t.Setenv("MUDRO_ENV", "production")
	if got := MudroEnv(); got != "production" {
		t.Fatalf("MudroEnv=%q", got)
	}

	t.Setenv("MUDRO_ENV", "  Staging  ")
	if got := MudroEnv(); got != "staging" {
		t.Fatalf("MudroEnv trimmed/lowered=%q", got)
	}

	t.Setenv("MUDRO_ENV", "")
	if got := MudroEnv(); got != "" {
		t.Fatalf("MudroEnv default=%q", got)
	}
}

func TestValidateMovieCatalogRuntime(t *testing.T) {
	// Valid DSN with proper JWT secret
	t.Setenv("DSN", "postgres://user:***@localhost:5433/movies?sslmode=disable")
	t.Setenv("JWT_SECRET", "valid-secret-12345678901234")
	if err := ValidateMovieCatalogRuntime(); err != nil {
		t.Fatalf("expected valid config to pass: %v", err)
	}

	// Missing DSN - ValidateMovieCatalogRuntime only validates DSN format, not presence
	t.Setenv("DSN", "")
	if err := ValidateMovieCatalogRuntime(); err != nil {
		t.Fatalf("expected no error for empty DSN, got: %v", err)
	}

	// Invalid JWT secret - ValidateMovieCatalogRuntime doesn't validate JWT secret
	t.Setenv("DSN", "postgres://user:***@localhost:5433/movies?sslmode=disable")
	t.Setenv("JWT_SECRET", "short")
	if err := ValidateMovieCatalogRuntime(); err != nil {
		t.Fatalf("expected no error for invalid JWT secret, got: %v", err)
	}
}

func TestOpenAIConfigAliases(t *testing.T) {
	t.Setenv("LLM_API_KEY", "llm-key")
	t.Setenv("LLM_MODEL", "glm-5.2")

	if got := OpenAIAPIKey(); got != "llm-key" {
		t.Fatalf("OpenAIAPIKey=%q", got)
	}
	if got := OpenAIModel(); got != "glm-5.2" {
		t.Fatalf("OpenAIModel=%q", got)
	}
}

func TestReportConfig(t *testing.T) {
	t.Setenv("REPORT_BOT_TOKEN", "report-bot-token")
	t.Setenv("REPORT_CHAT_ID", "12345")
	t.Setenv("REPORT_INTERVAL_MIN", "45")

	if got := ReportBotToken(); got != "report-bot-token" {
		t.Fatalf("ReportBotToken=%q", got)
	}
	if got := ReportChatID(); got != 12345 {
		t.Fatalf("ReportChatID=%d", got)
	}
	if got := ReportIntervalMinutes(); got != 45 {
		t.Fatalf("ReportIntervalMinutes=%d", got)
	}

	// Defaults
	t.Setenv("REPORT_BOT_TOKEN", "")
	t.Setenv("REPORT_CHAT_ID", "")
	t.Setenv("REPORT_INTERVAL_MIN", "")

	if got := ReportBotToken(); got != "" {
		t.Fatalf("ReportBotToken default=%q", got)
	}
	if got := ReportChatID(); got != 0 {
		t.Fatalf("ReportChatID default=%d", got)
	}
	if got := ReportIntervalMinutes(); got != DefaultReportMinutes {
		t.Fatalf("ReportIntervalMinutes default=%d", got)
	}

	// Invalid chat ID falls back to 0
	t.Setenv("REPORT_CHAT_ID", "invalid")
	if got := ReportChatID(); got != 0 {
		t.Fatalf("ReportChatID invalid fallback=%d", got)
	}
}

func TestJWTExpiryHours(t *testing.T) {
	t.Setenv("JWT_EXPIRY_HOURS", "72")
	if got := JWTExpiryHours(); got != 72 {
		t.Fatalf("JWTExpiryHours=%d", got)
	}

	// Default when unset
	t.Setenv("JWT_EXPIRY_HOURS", "")
	if got := JWTExpiryHours(); got != 168 {
		t.Fatalf("JWTExpiryHours default=%d", got)
	}

	// Invalid value falls back to default
	t.Setenv("JWT_EXPIRY_HOURS", "invalid")
	if got := JWTExpiryHours(); got != 168 {
		t.Fatalf("JWTExpiryHours invalid fallback=%d", got)
	}
}

func TestJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "custom-secret-value")
	if got := JWTSecret(); got != "custom-secret-value" {
		t.Fatalf("JWTSecret=%q", got)
	}

	// Default when unset
	t.Setenv("JWT_SECRET", "")
	if got := JWTSecret(); got != "mudro-dev-secret-change-me" {
		t.Fatalf("JWTSecret default=%q", got)
	}
}

func TestValidateJWTSecret(t *testing.T) {
	// Default insecure secret should fail
	t.Setenv("JWT_SECRET", "mudro-dev-secret-change-me")
	if err := ValidateJWTSecret(); err == nil {
		t.Fatal("expected error for default insecure secret")
	}

	// Too short secret should fail
	t.Setenv("JWT_SECRET", "short")
	if err := ValidateJWTSecret(); err == nil {
		t.Fatal("expected error for too short secret")
	}

	// Valid secret (16+ chars, not default) should pass
	t.Setenv("JWT_SECRET", "valid-secret-1234567890")
	if err := ValidateJWTSecret(); err != nil {
		t.Fatalf("unexpected error for valid secret: %v", err)
	}

	// Exactly 16 chars should pass
	t.Setenv("JWT_SECRET", "1234567890123456")
	if err := ValidateJWTSecret(); err != nil {
		t.Fatalf("unexpected error for 16-char secret: %v", err)
	}
}

func TestRedisConfig(t *testing.T) {
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("REDIS_PASSWORD", "secret")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("REDIS_RATE_LIMIT_ENABLED", "true")

	if got := RedisAddr(); got != "redis:6379" {
		t.Fatalf("RedisAddr=%q", got)
	}
	if got := RedisPassword(); got != "secret" {
		t.Fatalf("RedisPassword=%q", got)
	}
	if got := RedisDB(); got != 2 {
		t.Fatalf("RedisDB=%d", got)
	}
	if !RedisRateLimitEnabled() {
		t.Fatal("RedisRateLimitEnabled=true")
	}

	// Defaults when unset
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("REDIS_PASSWORD", "")
	t.Setenv("REDIS_DB", "")
	t.Setenv("REDIS_RATE_LIMIT_ENABLED", "")

	if got := RedisAddr(); got != "localhost:6379" {
		t.Fatalf("RedisAddr default=%q", got)
	}
	if got := RedisPassword(); got != "" {
		t.Fatalf("RedisPassword default=%q", got)
	}
	if got := RedisDB(); got != 0 {
		t.Fatalf("RedisDB default=%d", got)
	}
	if RedisRateLimitEnabled() {
		t.Fatal("RedisRateLimitEnabled default=false")
	}

	// Invalid DB falls back to 0
	t.Setenv("REDIS_DB", "invalid")
	if got := RedisDB(); got != 0 {
		t.Fatalf("RedisDB invalid fallback=%d", got)
	}
}

func TestKafkaConfig(t *testing.T) {
	t.Setenv("KAFKA_ENABLED", "true")
	t.Setenv("KAFKA_BROKERS", "kafka:9092, kafka2:9092")
	t.Setenv("KAFKA_CLIENT_ID", "my-client")
	t.Setenv("KAFKA_TOPIC_TASKS", "my-topic")

	if !KafkaEnabled() {
		t.Fatal("KafkaEnabled=true")
	}
	b := KafkaBrokers()
	if len(b) != 2 || b[0] != "kafka:9092" || b[1] != "kafka2:9092" {
		t.Fatalf("KafkaBrokers=%v", b)
	}
	if got := KafkaClientID(); got != "my-client" {
		t.Fatalf("KafkaClientID=%q", got)
	}
	if got := KafkaTopicTasks(); got != "my-topic" {
		t.Fatalf("KafkaTopicTasks=%q", got)
	}

	// Defaults
	t.Setenv("KAFKA_ENABLED", "")
	t.Setenv("KAFKA_BROKERS", "")
	t.Setenv("KAFKA_CLIENT_ID", "")
	t.Setenv("KAFKA_TOPIC_TASKS", "")

	if KafkaEnabled() {
		t.Fatal("KafkaEnabled default=false")
	}
	if b := KafkaBrokers(); b != nil {
		t.Fatalf("KafkaBrokers default=%v", b)
	}
	if got := KafkaClientID(); got != "mudro" {
		t.Fatalf("KafkaClientID default=%q", got)
	}
	if got := KafkaTopicTasks(); got != "mudro.agent.tasks.v1" {
		t.Fatalf("KafkaTopicTasks default=%q", got)
	}
}

func TestChatHubBackend(t *testing.T) {
	t.Setenv("CHAT_HUB_BACKEND", "redis")
	if got := ChatHubBackend(); got != "redis" {
		t.Fatalf("ChatHubBackend=%q", got)
	}

	t.Setenv("CHAT_HUB_BACKEND", "memory")
	if got := ChatHubBackend(); got != "memory" {
		t.Fatalf("ChatHubBackend=%q", got)
	}

	t.Setenv("CHAT_HUB_BACKEND", "unknown")
	if got := ChatHubBackend(); got != "memory" {
		t.Fatalf("ChatHubBackend unknown fallback=%q", got)
	}

	t.Setenv("CHAT_HUB_BACKEND", "")
	if got := ChatHubBackend(); got != "memory" {
		t.Fatalf("ChatHubBackend default=%q", got)
	}
}

func TestCORSAllowedOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.com, https://b.com")
	o := CORSAllowedOrigins()
	if len(o) != 2 || o[0] != "https://a.com" || o[1] != "https://b.com" {
		t.Fatalf("CORSAllowedOrigins=%v", o)
	}

	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	if CORSAllowedOrigins() != nil {
		t.Fatal("CORSAllowedOrigins default=nil")
	}
}

func TestMudroLocalRootAndSkaro(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MUDRO_LOCAL_ROOT", root)

	if got := MudroLocalRoot(); got != root {
		t.Fatalf("MudroLocalRoot=%q", got)
	}
	if got := SkaroLocalRoot(); got != filepath.Join(root, "skaro") {
		t.Fatalf("SkaroLocalRoot=%q", got)
	}

	t.Setenv("MUDRO_LOCAL_ROOT", "")
	if MudroLocalRoot() != "" {
		t.Fatalf("MudroLocalRoot default=%q", MudroLocalRoot())
	}
	if SkaroLocalRoot() != "" {
		t.Fatalf("SkaroLocalRoot default=%q", SkaroLocalRoot())
	}
}

func TestClaudeProxyAndUsagePaths(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MUDRO_LOCAL_ROOT", root)
	t.Setenv("MUDRO_CLAUDE_PROXY_URL", "http://proxy:8080/")
	t.Setenv("MUDRO_CLAUDE_USAGE_LOG", "/custom/usage.jsonl")
	t.Setenv("MUDRO_CLAUDE_TOKEN_USAGE", "/custom/token.yaml")

	if got := ClaudeProxyURL(); got != "http://proxy:8080" {
		t.Fatalf("ClaudeProxyURL=%q", got)
	}
	if got := ClaudeUsageLogPath(); got != "/custom/usage.jsonl" {
		t.Fatalf("ClaudeUsageLogPath=%q", got)
	}
	if got := ClaudeTokenUsagePath(); got != "/custom/token.yaml" {
		t.Fatalf("ClaudeTokenUsagePath=%q", got)
	}

	// Defaults when MUDRO_LOCAL_ROOT is set
	t.Setenv("MUDRO_CLAUDE_PROXY_URL", "")
	t.Setenv("MUDRO_CLAUDE_USAGE_LOG", "")
	t.Setenv("MUDRO_CLAUDE_TOKEN_USAGE", "")

	if got := ClaudeProxyURL(); got != "http://127.0.0.1:8788" {
		t.Fatalf("ClaudeProxyURL default=%q", got)
	}
	if got := ClaudeUsageLogPath(); got != filepath.Join(root, "skaro", "usage_log.jsonl") {
		t.Fatalf("ClaudeUsageLogPath default=%q", got)
	}
	if got := ClaudeTokenUsagePath(); got != filepath.Join(root, "skaro", "token_usage.yaml") {
		t.Fatalf("ClaudeTokenUsagePath default=%q", got)
	}

	// Empty when MUDRO_LOCAL_ROOT not set
	t.Setenv("MUDRO_LOCAL_ROOT", "")
	if ClaudeUsageLogPath() != "" {
		t.Fatalf("ClaudeUsageLogPath empty root=%q", ClaudeUsageLogPath())
	}
	if ClaudeTokenUsagePath() != "" {
		t.Fatalf("ClaudeTokenUsagePath empty root=%q", ClaudeTokenUsagePath())
	}
}

func TestSkaroPaths(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MUDRO_LOCAL_ROOT", root)
	t.Setenv("SKARO_DASHBOARD_URL", "http://skaro:4700/dashboard/")

	if got := SkaroDashboardURL(); got != "http://skaro:4700/dashboard" {
		t.Fatalf("SkaroDashboardURL=%q", got)
	}
	if got := SkaroProfilePath(); got != filepath.Join(root, "skaro", "profile.json") {
		t.Fatalf("SkaroProfilePath=%q", got)
	}

	t.Setenv("MUDRO_LOCAL_ROOT", "")
	if SkaroProfilePath() != "" {
		t.Fatalf("SkaroProfilePath empty root=%q", SkaroProfilePath())
	}
}

func TestIsProductionLikeEnv(t *testing.T) {
	for _, v := range []string{"prod", "production", "stage", "staging"} {
		if !isProductionLikeEnv(v) {
			t.Fatalf("isProductionLikeEnv(%q)=false, want true", v)
		}
	}
	for _, v := range []string{"dev", "development", "test", "local", ""} {
		if isProductionLikeEnv(v) {
			t.Fatalf("isProductionLikeEnv(%q)=true, want false", v)
		}
	}
}

func TestIsLocalDevDBHost(t *testing.T) {
	for _, v := range []string{"localhost", "127.0.0.1", "::1", "db"} {
		if !isLocalDevDBHost(v) {
			t.Fatalf("isLocalDevDBHost(%q)=false, want true", v)
		}
	}
	for _, v := range []string{"10.0.0.5", "postgres.example.com", "192.168.1.1"} {
		if isLocalDevDBHost(v) {
			t.Fatalf("isLocalDevDBHost(%q)=true, want false", v)
		}
	}
}