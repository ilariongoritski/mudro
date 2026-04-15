import { createSlice } from '@reduxjs/toolkit'
import type { PayloadAction } from '@reduxjs/toolkit'

export interface User {
  id: number
  username: string
  email?: string
  role?: string
  avatar_url?: string
  isPremium?: boolean
}

export interface SessionState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  isInitialized: boolean
}

const initialState: SessionState = {
  token: null,
  user: null,
  isAuthenticated: false,
  isInitialized: false,
}

export const sessionSlice = createSlice({
  name: 'session',
  initialState,
  reducers: {
    setSession: (state: SessionState, action: PayloadAction<{ user: User; token: string } | null>) => {
      if (action.payload) {
        state.user = action.payload.user
        state.token = action.payload.token
        state.isAuthenticated = true
      } else {
        state.user = null
        state.token = null
        state.isAuthenticated = false
      }
      state.isInitialized = true
    },
    setCredentials: (state: SessionState, action: PayloadAction<{ user: User; token: string }>) => {
      state.user = action.payload.user
      state.token = action.payload.token
      state.isAuthenticated = true
      state.isInitialized = true
    },
    setIsInitialized: (state: SessionState, action: PayloadAction<boolean>) => {
      state.isInitialized = action.payload
    },
    logout: (state: SessionState) => {
      state.user = null
      state.token = null
      state.isAuthenticated = false
    },
  },
})

export const { setSession, setCredentials, setIsInitialized, logout } = sessionSlice.actions
export const sessionReducer = sessionSlice.reducer
