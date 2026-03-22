import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

import { env } from '@/shared/config/env'

export const mudroApi = createApi({
  reducerPath: 'mudroApi',
  baseQuery: fetchBaseQuery({
    baseUrl: env.apiBaseUrl,
    prepareHeaders: (headers) => {
      const token = localStorage.getItem('token')
      if (token) headers.set('Authorization', `Bearer ${token}`)
      return headers
    },
  }),
  tagTypes: ['Feed', 'Auth'],
  endpoints: () => ({}),
})
