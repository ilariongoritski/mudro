import { mudroApi } from '@/shared/api/mudroApi'

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

interface RawCasinoActivityItem {
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

interface RawCasinoLiveFeedItem {
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

interface RawCasinoReactionItem {
  activity_id: number
  emoji: string
  count: number
  player: RawCasinoPlayerBadge
  game_type: string
  net_result: number
  created_at: string
  latest_at: string
}

interface RawCasinoProfileResponse {
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
  bet_type: RouletteBetType
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

interface RawRouletteHistoryItem {
  round_id?: number | string
  id?: number | string
  winning_number?: number | null
  winning_color?: RouletteColor
  resolved_at?: string | null
  created_at?: string | null
}

interface RawRouletteStateResponse {
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

interface RawRouletteBetResponse {
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

interface RawBonusStateResponse {
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

interface RawBonusClaimResponse {
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

const normalizeActivityItem = (item: RawCasinoActivityItem): CasinoActivityItem => ({
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

const normalizeLiveFeedItem = (item: RawCasinoLiveFeedItem): CasinoLiveFeedItem => ({
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

const normalizeReactionItem = (item: RawCasinoReactionItem): CasinoReactionItem => ({
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

const normalizeRouletteBetItem = (item: RawRouletteBetItem): RouletteBetItem => ({
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

const normalizeRouletteHistoryItem = (
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

const normalizeRouletteBetResponse = (raw: RawRouletteBetResponse): RouletteBetResponse => ({
  balance: raw.balance,
  round_id: raw.round_id != null ? String(raw.round_id) : undefined,
  status: raw.status ?? 'accepted',
  accepted_bets: raw.accepted_bets ?? raw.bets?.length ?? 0,
  state: normalizeRouletteStateResponse(raw.state) ?? undefined,
})

const normalizeProfile = (profile: RawCasinoProfileResponse): CasinoProfileResponse => ({
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

const normalizeBonusState = (state?: RawBonusStateResponse | null): BonusStateResponse => {
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

const normalizeBonusClaimResponse = (response: RawBonusClaimResponse): BonusStateResponse =>
  normalizeBonusState(
    response.state
      ? {
          ...response.state,
          claim_status: response.state.claim_status ?? response.status,
          verification_status: response.state.verification_status ?? response.verification_status,
        }
      : undefined,
  )

export const casinoApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    getCasinoBalance: build.query<CasinoBalanceResponse, void>({
      query: () => ({
        url: '/casino/balance',
        cache: 'no-store',
      }),
      providesTags: ['Casino'],
    }),
    getCasinoHistory: build.query<CasinoHistoryResponse, number | void>({
      query: (limit = 8) => ({
        url: '/casino/history',
        cache: 'no-store',
        params: { limit },
      }),
      providesTags: ['Casino'],
    }),
    getCasinoProfile: build.query<CasinoProfileResponse, void>({
      query: () => ({
        url: '/casino/profile',
        cache: 'no-store',
      }),
      transformResponse: (response: RawCasinoProfileResponse) => normalizeProfile(response),
      providesTags: ['Casino'],
    }),
    getCasinoActivity: build.query<{ items: CasinoActivityItem[] }, number | void>({
      query: (limit = 10) => ({
        url: '/casino/activity',
        cache: 'no-store',
        params: { limit },
      }),
      transformResponse: (response: { items?: RawCasinoActivityItem[] }) => ({
        items: (response.items ?? []).map(normalizeActivityItem),
      }),
      providesTags: ['Casino'],
    }),
    getCasinoLiveFeed: build.query<CasinoLiveFeedResponse, number | void>({
      query: (limit = 12) => ({
        url: '/casino/live-feed',
        cache: 'no-store',
        params: { limit },
      }),
      transformResponse: (response: { items?: RawCasinoLiveFeedItem[] }) => ({
        items: (response.items ?? []).map(normalizeLiveFeedItem),
      }),
      providesTags: ['Casino'],
    }),
    getCasinoTopWins: build.query<CasinoLiveFeedResponse, number | void>({
      query: (limit = 6) => ({
        url: '/casino/top-wins',
        cache: 'no-store',
        params: { limit },
      }),
      transformResponse: (response: { items?: RawCasinoLiveFeedItem[] }) => ({
        items: (response.items ?? []).map(normalizeLiveFeedItem),
      }),
      providesTags: ['Casino'],
    }),
    getCasinoReactions: build.query<CasinoReactionListResponse, number | void>({
      query: (limit = 8) => ({
        url: '/casino/reactions',
        cache: 'no-store',
        params: { limit },
      }),
      transformResponse: (response: { items?: RawCasinoReactionItem[] }) => ({
        items: (response.items ?? []).map(normalizeReactionItem),
      }),
      providesTags: ['Casino'],
    }),
    addCasinoReaction: build.mutation<CasinoReactionItem, CasinoReactionRequest>({
      query: (body) => ({
        url: '/casino/reactions',
        method: 'POST',
        body,
      }),
      transformResponse: (response: RawCasinoReactionItem) => normalizeReactionItem(response),
      invalidatesTags: ['Casino'],
    }),
    getCasinoConfig: build.query<CasinoConfigResponse, void>({
      query: () => ({
        url: '/casino/config',
        cache: 'no-store',
      }),
      providesTags: ['Casino'],
    }),
    getRouletteState: build.query<RouletteStateResponse, void>({
      query: () => ({
        url: '/casino/roulette/state',
        cache: 'no-store',
      }),
      transformResponse: (response: RawRouletteStateResponse) => normalizeRouletteStateResponse(response) ?? {
        round_id: '',
        phase: 'idle',
      },
      providesTags: ['Casino'],
    }),
    getRouletteHistory: build.query<RouletteHistoryResponse, number | void>({
      query: (limit = 12) => ({
        url: '/casino/roulette/history',
        cache: 'no-store',
        params: { limit },
      }),
      transformResponse: (response: { items?: RawRouletteHistoryItem[] }) => ({
        items: (response.items ?? []).map((item) => normalizeRouletteHistoryItem(item, 'result')),
      }),
      providesTags: ['Casino'],
    }),
    placeRouletteBets: build.mutation<RouletteBetResponse, RouletteBetRequest>({
      query: (body) => ({
        url: '/casino/roulette/bets',
        method: 'POST',
        body: {
          round_id: body.round_id ? Number.parseInt(body.round_id, 10) : undefined,
          bets: body.bets.map((bet) => ({
            bet_type: bet.bet_type,
            bet_value: bet.bet_value != null ? String(bet.bet_value) : '',
            stake: bet.stake,
          })),
        },
      }),
      transformResponse: (response: RawRouletteBetResponse) => normalizeRouletteBetResponse(response),
      invalidatesTags: ['Casino'],
    }),
    getPlinkoConfig: build.query<PlinkoConfigResponse, void>({
      query: () => ({
        url: '/casino/plinko/config',
        cache: 'no-store',
      }),
      providesTags: ['Casino'],
    }),
    getPlinkoState: build.query<PlinkoStateResponse, void>({
      query: () => ({
        url: '/casino/plinko/state',
        cache: 'no-store',
      }),
      providesTags: ['Casino'],
    }),
    dropPlinko: build.mutation<PlinkoDropResponse, PlinkoDropRequest>({
      query: (body) => ({
        url: '/casino/plinko/drop',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Casino'],
    }),
    getCasinoBonusState: build.query<BonusStateResponse, void>({
      query: () => ({
        url: '/casino/bonus/state',
        cache: 'no-store',
      }),
      transformResponse: (response: RawBonusStateResponse) => normalizeBonusState(response),
      providesTags: ['Casino'],
    }),
    claimCasinoBonusSubscription: build.mutation<BonusStateResponse, { initData?: string | null } | void>({
      query: (payload) => {
        const initData = typeof payload === 'object' ? payload?.initData?.trim() ?? '' : ''
        return {
          url: '/casino/bonus/claim-subscription',
          method: 'POST',
          headers: initData
            ? {
                'X-Telegram-Init-Data': initData,
              }
            : undefined,
          body: initData
            ? {
                init_data: initData,
              }
            : undefined,
        }
      },
      transformResponse: (response: RawBonusClaimResponse) => normalizeBonusClaimResponse(response),
      invalidatesTags: ['Casino'],
    }),
    updateCasinoConfig: build.mutation<{ status: string }, CasinoConfigResponse>({
      query: (body) => ({
        url: '/casino/config',
        method: 'PUT',
        body,
      }),
      invalidatesTags: ['Casino'],
    }),
    spinCasino: build.mutation<CasinoSpinResponse, CasinoSpinRequest>({
      query: (body) => ({
        url: '/casino/spin',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Casino'],
    }),
    getBlackjackState: build.query<BlackjackStateResponse | { status: 'no_game' }, void>({
      query: () => ({
        url: '/casino/blackjack/state',
        cache: 'no-store',
      }),
      providesTags: ['Casino'],
    }),
    startBlackjack: build.mutation<BlackjackStateResponse, BlackjackStartRequest>({
      query: (body) => ({
        url: '/casino/blackjack/start',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Casino'],
    }),
    actionBlackjack: build.mutation<BlackjackStateResponse, BlackjackActionRequest>({
      query: (body) => ({
        url: '/casino/blackjack/action',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Casino'],
    }),
  }),
})

export const {
  useGetCasinoBalanceQuery,
  useGetCasinoHistoryQuery,
  useGetCasinoProfileQuery,
  useGetCasinoActivityQuery,
  useGetCasinoLiveFeedQuery,
  useGetCasinoTopWinsQuery,
  useGetCasinoReactionsQuery,
  useAddCasinoReactionMutation,
  useGetCasinoConfigQuery,
  useGetRouletteStateQuery,
  useGetRouletteHistoryQuery,
  usePlaceRouletteBetsMutation,
  useGetPlinkoConfigQuery,
  useGetPlinkoStateQuery,
  useDropPlinkoMutation,
  useGetCasinoBonusStateQuery,
  useClaimCasinoBonusSubscriptionMutation,
  useUpdateCasinoConfigMutation,
  useSpinCasinoMutation,
  useGetBlackjackStateQuery,
  useStartBlackjackMutation,
  useActionBlackjackMutation,
} = casinoApi
