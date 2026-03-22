import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { RootState } from '@/app/store'
import { env } from '@/shared/config/env'

export const mudroApi = createApi({
  reducerPath: 'mudroApi',
  baseQuery: fetchBaseQuery({ 
    baseUrl: env.apiBaseUrl,
    prepareHeaders: (headers, { getState }) => {
      const token = (getState() as RootState).session?.token;
      if (token) {
        headers.set('authorization', `Bearer ${token}`)
      }
      return headers
    },
  }),
  tagTypes: ['Feed', 'Auth'],
  endpoints: () => ({}),
})
