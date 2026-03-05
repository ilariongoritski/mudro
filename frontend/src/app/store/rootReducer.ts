import { combineReducers } from '@reduxjs/toolkit'

import { feedFiltersReducer } from '@/features/feed-controls/model/feedFiltersSlice'
import { mudroApi } from '@/shared/api/mudroApi'

export const rootReducer = combineReducers({
  feedFilters: feedFiltersReducer,
  [mudroApi.reducerPath]: mudroApi.reducer,
})
