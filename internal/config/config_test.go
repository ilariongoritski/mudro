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
