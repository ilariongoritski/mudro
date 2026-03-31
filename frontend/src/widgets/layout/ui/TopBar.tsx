import { Link, useLocation } from 'react-router-dom'

const pageMeta: Record<string, { title: string; subtitle: string }> = {
  '/': {
    title: 'Лента',
    subtitle: 'Единый архив, фильтры по источникам и быстрый путь в control plane.',
  },
  '/movies': {
    title: 'Фильмы',
    subtitle: 'Дополнительный экран для legacy media browsing.',
  },
  '/chat': {
    title: 'Чат',
    subtitle: 'Legacy workspace для общения и входа в bot flow.',
  },
  '/profile': {
    title: 'Профиль',
    subtitle: 'Экран аккаунта и состояния авторизации.',
  },
  '/auth': {
    title: 'Auth',
    subtitle: 'Вход и bootstrap сессии.',
  },
  '/casino': {
    title: 'Casino',
    subtitle: 'Изолированная игровая поверхность со своим runtime.',
  },
  '/tma/casino': {
    title: 'Casino Mini App',
    subtitle: 'Компактный Telegram-native casino view.',
  },
  '/orchestration': {
    title: 'Control plane',
    subtitle: 'Локальный bridge, branch state и runtime signals.',
  },
}

export const TopBar = () => {
  const { pathname } = useLocation()
  const meta = pageMeta[pathname] ?? { title: 'Mudro', subtitle: 'Local-first workspace and control surface.' }

  return (
    <header className="mudro-topbar">
      <div className="mudro-topbar__copy">
        <span className="mudro-topbar__eyebrow">MUDRO workspace</span>
        <h1 className="mudro-topbar__title">{meta.title}</h1>
        <p className="mudro-topbar__subtitle">{meta.subtitle}</p>
      </div>

      <div className="mudro-topbar__actions">
        <span className="mudro-topbar__chip">Opus local</span>
        <span className="mudro-topbar__chip">Magic MCP</span>
        <Link to="/orchestration#bridge" className="mudro-topbar__chip mudro-topbar__chip_link">
          Карта bridge
        </Link>
      </div>
    </header>
  )
}
