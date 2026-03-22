import { mudroApi } from '@/shared/api/mudroApi'
import type { User } from '../model/sessionSlice'

export const authApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    login: build.mutation<{ token: string; user: User }, any>({
      query: (credentials) => ({
        url: '/auth/login',
        method: 'POST',
        body: credentials,
      }),
    }),
    register: build.mutation<any, any>({
      query: (credentials) => ({
        url: '/auth/register',
        method: 'POST',
        body: credentials,
      }),
    }),
    me: build.query<User, void>({
      query: () => '/auth/me',
    }),
  }),
})

export const { useLoginMutation, useRegisterMutation, useMeQuery } = authApi
