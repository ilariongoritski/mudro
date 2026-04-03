import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { BaseQueryFn, FetchArgs, FetchBaseQueryError } from '@reduxjs/toolkit/query'
import type { RootState } from '@/app/store'
import { env } from '@/shared/config/env'

const rawBaseQuery = fetchBaseQuery({
  baseUrl: env.apiBaseUrl,
  prepareHeaders: (headers, { getState }) => {
    const token = (getState() as RootState).session?.token;
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
  const result = await rawBaseQuery(args, api, extraOptions)

  // No refresh-token endpoint exists on the backend.
  // On 401, immediately clear local session so the user is redirected to login.
  if (result.error && result.error.status === 401) {
    const { logout } = await import('@/entities/session/model/sessionSlice')
    api.dispatch(logout())
  }

  return result
}

export const mudroApi = createApi({
  reducerPath: 'mudroApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Feed', 'Auth', 'Casino', 'Chat'],
  endpoints: () => ({}),
})
