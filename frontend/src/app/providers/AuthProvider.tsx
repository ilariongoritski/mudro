import React, { useEffect } from 'react'
import type { User } from '@/entities/session/model/sessionSlice'
import { setCredentials, setIsInitialized, logout } from '@/entities/session/model/sessionSlice'
import { useAppDispatch, useAppSelector } from '@/shared/lib/hooks/storeHooks'

const SESSION_STORAGE_KEY = 'mudro.session'

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const dispatch = useAppDispatch()
  const session = useAppSelector((state) => state.session)

  useEffect(() => {
    if (typeof window === 'undefined') {
      dispatch(setIsInitialized(true))
      return
    }

    const raw = window.localStorage.getItem(SESSION_STORAGE_KEY)
    if (!raw) {
      dispatch(setIsInitialized(true))
      return
    }

    try {
      const stored = JSON.parse(raw) as { token?: string; user?: User }
      if (stored?.token && stored.user) {
        dispatch(setCredentials({ token: stored.token, user: stored.user }))
        return
      }
    } catch {
      window.localStorage.removeItem(SESSION_STORAGE_KEY)
    }

    dispatch(setIsInitialized(true))
  }, [dispatch])

  useEffect(() => {
    if (typeof window === 'undefined' || !session.isInitialized) {
      return
    }

    if (session.token && session.user) {
      window.localStorage.setItem(
        SESSION_STORAGE_KEY,
        JSON.stringify({
          token: session.token,
          user: session.user,
        }),
      )
      return
    }

    window.localStorage.removeItem(SESSION_STORAGE_KEY)
  }, [session.isInitialized, session.token, session.user])

  useEffect(() => {
    const handleUnauthorizedLogout = () => {
      dispatch(logout())
    }

    window.addEventListener('mudro:unauthorized', handleUnauthorizedLogout)
    return () => window.removeEventListener('mudro:unauthorized', handleUnauthorizedLogout)
  }, [dispatch])

  return <>{children}</>
}
