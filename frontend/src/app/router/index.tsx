import { lazy, Suspense, type ReactNode } from 'react'
import { createBrowserRouter } from 'react-router'
import { AppLayout } from '@/widgets/layout/ui/AppLayout'
import { FeedPage } from '@/pages/feed-page/ui/FeedPage'

const MoviesPage = lazy(() => import('@/pages/movies-page/ui/MoviesPage'))
const ChatPage = lazy(() => import('@/pages/chat-page/ui/ChatPage'))
const ProfilePage = lazy(() => import('@/pages/profile-page/ui/ProfilePage'))
const AuthPage = lazy(() => import('@/pages/auth-page/ui/AuthPage'))

const withSuspense = (children: ReactNode) => (
  <Suspense fallback={<div className="flex items-center justify-center h-64 text-sm text-slate-400">Загрузка...</div>}>
    {children}
  </Suspense>
)

export const router = createBrowserRouter([
  {
    element: <AppLayout />,
    children: [
      { path: '/', element: <FeedPage /> },
      { path: '/movies', element: withSuspense(<MoviesPage />) },
      { path: '/chat', element: withSuspense(<ChatPage />) },
      { path: '/profile', element: withSuspense(<ProfilePage />) },
      { path: '/auth', element: withSuspense(<AuthPage />) },
    ],
  },
])
