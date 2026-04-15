import { Home, LogIn, LogOut, MessageCircle, Settings2, Shield, Sparkles, User, UserPlus } from 'lucide-react'
import { Link, NavLink } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { supabase } from '@/shared/api/supabase'
import { cn } from '@/shared/lib/utils'
import { MudroLogoMark } from '@/shared/ui/MudroLogoMark'

const navItems = [
  { to: '/', icon: Home, label: 'Лента', description: 'Посты и обновления' },
  { to: '/chat', icon: MessageCircle, label: 'Мессенджер', description: 'Общий чат' },
  { to: '/casino', icon: Sparkles, label: 'Казино', description: 'Игровой зал' },
  { to: '/orchestration', icon: Settings2, label: 'Контур', description: 'Статус и оркестрация' },
]

export const Sidebar = () => {
  const dispatch = useAppDispatch()
  const token = useAppSelector((state) => state.session.token)
  const user = useAppSelector((state) => state.session.user)

  return (
    <aside className="mudro-sidebar">
      <div className="mudro-sidebar__brand">
        <span className="mudro-sidebar__brand-mark"><MudroLogoMark /></span>
        <span className="mudro-sidebar__brand-copy">
          <strong>Mudro</strong>
          <small>Социальная сеть</small>
        </span>
      </div>

      <div className="mudro-sidebar__section">
        <span className="mudro-sidebar__section-label">Навигация</span>
        <nav className="mudro-sidebar__nav">
          {navItems.map(({ to, icon: Icon, label, description }) => (
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

          {user?.role === 'admin' && (
            <NavLink
              to="/admin"
              className={({ isActive }) => cn('mudro-sidebar__link mudro-sidebar__link_admin', isActive && 'mudro-sidebar__link_active')}
            >
              <span className="mudro-sidebar__link-icon">
                <Shield size={18} />
              </span>
              <span className="mudro-sidebar__link-copy">
                <strong>Админ</strong>
                <span>Панель управления</span>
              </span>
            </NavLink>
          )}
        </nav>
      </div>

      <div className="mudro-sidebar__footer">
        {token ? (
          <div className="mudro-sidebar__user">
            <NavLink
              to="/profile"
              className="mudro-sidebar__user-link"
            >
              <span className="mudro-sidebar__user-avatar">
                <User size={16} />
              </span>
              <span className="mudro-sidebar__user-copy">
                <strong>{user?.username ?? 'Пользователь'}</strong>
                <small>{user?.email ?? 'В системе'}</small>
              </span>
            </NavLink>
            <button
              type="button"
              className="mudro-sidebar__menu-item mudro-sidebar__logout"
              onClick={async () => {
                try {
                  await supabase.auth.signOut()
                  // AuthProvider should handle dispatch(logout()) via onAuthStateChange
                } catch (err) {
                  console.error('Logout failed:', err)
                }
              }}
            >
              <LogOut size={16} aria-hidden="true" />
            </button>
          </div>
        ) : (
          <div className="mudro-sidebar__auth-cta">
            <Link to="/login" className="mudro-sidebar__auth-btn mudro-sidebar__auth-btn_primary">
              <LogIn size={15} />
              <span>Войти</span>
            </Link>
            <Link to="/register" className="mudro-sidebar__auth-btn mudro-sidebar__auth-btn_secondary">
              <UserPlus size={15} />
              <span>Регистрация</span>
            </Link>
          </div>
        )}
        <span className="mudro-sidebar__live">
          <span className="mudro-sidebar__live-dot" />
          MVP live
        </span>
      </div>
    </aside>
  )
}
