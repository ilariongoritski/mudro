import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'

import './layout.css'

export const AppLayout = () => {
  return (
    <div className="mudro-app-shell">
      <a href="#mudro-main-content" className="skip-to-main">Перейти к основному содержимому</a>
      <Sidebar />
      <div className="mudro-app-shell__workspace">
        <TopBar />
        <main id="mudro-main-content" className="mudro-app-shell__main">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
