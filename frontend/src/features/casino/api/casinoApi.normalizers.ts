import type {
  BonusHistoryItem,
  BonusStateResponse,
  CasinoActivityItem,
  CasinoLiveFeedItem,
  CasinoPlayerBadge,
  CasinoProfileResponse,
  CasinoReactionItem,
  RouletteBetItem,
  RouletteBetResponse,
  RouletteColor,
  RoulettePhase,
  RouletteRoundHistoryItem,
  RouletteStateResponse,
} from '@/features/casino/api/casinoApi.types'

export interface RawCasinoActivityItem {
  id: number | string
  game_type: string
  game_ref?: string | null
  bet_amount: number
  payout_amount: number
  net_result: number
  status: string
  metadata?: unknown
  created_at: string
}

interface RawCasinoPlayerBadge {
  user_id: number
  username: string
  display_name?: string | null
  avatar_url?: string | null
}

export interface RawCasinoLiveFeedItem {
  id: number | string
  player: RawCasinoPlayerBadge
  game_type: string
  event_type: string
  game_ref?: string | null
  bet_amount: number
  payout_amount: number
  net_result: number
  status: string
  metadata?: unknown
  created_at: string
}

export interface RawCasinoReactionItem {
  activity_id: number
  emoji: string
  count: number
  player: RawCasinoPlayerBadge
  game_type: string
  net_result: number
  created_at: string
  latest_at: string
}

export interface RawCasinoProfileResponse {
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
  progress_current?: number
  xp_progress?: number
  progress_target?: number
  next_level_xp?: number
  last_game_at?: string | null
  recent_activity?: RawCasinoActivityItem[]
}

interface RawRouletteBetItem {
  id: number | string
  round_id: number | string
  user_id: number
  bet_type: RouletteBetItem['bet_type']
  bet_value?: string | number | null
  stake: number
  payout_amount?: number | null
  status: string
  created_at: string
}

interface RawRouletteRound {
  id: number | string
  status: RoulettePhase
  winning_number?: number | null
  winning_color?: RouletteColor
  display_sequence?: number[]
  result_sequence?: number[]
  betting_opens_at?: string | null
  betting_closes_at?: string | null
  spin_started_at?: string | null
  resolved_at?: string | null
  created_at?: string | null
}

export interface RawRouletteHistoryItem {
  round_id?: number | string
  id?: number | string
  winning_number?: number | null
  winning_color?: RouletteColor
  resolved_at?: string | null
  created_at?: string | null
}

export interface RawRouletteStateResponse {
  round?: RawRouletteRound
  phase?: RoulettePhase
  server_time?: string | null
  recent_results?: RawRouletteHistoryItem[]
  history?: RawRouletteHistoryItem[]
  my_bets?: RawRouletteBetItem[]
  round_id?: number | string
  betting_opens_at?: string | null
  betting_closes_at?: string | null
  spin_started_at?: string | null
  resolved_at?: string | null
  winning_number?: number | null
  winning_color?: RouletteColor
  display_sequence?: number[]
  result_sequence?: number[]
}

export interface RawRouletteBetResponse {
  balance?: number
  round_id?: number | string
  bets?: RawRouletteBetItem[]
  status?: string
  accepted_bets?: number
  state?: RawRouletteStateResponse
}

interface RawBonusClaimItem {
  id: number | string
  status?: string
  verification_status?: string
  verification_message?: string
  free_spins_granted?: number
  telegram_channel?: string
  claimed_at?: string | null
  created_at?: string | null
}

export interface RawBonusStateResponse {
  user_id?: number
  claimed?: boolean
  claim_status?: string
  verification_status?: string
  verification_message?: string
  free_spins_balance?: number
  free_spins_granted?: number
  telegram_channel?: string
  recent_claims?: RawBonusClaimItem[]
}

export interface RawBonusClaimResponse {
  status?: string
  verification_status?: string
  state?: RawBonusStateResponse
}

