import { createBrowserRouter, RouterProvider as RRProvider, Navigate } from 'react-router-dom'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'

import { FeedPage } from '@/pages/feed-page/ui/FeedPage'
import { LoginPage } from '@/pages/login-page/ui/LoginPage'
import { RegisterPage } from '@/pages/register-page/ui/RegisterPage'


const PublicRoute = ({ children }: { children: React.ReactNode }) => {
  const isAuthenticated = useAppSelector((state) => state.session.isAuthenticated)
  
  if (isAuthenticated) {
    return <Navigate to="/" replace />
  }
  return <>{children}</>
}

const AdminRoute = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated, user } = useAppSelector((state) => state.session)
  if (!isAuthenticated) return <Navigate to="/login" replace />
  if (user?.role !== 'admin') return <Navigate to="/" replace />
  return <>{children}</>
}

import { AdminPage } from '@/pages/admin-page/ui/AdminPage'
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
      element: <FeedPage />,
    },
    {
      path: '/admin',
      element: (
        <AdminRoute>
          <AdminPage />
        </AdminRoute>
      ),
    },
    {
      path: '*',
      element: <NotFoundPage />
    }
  ])

  return <RRProvider router={router} />
}
