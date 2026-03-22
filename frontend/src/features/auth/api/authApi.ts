import { mudroApi } from '@/shared/api/mudroApi'
import type { AuthUser } from '@/features/auth/model/authSlice'

interface AuthResponse {
  id: number
  username: string
  token: string
}

interface RegisterArgs {
  username: string
  email: string
  password: string
}

interface LoginArgs {
  email: string
  password: string
}

export const authApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    register: build.mutation<AuthResponse, RegisterArgs>({
      query: (body) => ({ url: '/auth/register', method: 'POST', body }),
    }),
    login: build.mutation<AuthResponse, LoginArgs>({
      query: (body) => ({ url: '/auth/login', method: 'POST', body }),
    }),
    getMe: build.query<AuthUser, void>({
      query: () => '/auth/me',
      providesTags: ['Auth'],
    }),
  }),
})

export const { useRegisterMutation, useLoginMutation, useGetMeQuery } = authApi