const normalizeMetadata = (metadata: unknown): Record<string, unknown> | undefined => {
  if (!metadata) {
    return undefined
  }
  if (typeof metadata === 'object' && !Array.isArray(metadata)) {
    return metadata as Record<string, unknown>
  }
  if (typeof metadata === 'string') {
    try {
      const parsed = JSON.parse(metadata)
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        return parsed as Record<string, unknown>
      }
    } catch {
      return { raw: metadata }
    }
  }
  return undefined
}

export const normalizeActivityItem = (item: RawCasinoActivityItem): CasinoActivityItem => ({
  id: item.id,
  game_type: item.game_type,
  game_ref: item.game_ref ?? null,
  bet_amount: item.bet_amount,
  payout_amount: item.payout_amount,
  net_result: item.net_result,
  status: item.status,
  metadata: normalizeMetadata(item.metadata),
  created_at: item.created_at,
})

const normalizePlayerBadge = (player: RawCasinoPlayerBadge): CasinoPlayerBadge => ({
  user_id: player.user_id,
  username: player.username,
  display_name: player.display_name ?? player.username,
  avatar_url: player.avatar_url ?? null,
})

export const normalizeLiveFeedItem = (item: RawCasinoLiveFeedItem): CasinoLiveFeedItem => ({
  id: item.id,
  player: normalizePlayerBadge(item.player),
  game_type: item.game_type,
  event_type: item.event_type,
  game_ref: item.game_ref ?? null,
  bet_amount: item.bet_amount,
  payout_amount: item.payout_amount,
  net_result: item.net_result,
  status: item.status,
  metadata: normalizeMetadata(item.metadata),
  created_at: item.created_at,
})

export const normalizeReactionItem = (item: RawCasinoReactionItem): CasinoReactionItem => ({
  activity_id: item.activity_id,
  emoji: item.emoji,
  count: item.count,
  player: normalizePlayerBadge(item.player),
  game_type: item.game_type,
  net_result: item.net_result,
  created_at: item.created_at,
  latest_at: item.latest_at,
})

const parseBetValue = (value: string | number | null | undefined): number | null => {
  if (typeof value === 'number') {
    return Number.isFinite(value) ? value : null
  }
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number.parseInt(value, 10)
    return Number.isFinite(parsed) ? parsed : null
  }
  return null
}

export const normalizeRouletteBetItem = (item: RawRouletteBetItem): RouletteBetItem => ({
  id: item.id,
  round_id: String(item.round_id),
  user_id: item.user_id,
  bet_type: item.bet_type,
  bet_value: parseBetValue(item.bet_value),
  stake: item.stake,
  payout_amount: item.payout_amount ?? 0,
  status: item.status,
  created_at: item.created_at,
})

export const normalizeRouletteHistoryItem = (
  item: RawRouletteHistoryItem,
  fallbackPhase: RoulettePhase = 'result',
): RouletteRoundHistoryItem => ({
  round_id: String(item.round_id ?? item.id ?? ''),
  winning_number: item.winning_number ?? null,
  winning_color: item.winning_color ?? 'unknown',
  phase: fallbackPhase,
  resolved_at: item.resolved_at ?? null,
  created_at: item.created_at ?? item.resolved_at ?? null,
})

