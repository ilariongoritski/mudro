package casino

import (
	"time"
)

const (
	HouseAccountCode      = "SYSTEM_HOUSE_POOL"
	SettlementAccountCode = "SYSTEM_SETTLEMENT_POOL"
)

type Account struct {
	ID       string
	UserID   string
	Code     string
	Currency string
	Balance  float64
}

type Round struct {
	ID             string
	UserID         string
	ServerSeed     string
	ServerSeedHash string
	ClientSeed     string
	Nonce          int
	RoundHash      string
	Roll           *int
	BetAmount      *float64
	PayoutAmount   *float64
	Multiplier     *float64
	TierLabel      *string
	Status         string
	CreatedAt      time.Time
}

type IdempotencyKey struct {
	ID           string
	Status       string
	ResponseJSON []byte
}

type BetInput struct {
	UserID         string
	RoundID        string
	BetAmount      float64
	ClientSeed     string
	IdempotencyKey string
}

type BetResult struct {
	RoundID        string  `json:"roundId"`
	Roll           int     `json:"roll"`
	BetAmount      float64 `json:"betAmount"`
	PayoutAmount   float64 `json:"payoutAmount"`
	Multiplier     float64 `json:"multiplier"`
	TierLabel      string  `json:"tierLabel"`
	TierSymbol     string  `json:"tierSymbol"`
	Balance        float64 `json:"balance"`
	Nonce          int     `json:"nonce"`
	RoundHash      string  `json:"roundHash"`
	ServerSeedHash string  `json:"serverSeedHash"`
	ServerSeed     string  `json:"serverSeed"`
	FromCache      bool    `json:"fromCache"`
}

type FaucetResult struct {
	Amount  float64 `json:"amount"`
	Balance float64 `json:"balance"`
}