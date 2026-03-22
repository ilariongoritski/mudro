import { mudroApi } from '@/shared/api/mudroApi'

export interface CasinoBalanceResponse {
  balance: number
  currency?: string
  rtp?: number
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

export interface CasinoConfigResponse {
  rtp_percent: number
  initial_balance: number
  symbol_weights: Record<string, number>
  paytable: Record<string, number>
  updated_at: string
}

export interface CasinoSpinRequest {
  bet: number
}

export interface CasinoSpinResponse {
  balance: number
  win: number
  symbols: string[]
  multiplier?: number
}

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
    getCasinoConfig: build.query<CasinoConfigResponse, void>({
      query: () => ({
        url: '/casino/config',
        cache: 'no-store',
      }),
      providesTags: ['Casino'],
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
  }),
})

export const {
  useGetCasinoBalanceQuery,
  useGetCasinoHistoryQuery,
  useGetCasinoConfigQuery,
  useUpdateCasinoConfigMutation,
  useSpinCasinoMutation,
} = casinoApi
