import { lazy, Suspense } from 'react'
import { createBrowserRouter } from 'react-router'
import { AppLayout } from '@/widgets/layout/ui/AppLayout'
import { FeedPage } from '@/pages/feed-page/ui/FeedPage'

const MoviesPage = lazy(() => import('@/pages/movies-page/ui/MoviesPage'))
const ChatPage = lazy(() => import('@/pages/chat-page/ui/ChatPage'))
const ProfilePage = lazy(() => import('@/pages/profile-page/ui/ProfilePage'))

const SuspenseWrap = ({ children }: { children: React.ReactNode }) => (
  <Suspense fallback={<div className="flex items-center justify-center h-64 text-sm text-slate-400">Загрузка...</div>}>
    {children}
  </Suspense>
)

export const router = createBrowserRouter([
  {
    element: <AppLayout />,
    children: [
      { path: '/', element: <FeedPage /> },
      { path: '/movies', element: <SuspenseWrap><MoviesPage /></SuspenseWrap> },
      { path: '/chat', element: <SuspenseWrap><ChatPage /></SuspenseWrap> },
      { path: '/profile', element: <SuspenseWrap><ProfilePage /></SuspenseWrap> },
    ],
  },
])
