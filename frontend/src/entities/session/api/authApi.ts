import { mudroApi } from '@/shared/api/mudroApi'
import type { User } from '../model/sessionSlice'

interface LoginCredentials {
  login?: string
  email?: string
  password: string
}

interface RegisterCredentials {
  login?: string
  username?: string
  email?: string
  password: string
}

interface RawUser {
  id: number
  username: string
  email?: string | null
  role: string
  isPremium?: boolean
  is_premium?: boolean
}

interface AuthResult {
  token: string
  user: User
}

interface RawAuthResult {
  token: string
  user: RawUser
}

const normalizeUser = (user: RawUser): User => ({
  id: user.id,
  username: user.username,
  email: user.email ?? undefined,
  role: user.role,
  isPremium: user.isPremium ?? user.is_premium ?? false,
})

const normalizeAuthResult = (result: RawAuthResult): AuthResult => ({
  token: result.token,
  user: normalizeUser(result.user),
})

export const authApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    login: build.mutation<AuthResult, LoginCredentials>({
      query: (credentials) => ({
        url: '/auth/login',
        method: 'POST',
        body: credentials,
      }),
      transformResponse: (response: RawAuthResult) => normalizeAuthResult(response),
    }),
    register: build.mutation<AuthResult, RegisterCredentials>({
      query: (credentials) => ({
        url: '/auth/register',
        method: 'POST',
        body: credentials,
      }),
      transformResponse: (response: RawAuthResult) => normalizeAuthResult(response),
    }),
    me: build.query<User, void>({
      query: () => '/auth/me',
      transformResponse: (response: RawUser) => normalizeUser(response),
    }),
    refresh: build.mutation<AuthResult, void>({
      query: () => ({
        url: '/auth/refresh',
        method: 'POST',
      }),
      transformResponse: (response: RawAuthResult) => normalizeAuthResult(response),
    }),
    telegramBootstrap: build.mutation<AuthResult, { initData: string }>({
      query: (payload) => ({
        url: '/auth/telegram',
        method: 'POST',
        body: payload,
      }),
      transformResponse: (response: RawAuthResult) => normalizeAuthResult(response),
    }),
  }),
})

export const { useLoginMutation, useRegisterMutation, useMeQuery, useRefreshMutation, useTelegramBootstrapMutation } = authApi
