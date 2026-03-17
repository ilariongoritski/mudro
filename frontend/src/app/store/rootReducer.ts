import { combineReducers } from '@reduxjs/toolkit'

import { feedFiltersReducer } from '@/features/feed-controls/model/feedFiltersSlice'
import { sessionReducer } from '@/entities/session/model/sessionSlice'
import { mudroApi } from '@/shared/api/mudroApi'

export const rootReducer = combineReducers({
  session: sessionReducer,
  feedFilters: feedFiltersReducer,
  [mudroApi.reducerPath]: mudroApi.reducer,
})
