import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { BaseQueryFn, FetchArgs, FetchBaseQueryError } from '@reduxjs/toolkit/query'
import type { RootState } from '@/app/store'
import { setCredentials, logout } from '@/entities/session/model/sessionSlice'
import { env } from '@/shared/config/env'

const rawBaseQuery = fetchBaseQuery({
  baseUrl: env.apiBaseUrl,
  prepareHeaders: (headers, { getState }) => {
    const token = (getState() as RootState).session?.token
    if (token) {
      headers.set('authorization', `Bearer ${token}`)
    }
    return headers
  },
})

const baseQueryWithReauth: BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError> = async (
  args,
  api,
  extraOptions,
) => {
  let result = await rawBaseQuery(args, api, extraOptions)

  if (result.error && result.error.status === 401) {
    const token = (api.getState() as RootState).session?.token
    if (token) {
      const refreshResult = await rawBaseQuery(
        { url: '/auth/refresh', method: 'POST' },
        api,
        extraOptions,
      )
      if (refreshResult.data) {
        const data = refreshResult.data as {
          token: string
          user: {
            id: number
            username: string
            email?: string | null
            role: string
            isPremium?: boolean
            is_premium?: boolean
          }
        }
        api.dispatch(setCredentials({
          token: data.token,
          user: {
            id: data.user.id,
            username: data.user.username,
            email: data.user.email ?? undefined,
            role: data.user.role,
            isPremium: data.user.isPremium ?? data.user.is_premium ?? false,
          },
        }))
        // Retry the original request
        result = await rawBaseQuery(args, api, extraOptions)
      } else {
        api.dispatch(logout())
      }
    }
  }

  return result
}

export const mudroApi = createApi({
  reducerPath: 'mudroApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Feed', 'Auth', 'Casino', 'Chat'],
  endpoints: () => ({}),
})
