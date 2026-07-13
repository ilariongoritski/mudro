import { mudroApi } from '@/shared/api/mudroApi'

export interface AdminUser {
  id: number | string
  username?: string
  email?: string
  role: string
}

export interface AdminUsersResponse {
  status?: string
  users?: AdminUser[]
}

export interface AdminStats {
  total_users?: number
  active_subscriptions?: number
}

export interface RuntimeProvider {
  name: string
  configured: boolean
  model?: string
  limit?: string
}

export interface RuntimeLimits {
  requests_per_second?: string
  burst?: string
}

export interface RuntimeService {
  name: string
  status: 'healthy' | 'unavailable' | 'unknown'
}

export interface RuntimeDashboard {
  providers: RuntimeProvider[]
  limits: RuntimeLimits
  services: RuntimeService[]
}

export const adminApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    getAdminUsers: build.query<AdminUsersResponse, void>({
      query: () => '/admin/users',
      providesTags: ['Auth'],
    }),
    getAdminStats: build.query<AdminStats, void>({
      query: () => '/admin/stats',
      providesTags: ['Auth'],
    }),
    getRuntimeDashboard: build.query<RuntimeDashboard, void>({
      query: () => '/admin/runtime',
      providesTags: ['Auth'],
    }),
  }),
})

export const { useGetAdminUsersQuery, useGetAdminStatsQuery, useGetRuntimeDashboardQuery } = adminApi
