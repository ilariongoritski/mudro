import { AppRouterProvider } from '@/app/providers/RouterProvider'
import { ErrorBoundary } from '@/shared/ui/ErrorBoundary'
import { AuthProvider } from '@/app/providers/AuthProvider'
import '@/shared/ui/ErrorBoundary.css'

export const App = () => {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <AppRouterProvider />
      </AuthProvider>
    </ErrorBoundary>
  )
}
