import { Home, Film, MessageCircle, User } from 'lucide-react'
import { NavLink } from 'react-router'
import { cn } from '@/shared/lib/utils'

const navItems = [
  { to: '/', icon: Home, label: 'Лента' },
  { to: '/movies', icon: Film, label: 'Фильмы' },
  { to: '/chat', icon: MessageCircle, label: 'Чат' },
  { to: '/profile', icon: User, label: 'Профиль' },
]

export const Sidebar = () => {
  return (
    <aside className="hidden md:flex flex-col w-60 shrink-0 border-r border-slate-200 bg-white h-screen sticky top-0">
      <div className="flex items-center gap-2.5 px-5 py-5">
        <span className="flex items-center justify-center w-9 h-9 rounded-xl bg-mudro-pink text-white font-bold text-lg">
          M
        </span>
        <span className="flex flex-col leading-tight">
          <strong className="text-sm font-semibold text-mudro-text">Mudro</strong>
          <small className="text-xs text-mudro-muted">живой архив</small>
        </span>
      </div>

      <nav className="flex flex-col gap-1 px-3 mt-2 flex-1">
        {navItems.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-colors',
                isActive
                  ? 'bg-mudro-pink/10 text-mudro-pink'
                  : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900',
              )
            }
          >
            <Icon size={18} />
            {label}
          </NavLink>
        ))}
      </nav>

      <div className="px-5 py-4 border-t border-slate-100">
        <span className="inline-flex items-center gap-1.5 text-xs text-mudro-muted">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
          MVP · server live
        </span>
      </div>
    </aside>
  )
}
