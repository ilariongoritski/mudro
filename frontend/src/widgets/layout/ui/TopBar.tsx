import { useLocation } from 'react-router-dom'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'

const pageMeta: Record<string, { title: string; subtitle: string }> = {
  '/': {
    title: 'Лента',
    subtitle: 'Посты и обновления из всех источников',
  },
  '/chat': {
    title: 'Мессенджер',
    subtitle: 'Общий чат в реальном времени',
  },
  '/movies': {
    title: 'Кинотоп',
    subtitle: 'Фильмы и сериалы',
  },
  '/casino': {
    title: 'Казино',
    subtitle: 'Игровой зал',
  },
  '/profile': {
    title: 'Профиль',
    subtitle: 'Ваш аккаунт и настройки',
  },
  '/orchestration': {
    title: 'Control plane',
    subtitle: 'Управление и мониторинг',
  },
  '/admin': {
    title: 'Администрирование',
    subtitle: 'Управление пользователями и системой',
  },
}

export const TopBar = () => {
  const { pathname } = useLocation()
  const user = useAppSelector((state) => state.session.user)
  const meta = pageMeta[pathname] ?? { title: 'Mudro', subtitle: '' }

  return (
    <header className="mudro-topbar">
      <div className="mudro-topbar__copy">
        <span className="mudro-topbar__eyebrow">Mudro</span>
        <h1 className="mudro-topbar__title">{meta.title}</h1>
        {meta.subtitle && <p className="mudro-topbar__subtitle">{meta.subtitle}</p>}
      </div>

      {user && (
        <div className="mudro-topbar__actions">
          <span className="mudro-topbar__chip">{user.username}</span>
        </div>
      )}
    </header>
  )
}
