import { createBrowserRouter, RouterProvider as RRProvider, Navigate } from 'react-router-dom'
import { useSelector } from 'react-redux'

import type { RootState } from '@/app/store'
import { FeedPage } from '@/pages/feed-page/ui/FeedPage'
import { LoginPage } from '@/pages/login-page/ui/LoginPage'
import { RegisterPage } from '@/pages/register-page/ui/RegisterPage'

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const isAuthenticated = useSelector((state: RootState) => state.session.isAuthenticated)
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

const PublicRoute = ({ children }: { children: React.ReactNode }) => {
  const isAuthenticated = useSelector((state: RootState) => state.session.isAuthenticated)
  
  if (isAuthenticated) {
    return <Navigate to="/" replace />
  }
  return <>{children}</>
}

import { NotFoundPage } from '@/pages/not-found-page/ui/NotFoundPage'

export const AppRouterProvider = () => {
  const router = createBrowserRouter([
    {
      path: '/login',
      element: (
        <PublicRoute>
          <LoginPage />
        </PublicRoute>
      ),
    },
    {
      path: '/register',
      element: (
        <PublicRoute>
          <RegisterPage />
        </PublicRoute>
      ),
    },
    {
      path: '/',
      element: (
        <ProtectedRoute>
          <FeedPage />
        </ProtectedRoute>
      ),
    },
    {
      path: '*',
      element: <NotFoundPage />
    }
  ])

  return <RRProvider router={router} />
}
