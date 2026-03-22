import { RouterProvider } from 'react-router'
import { router } from '@/app/router'

export const App = () => {
  return <RouterProvider router={router} />
}
