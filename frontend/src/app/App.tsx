import { AppRouterProvider } from '@/app/providers/RouterProvider'
import { ErrorBoundary } from '@/shared/ui/ErrorBoundary'
import { AuthProvider } from '@/app/providers/AuthProvider'
import { E2EEProvider } from '@/app/providers/E2EEProvider'
import '@/shared/ui/ErrorBoundary.css'

export const App = () => {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <E2EEProvider>
          <AppRouterProvider />
        </E2EEProvider>
      </AuthProvider>
    </ErrorBoundary>
  )
}
