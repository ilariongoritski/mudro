package casino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DefaultAddr               = ":8081"
	DefaultDSN                = ""
	DefaultInitialCoins       = 500
	DefaultRTPBasisPoint      = 9500
	DefaultMaxBet             = 1000
	DefaultRouletteBettingMS  = 15000
	DefaultRouletteLockMS     = 1500
	DefaultRouletteSpinMS     = 4500
	DefaultRouletteResultMS   = 3500
	DefaultBonusFreeSpins     = 10
	DefaultTelegramAPIBaseURL = "https://api.telegram.org"
)

func Addr() string {
	if v := strings.TrimSpace(os.Getenv("CASINO_ADDR")); v != "" {
		return v
	}
	return DefaultAddr
}

func DSN() string {
	if v := strings.TrimSpace(os.Getenv("CASINO_DSN")); v != "" {
		return v
	}
	return DefaultDSN
}

func MainDSN() string {
	if v := strings.TrimSpace(os.Getenv("CASINO_MAIN_DSN")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("DATABASE_URL")); v != "" {
		return v
	}
	return ""
}

// InternalSecret returns the shared secret for internal service-to-service auth.
// When set, all non-health endpoints require X-Internal-Secret header.
func InternalSecret() string {
	return strings.TrimSpace(os.Getenv("CASINO_INTERNAL_SECRET"))
}

func InitialCoins() int64 {
	if v := strings.TrimSpace(os.Getenv("CASINO_START_BALANCE")); v != "" {
		if n, ok := parsePositiveInt64(v); ok {
			return n
		}
	}
	if v := strings.TrimSpace(os.Getenv("CASINO_INITIAL_COINS")); v != "" {
		if n, ok := parsePositiveInt64(v); ok {
			return n
		}
	}
	return DefaultInitialCoins
}

func RTPPercent() float64 {
	if v := strings.TrimSpace(os.Getenv("CASINO_RTP_BPS")); v != "" {
		if n, ok := parsePositiveInt64(v); ok {
			return float64(n) / 100
		}
	}
	return float64(DefaultRTPBasisPoint) / 100
}

func MaxBet() int64 {
	if v := strings.TrimSpace(os.Getenv("CASINO_MAX_BET")); v != "" {
		if n, ok := parsePositiveInt64(v); ok {
			return n
		}
	}
	return DefaultMaxBet
}

func RouletteBettingDuration() time.Duration {
	return durationFromEnv("CASINO_ROULETTE_BETTING_MS", DefaultRouletteBettingMS)
}

func RouletteLockDuration() time.Duration {
	return durationFromEnv("CASINO_ROULETTE_LOCK_MS", DefaultRouletteLockMS)
}

func RouletteSpinDuration() time.Duration {
	return durationFromEnv("CASINO_ROULETTE_SPIN_MS", DefaultRouletteSpinMS)
}

func RouletteResultDuration() time.Duration {
	return durationFromEnv("CASINO_ROULETTE_RESULT_MS", DefaultRouletteResultMS)
}

func BonusFreeSpins() int64 {
	if v := strings.TrimSpace(os.Getenv("CASINO_BONUS_FREE_SPINS")); v != "" {
		if n, ok := parsePositiveInt64(v); ok {
			return n
		}
	}
	return DefaultBonusFreeSpins
}

func BonusTelegramBotToken() string {
	if v := strings.TrimSpace(os.Getenv("CASINO_BONUS_TELEGRAM_BOT_TOKEN")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("CASINO_BOT_TOKEN")); v != "" {
		return v
	}
	return ""
}

func BonusTelegramChannel() string {
	if v := strings.TrimSpace(os.Getenv("CASINO_BONUS_TELEGRAM_CHANNEL")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("CASINO_BONUS_CHANNEL")); v != "" {
		return v
	}
	return ""
}

func BonusTelegramAPIBaseURL() string {
	if v := strings.TrimSpace(os.Getenv("CASINO_BONUS_TELEGRAM_API_BASE")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return DefaultTelegramAPIBaseURL
}

func OpenPool(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(DSN())
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 6
	return pgxpool.NewWithConfig(ctx, cfg)
}

func OpenMainPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := MainDSN()
	if strings.TrimSpace(dsn) == "" {
		return nil, nil
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 4
	return pgxpool.NewWithConfig(ctx, cfg)
}

func DefaultConfig() Config {
	return Config{
		RTPPercent:     RTPPercent(),
		InitialBalance: InitialCoins(),
		SymbolWeights: map[string]int{
			"cherry":  30,
			"lemon":   24,
			"bar":     18,
			"seven":   10,
			"diamond": 6,
		},
		Paytable: map[string]int64{
			"cherry":  2,
			"lemon":   3,
			"bar":     5,
			"seven":   10,
			"diamond": 20,
		},
	}
}

func ValidateConfig(cfg Config) error {
	if cfg.RTPPercent <= 0 || cfg.RTPPercent > 300 {
		return errors.New("rtp_percent must be between 0 and 300")
	}
	if cfg.InitialBalance <= 0 {
		return errors.New("initial_balance must be positive")
	}
	if len(cfg.SymbolWeights) == 0 {
		return errors.New("symbol_weights must not be empty")
	}
	if len(cfg.Paytable) == 0 {
		return errors.New("paytable must not be empty")
	}
	for symbol, weight := range cfg.SymbolWeights {
		if strings.TrimSpace(symbol) == "" || weight <= 0 {
			return fmt.Errorf("invalid weight for symbol %q", symbol)
		}
	}
	for symbol, payout := range cfg.Paytable {
		if strings.TrimSpace(symbol) == "" || payout <= 0 {
			return fmt.Errorf("invalid payout for symbol %q", symbol)
		}
	}
	return nil
}

func marshalJSON(value any) []byte {
	raw, _ := json.Marshal(value)
	return raw
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

func durationFromEnv(name string, fallbackMS int64) time.Duration {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		if n, ok := parsePositiveInt64(v); ok {
			return time.Duration(n) * time.Millisecond
		}
	}
	return time.Duration(fallbackMS) * time.Millisecond
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
