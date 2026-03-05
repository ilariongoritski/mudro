import { createSlice, type PayloadAction } from '@reduxjs/toolkit'

import type { FeedSort, FeedSource } from '@/entities/post/model/types'

interface FeedFiltersState {
  source: FeedSource
  sort: FeedSort
  limit: number
}

const initialState: FeedFiltersState = {
  source: 'all',
  sort: 'desc',
  limit: 12,
}

const feedFiltersSlice = createSlice({
  name: 'feedFilters',
  initialState,
  reducers: {
    setSource: (state, action: PayloadAction<FeedSource>) => {
      state.source = action.payload
    },
    setSort: (state, action: PayloadAction<FeedSort>) => {
      state.sort = action.payload
    },
    setLimit: (state, action: PayloadAction<number>) => {
      state.limit = action.payload
    },
  },
})

export const { setSource, setSort, setLimit } = feedFiltersSlice.actions
export const feedFiltersReducer = feedFiltersSlice.reducer
