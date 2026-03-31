import { useEffect } from 'react'
import { AppRouterProvider } from '@/app/providers/RouterProvider'
import { ErrorBoundary } from '@/shared/ui/ErrorBoundary'
import { useAppDispatch, useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { useMeQuery } from '@/entities/session/api/authApi'
import { setCredentials, logout } from '@/entities/session/model/sessionSlice'
import '@/shared/ui/ErrorBoundary.css'

const AuthInitializer = ({ children }: { children: React.ReactNode }) => {
  const dispatch = useAppDispatch()
  const token = useAppSelector((state) => state.session.token)
  const { data: user, error } = useMeQuery(undefined, { skip: !token })

  useEffect(() => {
    if (user && token) {
      dispatch(setCredentials({ user, token }))
    }
    if (error) {
      dispatch(logout())
    }
  }, [user, error, token, dispatch])

  return <>{children}</>
}

export const App = () => {
  return (
    <ErrorBoundary>
      <AuthInitializer>
        <AppRouterProvider />
      </AuthInitializer>
    </ErrorBoundary>
  )
}
