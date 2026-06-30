export interface CasinoBalanceResponse {
  balance: number
  free_spins_balance?: number
  bonus_claimed?: boolean
  currency?: string
  rtp?: number
}

export interface CasinoPlayerBadge {
  user_id: number
  username: string
  display_name?: string | null
  avatar_url?: string | null
}

export interface CasinoActivityItem {
  id: number | string
  game_type: string
  game_ref?: string | null
  bet_amount: number
  payout_amount: number
  net_result: number
  status: string
  metadata?: Record<string, unknown>
  created_at: string
}

export interface CasinoHistoryItem {
  id: number | string
  bet: number
  win: number
  symbols: string[]
  created_at: string
}

export interface CasinoHistoryResponse {
  items: CasinoHistoryItem[]
}

export interface CasinoProfileResponse {
  user_id: number
  username: string
  display_name?: string | null
  avatar_url?: string | null
  balance: number
  total_wagered: number
  total_won: number
  games_played: number
  roulette_rounds_played: number
  level: number
  xp_progress: number
  progress_target?: number
  next_level_xp: number
  last_game_at?: string | null
  recent_activity?: CasinoActivityItem[]
}

export interface CasinoLiveFeedItem {
  id: number | string
  player: CasinoPlayerBadge
  game_type: string
  event_type: string
  game_ref?: string | null
  bet_amount: number
  payout_amount: number
  net_result: number
  status: string
  metadata?: Record<string, unknown>
  created_at: string
}

export interface CasinoLiveFeedResponse {
  items: CasinoLiveFeedItem[]
}

export interface CasinoReactionItem {
  activity_id: number
  emoji: string
  count: number
  player: CasinoPlayerBadge
  game_type: string
  net_result: number
  created_at: string
  latest_at: string
}

export interface CasinoReactionListResponse {
  items: CasinoReactionItem[]
}

export interface CasinoReactionRequest {
  activity_id: number
  emoji: string
}

export interface CasinoConfigResponse {
  rtp_percent: number
  initial_balance: number
  symbol_weights: Record<string, number>
  paytable: Record<string, number>
  updated_at: string
}

export type RoulettePhase = 'betting' | 'locking' | 'spinning' | 'result' | 'idle'
export type RouletteBetType = 'straight' | 'red' | 'black' | 'green' | 'odd' | 'even' | 'low' | 'high'
export type RouletteColor = 'red' | 'black' | 'green' | 'unknown'

export interface RouletteRoundHistoryItem {
  round_id: string
  winning_number: number | null
  winning_color: RouletteColor
  phase: RoulettePhase
  betting_opens_at?: string | null
  betting_closes_at?: string | null
  spin_started_at?: string | null
  resolved_at?: string | null
  created_at?: string | null
}

export interface RouletteBetItem {
  id: number | string
  round_id: string
  user_id: number
  bet_type: RouletteBetType
  bet_value?: number | null
  stake: number
  payout_amount?: number | null
  status: string
  created_at: string
}

export interface RouletteStateResponse {
  round_id: string
  phase: RoulettePhase
  server_time?: string | null
  betting_opens_at?: string | null
  betting_closes_at?: string | null
  spin_started_at?: string | null
  resolved_at?: string | null
  winning_number?: number | null
  winning_color?: RouletteColor
  display_sequence?: number[]
  result_sequence?: number[]
  history?: RouletteRoundHistoryItem[]
  my_bets?: RouletteBetItem[]
}

export interface RouletteHistoryResponse {
  items: RouletteRoundHistoryItem[]
}

export interface RouletteBetRequestItem {
  bet_type: RouletteBetType
  bet_value?: number | null
  stake: number
}

export interface RouletteBetRequest {
  round_id?: string
  bets: RouletteBetRequestItem[]
}

export interface RouletteBetResponse {
  balance?: number
  round_id?: string
  status?: string
  accepted_bets?: number
  state?: RouletteStateResponse
}

export interface RouletteInstantSpinResponse {
  winning_number: number
  winning_color: RouletteColor
  display_sequence: number[]
  result_sequence: number[]
  payout_amount: number
  balance: number
  bets: RouletteBetItem[]
}

export interface CasinoSpinRequest {
  bet: number
}

export interface CasinoSpinResponse {
  balance: number
  free_spins_balance?: number
  free_spin_used?: boolean
  win: number
  symbols: string[]
  multiplier?: number
}

export type PlinkoRisk = 'low' | 'medium' | 'high'

export interface PlinkoConfigResponse {
  rows: number
  slots: number
  min_bet: number
  max_bet: number
  multipliers: Record<PlinkoRisk, number[]>
}

export interface PlinkoStateResponse {
  config: PlinkoConfigResponse
  balance: number
}

export interface PlinkoDropRequest {
  bet: number
  risk: PlinkoRisk
}

export interface PlinkoDropResponse {
  balance: number
  bet: number
  risk: PlinkoRisk
  path: number[]
  rows: number
  slot_index: number
  multiplier: number
  payout: number
  net_result: number
  status: string
  created_at: string
}

export interface BlackjackCard {
  suit: 'hearts' | 'diamonds' | 'clubs' | 'spades'
  rank: string
  value: number
}

export interface BlackjackHand {
  cards: BlackjackCard[]
  score: number
  is_bust: boolean
}

export type BlackjackStatus = 'player_turn' | 'dealer_turn' | 'resolved'

export interface BlackjackStateResponse {
  id: number
  user_id: number
  bet: number
  player_hand: BlackjackHand
  dealer_hand: BlackjackHand
  status: BlackjackStatus
  winner?: 'player' | 'dealer' | 'push' | null
  payout: number
  created_at: string
}

export interface BlackjackStartRequest {
  bet: number
}

export interface BlackjackActionRequest {
  action: 'hit' | 'stand'
}

export interface BonusHistoryItem {
  id: string | number
  title: string
  amount?: number | null
  status?: string
  created_at?: string | null
}

export interface BonusStateResponse {
  subscription_required?: boolean
  subscribed?: boolean
  free_spins_total?: number
  free_spins_available?: number
  claim_ready?: boolean
  claim_status?: string
  bonus_label?: string
  fair_strip?: string
  history?: BonusHistoryItem[]
  verification_status?: string
  verification_message?: string
  telegram_channel?: string
  claimed?: boolean
}
