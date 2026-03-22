import { lazy, Suspense, type ReactNode } from 'react'
import { createBrowserRouter, RouterProvider as RRProvider, Navigate } from 'react-router-dom'

import { useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { ErrorBoundary } from '@/shared/ui/ErrorBoundary'

import { FeedPage } from '@/pages/feed-page/ui/FeedPage'
import { LoginPage } from '@/pages/login-page/ui/LoginPage'
import { NotFoundPage } from '@/pages/not-found-page/ui/NotFoundPage'
import { RegisterPage } from '@/pages/register-page/ui/RegisterPage'

const AdminPage = lazy(() =>
  import('@/pages/admin-page/ui/AdminPage').then((module) => ({ default: module.AdminPage })),
)
const CasinoPage = lazy(() =>
  import('@/pages/casino-page/ui/CasinoPage').then((module) => ({ default: module.CasinoPage })),
)
const CasinoMiniAppPage = lazy(() =>
  import('@/pages/casino-miniapp-page/ui/CasinoMiniAppPage').then((module) => ({ default: module.CasinoMiniAppPage })),
)
const OrchestrationPage = lazy(() =>
  import('@/pages/orchestration-page/ui/OrchestrationPage').then((module) => ({ default: module.OrchestrationPage })),
)

const suspenseWrap = (children: ReactNode) => (
  <Suspense fallback={<div className="p-6 text-sm text-slate-500">Загрузка интерфейса...</div>}>{children}</Suspense>
)

const casinoBoundaryWrap = (children: ReactNode) => (
  <ErrorBoundary
    fallback={
      <div className="m-4 rounded-2xl border border-rose-200 bg-rose-50 p-6 text-rose-800">
        <h2 className="m-0 text-lg font-semibold">Ошибка модуля казино</h2>
        <p className="mb-0 mt-2 text-sm">Обновите страницу или вернитесь в ленту. Основной интерфейс продолжит работу.</p>
      </div>
    }
  >
    {suspenseWrap(children)}
  </ErrorBoundary>
)

const PublicRoute = ({ children }: { children: ReactNode }) => {
  const isAuthenticated = useAppSelector((state) => state.session.isAuthenticated)

  if (isAuthenticated) {
    return <Navigate to="/" replace />
  }
  return <>{children}</>
}

const AdminRoute = ({ children }: { children: ReactNode }) => {
  const { isAuthenticated, user } = useAppSelector((state) => state.session)
  if (!isAuthenticated) return <Navigate to="/login" replace />
  if (user?.role !== 'admin') return <Navigate to="/" replace />
  return <>{children}</>
}

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
      path: '/casino',
      element: casinoBoundaryWrap(<CasinoPage />),
    },
    {
      path: '/tma/casino',
      element: casinoBoundaryWrap(<CasinoMiniAppPage />),
    },
    {
      path: '/casino/miniapp',
      element: <Navigate to="/tma/casino" replace />,
    },
    {
      path: '/orchestration',
      element: suspenseWrap(<OrchestrationPage />),
    },
    {
      path: '/admin',
      element: (
        <AdminRoute>
          {suspenseWrap(<AdminPage />)}
        </AdminRoute>
      ),
    },
    {
      path: '*',
      element: <NotFoundPage />,
    },
  ])

  return <RRProvider router={router} />
}
