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

interface AuthResult {
  token: string
  user: User
}

export const authApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    login: build.mutation<AuthResult, LoginCredentials>({
      query: (credentials) => ({
        url: '/auth/login',
        method: 'POST',
        body: credentials,
      }),
    }),
    register: build.mutation<AuthResult, RegisterCredentials>({
      query: (credentials) => ({
        url: '/auth/register',
        method: 'POST',
        body: credentials,
      }),
    }),
    me: build.query<User, void>({
      query: () => '/auth/me',
    }),
    telegramBootstrap: build.mutation<AuthResult, { initData: string }>({
      query: (payload) => ({
        url: '/auth/telegram',
        method: 'POST',
        body: payload,
      }),
    }),
  }),
})

export const { useLoginMutation, useRegisterMutation, useMeQuery, useTelegramBootstrapMutation } = authApi
