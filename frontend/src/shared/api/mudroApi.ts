import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

import { env } from '@/shared/config/env'

export const mudroApi = createApi({
  reducerPath: 'mudroApi',
  baseQuery: fetchBaseQuery({ baseUrl: env.apiBaseUrl }),
  tagTypes: ['Feed'],
  endpoints: () => ({}),
})
