import React, { useEffect } from 'react'
import type { User } from '@/entities/session/model/sessionSlice'
import { setCredentials, setIsInitialized, setSession, logout } from '@/entities/session/model/sessionSlice'
import { env } from '@/shared/config/env'
import { useAppDispatch, useAppSelector } from '@/shared/lib/hooks/storeHooks'

const SESSION_STORAGE_KEY = 'mudro.session'

interface RawUser {
  id: number
  username: string
  email?: string | null
  role?: string
  isPremium?: boolean
  is_premium?: boolean
}

interface RawAuthResult {
  token: string
  user: RawUser
}

const normalizeUser = (user: RawUser): User => ({
  id: user.id,
  username: user.username,
  email: user.email ?? undefined,
  role: user.role,
  isPremium: user.isPremium ?? user.is_premium ?? false,
})

const authFetch = (path: string, token: string, init?: RequestInit) => {
  const headers = new Headers(init?.headers)
  headers.set('authorization', `Bearer ${token}`)

  return fetch(`${env.apiBaseUrl}${path}`, {
    ...init,
    headers,
  })
}

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const dispatch = useAppDispatch()
  const session = useAppSelector((state) => state.session)

  useEffect(() => {
    let isCancelled = false

    const finishEmptySession = () => {
      if (typeof window !== 'undefined') {
        window.localStorage.removeItem(SESSION_STORAGE_KEY)
      }
      if (!isCancelled) {
        dispatch(setSession(null))
      }
    }

    const initializeSession = async () => {
      if (typeof window === 'undefined') {
        dispatch(setIsInitialized(true))
        return
      }

      const raw = window.localStorage.getItem(SESSION_STORAGE_KEY)
      if (!raw) {
        dispatch(setIsInitialized(true))
        return
      }

      let stored: { token?: string }
      try {
        stored = JSON.parse(raw) as { token?: string }
      } catch {
        finishEmptySession()
        return
      }

      const token = stored.token
      if (!token) {
        finishEmptySession()
        return
      }

      try {
        const meResponse = await authFetch('/auth/me', token)
        if (meResponse.ok) {
          const user = normalizeUser((await meResponse.json()) as RawUser)
          if (!isCancelled) {
            dispatch(setCredentials({ token, user }))
          }
          return
        }

        const refreshResponse = await authFetch('/auth/refresh', token, { method: 'POST' })
        if (refreshResponse.ok) {
          const refreshed = (await refreshResponse.json()) as RawAuthResult
          if (!isCancelled) {
            dispatch(setCredentials({ token: refreshed.token, user: normalizeUser(refreshed.user) }))
          }
          return
        }
      } catch {
        // Treat an unverified restored token as logged out.
      }

      finishEmptySession()
    }

    void initializeSession()

    return () => {
      isCancelled = true
    }
  }, [dispatch])

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }
    if (!session.isInitialized) {
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
