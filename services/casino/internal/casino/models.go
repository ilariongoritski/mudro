package casino

import "time"

type ParticipantInput struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
}

type Config struct {
	RTPPercent     float64          `json:"rtp_percent"`
	InitialBalance int64            `json:"initial_balance"`
	SymbolWeights  map[string]int   `json:"symbol_weights"`
	Paytable       map[string]int64 `json:"paytable"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type Participant struct {
	UserID     int64      `json:"user_id"`
	Username   string     `json:"username,omitempty"`
	Email      string     `json:"email,omitempty"`
	Role       string     `json:"role,omitempty"`
	Coins      int64      `json:"coins"`
	SpinsCount int64      `json:"spins_count"`
	LastSpinAt *time.Time `json:"last_spin_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type SpinRecord struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Bet       int64     `json:"bet"`
	Win       int64     `json:"win"`
	Symbols   []string  `json:"symbols"`
	CreatedAt time.Time `json:"created_at"`
}

type SpinResult struct {
	Balance int64        `json:"balance"`
	Config  Config       `json:"config,omitempty"`
	History []SpinRecord `json:"history,omitempty"`
	Symbols []string     `json:"symbols,omitempty"`
	Win     int64        `json:"win,omitempty"`
}
