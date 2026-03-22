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
	DefaultAddr          = ":8081"
	DefaultDSN           = "postgres://postgres:postgres@localhost:5434/mudro_casino?sslmode=disable"
	DefaultInitialCoins  = 2500
	DefaultRTPBasisPoint = 9500
	DefaultMaxBet        = 1000
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

func InitialCoins() int64 {
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

func OpenPool(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(DSN())
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 6
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

func nowUTC() time.Time {
	return time.Now().UTC()
}