export const normalizeRouletteStateResponse = (raw: RawRouletteStateResponse | null | undefined): RouletteStateResponse | null => {
  if (!raw) {
    return null
  }

  const round = raw.round
  const phase = raw.phase ?? round?.status ?? 'idle'
  const historySource = raw.history ?? raw.recent_results ?? []

  return {
    round_id: String(raw.round_id ?? round?.id ?? ''),
    phase,
    server_time: raw.server_time ?? null,
    betting_opens_at: raw.betting_opens_at ?? round?.betting_opens_at ?? null,
    betting_closes_at: raw.betting_closes_at ?? round?.betting_closes_at ?? null,
    spin_started_at: raw.spin_started_at ?? round?.spin_started_at ?? null,
    resolved_at: raw.resolved_at ?? round?.resolved_at ?? null,
    winning_number: raw.winning_number ?? round?.winning_number ?? null,
    winning_color: raw.winning_color ?? round?.winning_color ?? 'unknown',
    display_sequence: raw.display_sequence ?? round?.display_sequence ?? [],
    result_sequence: raw.result_sequence ?? round?.result_sequence ?? [],
    history: historySource.map((item) => normalizeRouletteHistoryItem(item, 'result')),
    my_bets: (raw.my_bets ?? []).map(normalizeRouletteBetItem),
  }
}

export const normalizeRouletteBetResponse = (raw: RawRouletteBetResponse): RouletteBetResponse => ({
  balance: raw.balance,
  round_id: raw.round_id != null ? String(raw.round_id) : undefined,
  status: raw.status ?? 'accepted',
  accepted_bets: raw.accepted_bets ?? raw.bets?.length ?? 0,
  state: normalizeRouletteStateResponse(raw.state) ?? undefined,
})

export const normalizeProfile = (profile: RawCasinoProfileResponse): CasinoProfileResponse => ({
  user_id: profile.user_id,
  username: profile.username,
  display_name: profile.display_name ?? null,
  avatar_url: profile.avatar_url ?? null,
  balance: profile.balance,
  total_wagered: profile.total_wagered,
  total_won: profile.total_won,
  games_played: profile.games_played,
  roulette_rounds_played: profile.roulette_rounds_played,
  level: profile.level,
  xp_progress: profile.xp_progress ?? profile.progress_current ?? 0,
  progress_target: profile.progress_target ?? 1000,
  next_level_xp: profile.next_level_xp ?? profile.progress_target ?? 1000,
  last_game_at: profile.last_game_at ?? null,
  recent_activity: (profile.recent_activity ?? []).map(normalizeActivityItem),
})

const normalizeBonusHistoryItem = (item: RawBonusClaimItem): BonusHistoryItem => ({
  id: item.id,
  title: item.status === 'claimed' ? 'Подписка подтверждена' : 'Subscription bonus',
  amount: item.free_spins_granted ?? null,
  status: item.verification_status ?? item.status,
  created_at: item.claimed_at ?? item.created_at ?? null,
})

export const normalizeBonusState = (state?: RawBonusStateResponse | null): BonusStateResponse => {
  const verificationStatus = state?.verification_status ?? 'verification_required'
  const claimed = Boolean(state?.claimed)
  const freeSpinsAvailable = state?.free_spins_balance ?? 0
  const freeSpinsTotal = Math.max(state?.free_spins_granted ?? 0, freeSpinsAvailable, 10)
  const telegramChannel = state?.telegram_channel?.trim() || undefined

  return {
    subscription_required: !claimed,
    subscribed: verificationStatus === 'verified' || claimed,
    free_spins_total: freeSpinsTotal,
    free_spins_available: freeSpinsAvailable,
    claim_ready: !claimed,
    claim_status: state?.claim_status ?? (claimed ? 'claimed' : 'available'),
    bonus_label: `+${freeSpinsTotal} free spins`,
    fair_strip: `Provably fair: subscription-bonus-public · ${telegramChannel ?? 'telegram-channel'} · ${verificationStatus}`,
    history: (state?.recent_claims ?? []).map(normalizeBonusHistoryItem),
    verification_status: verificationStatus,
    verification_message: state?.verification_message ?? undefined,
    telegram_channel: telegramChannel,
    claimed,
  }
}

export const normalizeBonusClaimResponse = (response: RawBonusClaimResponse): BonusStateResponse =>
  normalizeBonusState(
    response.state
      ? {
          ...response.state,
          claim_status: response.state.claim_status ?? response.status,
          verification_status: response.state.verification_status ?? response.verification_status,
        }
      : undefined,
  )
