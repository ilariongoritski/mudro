import { mudroApi } from '@/shared/api/mudroApi'

export interface AdminUser {
  id: number | string
  username?: string
  email: string
  role: string
}

export interface AdminUsersResponse {
  status?: string
  users?: AdminUser[]
}

export interface AdminStats {
  active_subscriptions?: number
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
  }),
})

export const { useGetAdminUsersQuery, useGetAdminStatsQuery } = adminApi
