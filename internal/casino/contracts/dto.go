package contracts

import "encoding/json"

type AuthRequest struct {
	InitData string `json:"initData"`
}

type BetRequest struct {
	RoundID        string  `json:"roundId"`
	BetAmount      float64 `json:"betAmount"`
	ClientSeed     string  `json:"clientSeed"`
	IdempotencyKey string  `json:"idempotencyKey"`
}

type FaucetRequest struct {
	Amount float64 `json:"amount"`
}

type UpsertRTPProfileRequest struct {
	Name      string          `json:"name"`
	Rtp       float64         `json:"rtp"`
	Paytable  json.RawMessage `json:"paytable"`
	IsDefault bool            `json:"isDefault"`
}