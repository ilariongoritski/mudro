import { AppRouterProvider } from '@/app/providers/RouterProvider'
import { ErrorBoundary } from '@/shared/ui/ErrorBoundary'
import '@/shared/ui/ErrorBoundary.css'

export const App = () => {
  return (
    <ErrorBoundary>
      <AppRouterProvider />
    </ErrorBoundary>
  )
}
