package casino

import (
	"os"
	"strconv"
	"strings"
)

const (
	DefaultCasinoAddr   = ":8082"
	DefaultStartBalance = 10000.0
	DefaultFaucetAmount = 1000.0
	DefaultAdminKey     = ""
	DefaultBotToken     = ""
)

func CasinoAddr() string {
	return envOr("CASINO_ADDR", DefaultCasinoAddr)
}

func CasinoBotToken() string {
	return envOr("CASINO_BOT_TOKEN", DefaultBotToken)
}

func CasinoDemoMode() bool {
	return parseBool(os.Getenv("CASINO_DEMO_MODE"))
}

func CasinoAdminKey() string {
	return envOr("CASINO_ADMIN_KEY", DefaultAdminKey)
}

func CasinoStartBalance() float64 {
	return parseFloat(os.Getenv("CASINO_START_BALANCE"), DefaultStartBalance)
}

func CasinoFaucetAmount() float64 {
	return parseFloat(os.Getenv("CASINO_FAUCET_AMOUNT"), DefaultFaucetAmount)
}

func AllowedOrigins() []string {
	raw := strings.TrimSpace(os.Getenv("CASINO_ALLOWED_ORIGINS"))
	if raw == "" {
		return []string{"http://localhost:5173", "http://localhost:8080"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func parseBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

func parseFloat(v string, def float64) float64 {
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil || f <= 0 {
		return def
	}
	return f
}
