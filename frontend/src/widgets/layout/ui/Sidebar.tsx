import { Film, Home, LogIn, LogOut, MessageCircle, Sparkles, TabletSmartphone, User, Workflow } from 'lucide-react'
import { Link, NavLink } from 'react-router'
import { useAppDispatch, useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { logout } from '@/features/auth/model/authSlice'
import { cn } from '@/shared/lib/utils'

const primaryNavItems = [
  { to: '/', icon: Home, label: 'Лента', description: 'Единый архив' },
  { to: '/orchestration', icon: Workflow, label: 'Control plane', description: 'Opus + Magic' },
  { to: '/casino', icon: Sparkles, label: 'Casino', description: 'Изолированный runtime' },
  { to: '/tma/casino', icon: TabletSmartphone, label: 'Mini app', description: 'Telegram surface' },
]

const secondaryNavItems = [
  { to: '/movies', icon: Film, label: 'Фильмы' },
  { to: '/chat', icon: MessageCircle, label: 'Чат' },
  { to: '/profile', icon: User, label: 'Профиль' },
]

export const Sidebar = () => {
  const dispatch = useAppDispatch()
  const token = useAppSelector((state) => state.auth.token)
  const user = useAppSelector((state) => state.auth.user)

  return (
    <aside className="mudro-sidebar hidden md:flex">
      <div className="mudro-sidebar__brand">
        <span className="mudro-sidebar__brand-mark">M</span>
        <span className="mudro-sidebar__brand-copy">
          <strong>Mudro</strong>
          <small>Локальный control plane</small>
        </span>
      </div>

      <div className="mudro-sidebar__section">
        <span className="mudro-sidebar__section-label">Рабочая зона</span>
        <nav className="mudro-sidebar__nav">
          {primaryNavItems.map(({ to, icon: Icon, label, description }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) => cn('mudro-sidebar__link', isActive && 'mudro-sidebar__link_active')}
            >
              <span className="mudro-sidebar__link-icon">
                <Icon size={18} />
              </span>
              <span className="mudro-sidebar__link-copy">
                <strong>{label}</strong>
                <span>{description}</span>
              </span>
            </NavLink>
          ))}
        </nav>
      </div>

      <div className="mudro-sidebar__section">
        <span className="mudro-sidebar__section-label">Дополнительно</span>
        <nav className="mudro-sidebar__nav">
          {secondaryNavItems.map(({ to, icon: Icon, label }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) => cn('mudro-sidebar__link', isActive && 'mudro-sidebar__link_active')}
            >
              <span className="mudro-sidebar__link-icon">
                <Icon size={18} />
              </span>
              <span className="mudro-sidebar__link-copy">
                <strong>{label}</strong>
                <span>Вторичный экран</span>
              </span>
            </NavLink>
          ))}
        </nav>
      </div>

      <div className="mudro-sidebar__bridge">
        <span className="mudro-sidebar__section-label">Связка</span>
        <strong>Claude Opus ↔ Magic MCP</strong>
        <p>
          Локальный Opus ключ отвечает за reasoning и изменения. Magic MCP помогает с контекстом и визуальными
          паттернами, но браузер видит только status surface.
        </p>
        <div className="mudro-sidebar__bridge-flow" aria-label="Схема связки">
          <span className="mudro-sidebar__bridge-chip mudro-sidebar__bridge-chip_accent">Opus</span>
          <span className="mudro-sidebar__bridge-chip">MCP</span>
          <span className="mudro-sidebar__bridge-chip">MUDRO</span>
        </div>
        <Link to="/orchestration#bridge" className="mudro-sidebar__bridge-link">
          Карта bridge
        </Link>
      </div>

      <div className="mudro-sidebar__footer">
        {token ? (
          <div className="mudro-sidebar__user">
            <span className="mudro-sidebar__user-copy">
              <strong>{user?.username ?? 'Пользователь'}</strong>
              <small>{user?.email ?? 'В системе'}</small>
            </span>
            <button
              onClick={() => dispatch(logout())}
              className="mudro-sidebar__logout"
              title="Выйти"
            >
              <LogOut size={16} />
            </button>
          </div>
        ) : (
          <NavLink
            to="/auth"
            className={({ isActive }) => cn('mudro-sidebar__link', isActive && 'mudro-sidebar__link_active')}
          >
            <span className="mudro-sidebar__link-icon">
              <LogIn size={18} />
            </span>
            <span className="mudro-sidebar__link-copy">
              <strong>Войти</strong>
              <span>Открыть аккаунт</span>
            </span>
          </NavLink>
        )}
        <span className="mudro-sidebar__live">
          <span className="mudro-sidebar__live-dot" />
          MVP live
        </span>
      </div>
    </aside>
  )
}
