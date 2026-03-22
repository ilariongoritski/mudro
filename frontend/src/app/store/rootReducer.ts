import { combineReducers } from '@reduxjs/toolkit'

import { authReducer } from '@/features/auth/model/authSlice'
import { feedFiltersReducer } from '@/features/feed-controls/model/feedFiltersSlice'
import { mudroApi } from '@/shared/api/mudroApi'

export const rootReducer = combineReducers({
  auth: authReducer,
  feedFilters: feedFiltersReducer,
  [mudroApi.reducerPath]: mudroApi.reducer,
})
