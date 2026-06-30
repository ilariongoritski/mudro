import { mudroApi } from '@/shared/api/mudroApi'
import {
  normalizeActivityItem,
  normalizeBonusClaimResponse,
  normalizeBonusState,
  normalizeLiveFeedItem,
  normalizeProfile,
  normalizeReactionItem,
  normalizeRouletteBetItem,
  normalizeRouletteBetResponse,
  normalizeRouletteHistoryItem,
  normalizeRouletteStateResponse,
  type RawBonusClaimResponse,
  type RawBonusStateResponse,
  type RawCasinoActivityItem,
  type RawCasinoLiveFeedItem,
  type RawCasinoProfileResponse,
  type RawCasinoReactionItem,
  type RawRouletteBetResponse,
  type RawRouletteHistoryItem,
  type RawRouletteStateResponse,
} from '@/features/casino/api/casinoApi.normalizers'
import type {
  BlackjackActionRequest,
  BlackjackStartRequest,
  BlackjackStateResponse,
  BonusStateResponse,
  CasinoActivityItem,
  CasinoBalanceResponse,
  CasinoConfigResponse,
  CasinoHistoryResponse,
  CasinoLiveFeedResponse,
  CasinoProfileResponse,
  CasinoReactionItem,
  CasinoReactionListResponse,
  CasinoReactionRequest,
  CasinoSpinRequest,
  CasinoSpinResponse,
  PlinkoConfigResponse,
  PlinkoDropRequest,
  PlinkoDropResponse,
  PlinkoStateResponse,
  RouletteBetRequest,
  RouletteBetResponse,
  RouletteHistoryResponse,
  RouletteInstantSpinResponse,
  RouletteStateResponse,
} from '@/features/casino/api/casinoApi.types'

export type {
  BlackjackActionRequest,
  BlackjackCard,
  BlackjackHand,
  BlackjackStartRequest,
  BlackjackStateResponse,
  BlackjackStatus,
  BonusHistoryItem,
  BonusStateResponse,
  CasinoActivityItem,
  CasinoBalanceResponse,
  CasinoConfigResponse,
  CasinoHistoryItem,
  CasinoHistoryResponse,
  CasinoLiveFeedItem,
  CasinoLiveFeedResponse,
  CasinoPlayerBadge,
  CasinoProfileResponse,
  CasinoReactionItem,
  CasinoReactionListResponse,
  CasinoReactionRequest,
  CasinoSpinRequest,
  CasinoSpinResponse,
  PlinkoConfigResponse,
  PlinkoDropRequest,
  PlinkoDropResponse,
  PlinkoRisk,
  PlinkoStateResponse,
  RouletteBetItem,
  RouletteBetRequest,
  RouletteBetRequestItem,
  RouletteBetResponse,
  RouletteBetType,
  RouletteColor,
  RouletteHistoryResponse,
  RouletteInstantSpinResponse,
  RoulettePhase,
  RouletteRoundHistoryItem,
  RouletteStateResponse,
} from '@/features/casino/api/casinoApi.types'

export { normalizeRouletteStateResponse } from '@/features/casino/api/casinoApi.normalizers'

export const casinoApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    getCasinoBalance: build.query<CasinoBalanceResponse, void>({
      query: () => ({
        url: '/casino/balance',
        
      }),
      providesTags: ['Casino'],
    }),
    getCasinoHistory: build.query<CasinoHistoryResponse, number | void>({
      query: (limit = 8) => ({
        url: '/casino/history',
        
        params: { limit },
      }),
      providesTags: ['Casino'],
    }),
    getCasinoProfile: build.query<CasinoProfileResponse, void>({
      query: () => ({
        url: '/casino/profile',
        
      }),
      transformResponse: (response: RawCasinoProfileResponse) => normalizeProfile(response),
      providesTags: ['Casino'],
    }),
    getCasinoActivity: build.query<{ items: CasinoActivityItem[] }, number | void>({
      query: (limit = 10) => ({
        url: '/casino/activity',
        
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
        
      }),
      providesTags: ['Casino'],
    }),
    getRouletteState: build.query<RouletteStateResponse, void>({
      query: () => ({
        url: '/casino/roulette/state',
        
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
    instantRouletteSpin: build.mutation<RouletteInstantSpinResponse, RouletteBetRequest>({
      query: (body) => ({
        url: '/casino/roulette/instant-spin',
        method: 'POST',
        body: {
          bets: body.bets.map((bet) => ({
            bet_type: bet.bet_type,
            bet_value: bet.bet_value != null ? String(bet.bet_value) : '',
            stake: bet.stake,
          })),
        },
      }),
       transformResponse: (response: RouletteInstantSpinResponse) => ({
         ...response,
         bets: (response.bets ?? []).map(normalizeRouletteBetItem),
       }),
      invalidatesTags: ['Casino'],
    }),
    getPlinkoConfig: build.query<PlinkoConfigResponse, void>({
      query: () => ({
        url: '/casino/plinko/config',
        
      }),
      providesTags: ['Casino'],
    }),
    getPlinkoState: build.query<PlinkoStateResponse, void>({
      query: () => ({
        url: '/casino/plinko/state',
        
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
  useInstantRouletteSpinMutation,
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
