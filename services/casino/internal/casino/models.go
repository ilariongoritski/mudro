package casino

import (
	"encoding/json"
	"time"
)

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

type ActivityItem struct {
	ID           int64           `json:"id"`
	GameType     string          `json:"game_type"`
	GameRef      string          `json:"game_ref"`
	BetAmount    int64           `json:"bet_amount"`
	PayoutAmount int64           `json:"payout_amount"`
	NetResult    int64           `json:"net_result"`
	Status       string          `json:"status"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type ActivityList struct {
	Items []ActivityItem `json:"items"`
}

type PlayerBadge struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type LiveFeedItem struct {
	ID           int64           `json:"id"`
	Player       PlayerBadge     `json:"player"`
	GameType     string          `json:"game_type"`
	EventType    string          `json:"event_type"`
	GameRef      string          `json:"game_ref"`
	BetAmount    int64           `json:"bet_amount"`
	PayoutAmount int64           `json:"payout_amount"`
	NetResult    int64           `json:"net_result"`
	Status       string          `json:"status"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type LiveFeedResponse struct {
	Items []LiveFeedItem `json:"items"`
}

type ReactionFeedItem struct {
	ActivityID int64       `json:"activity_id"`
	Emoji      string      `json:"emoji"`
	Count      int64       `json:"count"`
	Player     PlayerBadge `json:"player"`
	GameType   string      `json:"game_type"`
	NetResult  int64       `json:"net_result"`
	CreatedAt  time.Time   `json:"created_at"`
	LatestAt   time.Time   `json:"latest_at"`
}

type ReactionList struct {
	Items []ReactionFeedItem `json:"items"`
}

type ReactionRequest struct {
	ActivityID int64  `json:"activity_id"`
	Emoji      string `json:"emoji"`
}

type PlayerProfile struct {
	UserID               int64          `json:"user_id"`
	Username             string         `json:"username"`
	DisplayName          string         `json:"display_name,omitempty"`
	AvatarURL            string         `json:"avatar_url,omitempty"`
	Balance              int64          `json:"balance"`
	FreeSpinsBalance     int64          `json:"free_spins_balance"`
	BonusClaimStatus     string         `json:"bonus_claim_status,omitempty"`
	BonusClaimedAt       *time.Time     `json:"bonus_claimed_at,omitempty"`
	BonusVerifiedAt      *time.Time     `json:"bonus_verified_at,omitempty"`
	TotalWagered         int64          `json:"total_wagered"`
	TotalWon             int64          `json:"total_won"`
	GamesPlayed          int64          `json:"games_played"`
	RouletteRoundsPlayed int64          `json:"roulette_rounds_played"`
	Level                int64          `json:"level"`
	XPProgress           int64          `json:"xp_progress"`
	ProgressTarget       int64          `json:"progress_target"`
	LastGameAt           *time.Time     `json:"last_game_at,omitempty"`
	ClientSeed           string         `json:"client_seed"`
	CurrentNonce         int            `json:"current_nonce"`
	ServerSeedHash       string         `json:"server_seed_hash"`
	RecentActivity       []ActivityItem `json:"recent_activity"`
}

type BonusKind string

const BonusKindSubscription BonusKind = "subscription"

type BonusState struct {
	UserID             int64            `json:"user_id"`
	BonusKind          BonusKind        `json:"bonus_kind"`
	Claimed            bool             `json:"claimed"`
	ClaimStatus        string           `json:"claim_status"`
	VerificationStatus string           `json:"verification_status"`
	VerificationMsg    string           `json:"verification_message,omitempty"`
	FreeSpinsBalance   int64            `json:"free_spins_balance"`
	FreeSpinsGranted   int64            `json:"free_spins_granted"`
	ClaimedAt          *time.Time       `json:"claimed_at,omitempty"`
	VerifiedAt         *time.Time       `json:"verified_at,omitempty"`
	TelegramUserID     *int64           `json:"telegram_user_id,omitempty"`
	TelegramUsername   string           `json:"telegram_username,omitempty"`
	TelegramChannel    string           `json:"telegram_channel,omitempty"`
	RecentClaims       []BonusClaimItem `json:"recent_claims,omitempty"`
}

type BonusClaimRequest struct {
	InitData         string `json:"init_data,omitempty"`
	TelegramInitData string `json:"telegram_init_data,omitempty"`
}

type BonusClaimResponse struct {
	Status             string     `json:"status"`
	VerificationStatus string     `json:"verification_status"`
	State              BonusState `json:"state"`
}

type BonusClaimItem struct {
	ID                 int64           `json:"id"`
	UserID             int64           `json:"user_id"`
	BonusKind          BonusKind       `json:"bonus_kind"`
	Status             string          `json:"status"`
	VerificationStatus string          `json:"verification_status"`
	VerificationMsg    string          `json:"verification_message,omitempty"`
	FreeSpinsGranted   int64           `json:"free_spins_granted"`
	TelegramUserID     *int64          `json:"telegram_user_id,omitempty"`
	TelegramUsername   string          `json:"telegram_username,omitempty"`
	TelegramChannel    string          `json:"telegram_channel,omitempty"`
	Metadata           json.RawMessage `json:"metadata,omitempty"`
	ClaimedAt          time.Time       `json:"claimed_at"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type BonusClaimList struct {
	Items []BonusClaimItem `json:"items"`
}

type RoulettePhase string

const (
	RoulettePhaseBetting  RoulettePhase = "betting"
	RoulettePhaseLocking  RoulettePhase = "locking"
	RoulettePhaseSpinning RoulettePhase = "spinning"
	RoulettePhaseResult   RoulettePhase = "result"
)

type RouletteBetInput struct {
	BetType  string `json:"bet_type"`
	BetValue string `json:"bet_value"`
	Stake    int64  `json:"stake"`
}

type RouletteBet struct {
	ID           int64     `json:"id"`
	RoundID      int64     `json:"round_id"`
	UserID       int64     `json:"user_id"`
	BetType      string    `json:"bet_type"`
	BetValue     string    `json:"bet_value"`
	Stake        int64     `json:"stake"`
	PayoutAmount int64     `json:"payout_amount"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type RouletteRound struct {
	ID              int64         `json:"id"`
	Status          RoulettePhase `json:"status"`
	WinningNumber   *int          `json:"winning_number,omitempty"`
	WinningColor    string        `json:"winning_color,omitempty"`
	DisplaySequence []int         `json:"display_sequence,omitempty"`
	ResultSequence  []int         `json:"result_sequence,omitempty"`
	BettingOpensAt  time.Time     `json:"betting_opens_at"`
	BettingClosesAt time.Time     `json:"betting_closes_at"`
	SpinStartedAt   *time.Time    `json:"spin_started_at,omitempty"`
	ResolvedAt      *time.Time    `json:"resolved_at,omitempty"`
	ServerSeed      string        `json:"-"`
	ServerSeedHash  string        `json:"server_seed_hash,omitempty"`
	ClientSeed      string        `json:"client_seed,omitempty"`
	Nonce           int64         `json:"nonce,omitempty"`
	RoundHash       string        `json:"round_hash,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
}

type RouletteHistoryItem struct {
	RoundID       int64     `json:"round_id"`
	WinningNumber int       `json:"winning_number"`
	WinningColor  string    `json:"winning_color"`
	ResolvedAt    time.Time `json:"resolved_at"`
}

type RouletteState struct {
	Round         RouletteRound         `json:"round"`
	Phase         RoulettePhase         `json:"phase"`
	ServerTime    time.Time             `json:"server_time"`
	SecondsLeft   int64                 `json:"seconds_left"`
	RecentResults []RouletteHistoryItem `json:"recent_results"`
	MyBets        []RouletteBet         `json:"my_bets,omitempty"`
}

type RoulettePlaceBetsRequest struct {
	RoundID int64              `json:"round_id,omitempty"`
	Bets    []RouletteBetInput `json:"bets"`
}

type RoulettePlaceBetsResponse struct {
	RoundID int64         `json:"round_id"`
	Balance int64         `json:"balance"`
	Bets    []RouletteBet `json:"bets"`
}

type PlinkoRisk string

const (
	PlinkoRiskLow    PlinkoRisk = "low"
	PlinkoRiskMedium PlinkoRisk = "medium"
	PlinkoRiskHigh   PlinkoRisk = "high"
)

type PlinkoConfig struct {
	Rows        int                      `json:"rows"`
	Slots       int                      `json:"slots"`
	MinBet      int64                    `json:"min_bet"`
	MaxBet      int64                    `json:"max_bet"`
	Multipliers map[PlinkoRisk][]float64 `json:"multipliers"`
}

type PlinkoState struct {
	Config  PlinkoConfig `json:"config"`
	Balance int64        `json:"balance"`
}

type PlinkoDropRequest struct {
	Bet  int64      `json:"bet"`
	Risk PlinkoRisk `json:"risk"`
}

type PlinkoDropResult struct {
	Balance    int64      `json:"balance"`
	Bet        int64      `json:"bet"`
	Risk       PlinkoRisk `json:"risk"`
	Path       []int      `json:"path"`
	Rows       int        `json:"rows"`
	SlotIndex  int        `json:"slot_index"`
	Multiplier float64    `json:"multiplier"`
	Payout     int64      `json:"payout"`
	NetResult  int64      `json:"net_result"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
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
	Balance          int64        `json:"balance"`
	FreeSpinsBalance int64        `json:"free_spins_balance,omitempty"`
	FreeSpinUsed     bool         `json:"free_spin_used,omitempty"`
	Config           Config       `json:"config,omitempty"`
	History          []SpinRecord `json:"history,omitempty"`
	Symbols          []string     `json:"symbols,omitempty"`
	Win              int64        `json:"win,omitempty"`
}

type BlackjackCard struct {
	Suit  string `json:"suit"`
	Rank  string `json:"rank"`
	Value int    `json:"value"`
}

type BlackjackHand struct {
	Cards  []BlackjackCard `json:"cards"`
	Score  int             `json:"score"`
	IsBust bool            `json:"is_bust"`
}

type BlackjackStatus string

const (
	BlackjackStatusPlayerTurn BlackjackStatus = "player_turn"
	BlackjackStatusDealerTurn BlackjackStatus = "dealer_turn"
	BlackjackStatusResolved   BlackjackStatus = "resolved"
)

type BlackjackState struct {
	ID         int64           `json:"id"`
	UserID     int64           `json:"user_id"`
	Bet        int64           `json:"bet"`
	PlayerHand BlackjackHand   `json:"player_hand"`
	DealerHand BlackjackHand   `json:"dealer_hand"`
	Status     BlackjackStatus `json:"status"`
	Winner     string          `json:"winner,omitempty"` // "player", "dealer", "push"
	Payout     int64           `json:"payout"`
	CreatedAt  time.Time       `json:"created_at"`
	ServerSeed string          `json:"-"`
	ClientSeed string          `json:"-"`
	Nonce      int64           `json:"-"`
}

type BlackjackAction string

const (
	BlackjackActionHit   BlackjackAction = "hit"
	BlackjackActionStand BlackjackAction = "stand"
)

type BlackjackGameRequest struct {
	Bet int64 `json:"bet"`
}

type RouletteInstantSpinRequest struct {
	Bets []RouletteBetInput `json:"bets"`
}

type RouletteInstantSpinResponse struct {
	WinningNumber   int           `json:"winning_number"`
	WinningColor    string        `json:"winning_color"`
	DisplaySequence []int         `json:"display_sequence"`
	ResultSequence  []int         `json:"result_sequence"`
	PayoutAmount    int64         `json:"payout_amount"`
	Balance         int64         `json:"balance"`
	Bets            []RouletteBet `json:"bets"`
}
