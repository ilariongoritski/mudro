import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'

import './layout.css'

export const AppLayout = () => {
  return (
    <div className="mudro-app-shell">
      <Sidebar />
      <div className="mudro-app-shell__workspace">
        <TopBar />
        <main className="mudro-app-shell__main">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
