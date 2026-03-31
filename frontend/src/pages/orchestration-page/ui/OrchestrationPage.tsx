import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'

import {
  normalizeOrchestrationStatus,
  useGetOrchestrationStatusQuery,
} from '@/features/orchestration/api/orchestrationApi'

import './OrchestrationPage.css'

const SKARO_DASHBOARD_URL = 'http://127.0.0.1:4700/dashboard'

const formatMoscowDateTime = (value: string) => {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    return 'Сейчас'
  }

  return new Intl.DateTimeFormat('ru-RU', {
    timeZone: 'Europe/Moscow',
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(parsed)
}

const formatUpdatedAt = (value: string) => {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    return 'неизвестно'
  }

  return new Intl.DateTimeFormat('ru-RU', {
    timeZone: 'Europe/Moscow',
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  }).format(parsed)
}

const fallbackStatus = (now: Date) =>
  normalizeOrchestrationStatus({
    branch: 'codex/casino-mvp',
    commit: 'unknown',
    updated_at: now.toISOString(),
    moscow_time: new Intl.DateTimeFormat('ru-RU', {
      timeZone: 'Europe/Moscow',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    }).format(now),
    dashboard_url: SKARO_DASHBOARD_URL,
    api_endpoint: '/api/orchestration/status',
    state: [
      'Локальный ключ Opus отвечает за планирование, ревью и большие изменения.',
      'Codex применяет изменения в репозиторий и держит main каноничным.',
      'Skaro используется как cockpit для статуса, валидации и контекстного контроля.',
    ],
    todo: [
      'Подключить боевой backend endpoint orchestration status.',
      'Показать usage по токенам и историю запусков локального runtime.',
      'Добавить локальный launcher профиля Skaro Claude под E:\\mudr\\_mudro-local.',
    ],
    done: [
      'Маршрут orchestration уже есть на фронтенде.',
      'Навигация из ленты ведёт в control plane.',
      'Правила локальной orchestration-схемы уже задокументированы.',
    ],
    status: [
      { label: 'API', value: 'fallback', tone: 'warn' },
      { label: 'Workspace', value: 'local', tone: 'ok' },
      { label: 'Runner', value: 'Codex + Skaro', tone: 'accent' },
      { label: 'Locale', value: 'Europe/Moscow', tone: 'neutral' },
    ],
  })

const quickActions = [
  { label: 'Открыть ленту', to: '/' },
  { label: 'Открыть casino mini app', to: '/tma/casino' },
  { label: 'Открыть Skaro dashboard', href: SKARO_DASHBOARD_URL, external: true },
] as const satisfies ReadonlyArray<
  | { label: string; to: string; external?: false }
  | { label: string; href: string; external: true }
>

const bridgeCards = [
  {
    title: 'Локальный Opus',
    text: 'Это основной reasoning-layer: план, архитектура, code review и тяжёлые итерации идут через локальный ключ и не смешиваются с UI-runtime.',
  },
  {
    title: 'Magic MCP',
    text: 'Используется как визуальный и контекстный помощник: паттерны интерфейсов, layout inspiration и точечная проверка shell. Он не заменяет Opus и не пишет в runtime.',
  },
  {
    title: 'MUDRO runtime',
    text: 'Локальные сервисы работают отдельно: feed-api, auth-api, orchestration-api и casino. Фронтенд только показывает состояние и вызывает нужные API.',
  },
]

const bridgeFlow = ['Запрос', 'Opus план', 'Codex apply', 'Magic UI check', 'main']

export const OrchestrationPage = () => {
  const [moscowNow, setMoscowNow] = useState(() => new Date())
  const [copyState, setCopyState] = useState<'idle' | 'copied' | 'error'>('idle')

  const { data, isFetching, isError, refetch } = useGetOrchestrationStatusQuery(undefined, {
    pollingInterval: 30000,
    refetchOnFocus: true,
  })

  useEffect(() => {
    const timer = window.setInterval(() => {
      setMoscowNow(new Date())
    }, 1000)

    return () => window.clearInterval(timer)
  }, [])

  const status = useMemo(() => data ?? fallbackStatus(moscowNow), [data, moscowNow])
  const apiState = isError ? 'offline' : isFetching ? 'refreshing' : 'live'

  const handleCopyCommit = async () => {
    try {
      await navigator.clipboard.writeText(status.commit)
      setCopyState('copied')
      window.setTimeout(() => setCopyState('idle'), 1400)
    } catch {
      setCopyState('error')
      window.setTimeout(() => setCopyState('idle'), 1400)
    }
  }

  return (
    <motion.main
      className="orchestration-page"
      initial={{ opacity: 0, y: 18 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35, ease: 'easeOut' }}
    >
      <header className="orchestration-page__topbar">
        <div className="orchestration-page__brand">
          <span className="orchestration-page__brand-mark">M</span>
          <div className="orchestration-page__brand-copy">
            <strong>Orchestration</strong>
            <small>control plane для MUDRO</small>
          </div>
        </div>

        <div className="orchestration-page__top-actions">
          <button type="button" className="orchestration-page__ghost-action" onClick={() => refetch()}>
            Обновить
          </button>
          <Link to="/" className="orchestration-page__ghost-action">
            Лента
          </Link>
          <Link to="/tma/casino" className="orchestration-page__ghost-action">
            Casino
          </Link>
        </div>
      </header>

      <section className="orchestration-page__hero">
        <div className="orchestration-page__hero-copy">
          <span className={`orchestration-page__status-chip orchestration-page__status-chip_${apiState}`}>
            {apiState === 'live' ? 'API live' : apiState === 'refreshing' ? 'API refresh' : 'API offline'}
          </span>
          <h1>Центр управления Opus, Magic MCP и локальным runtime MUDRO.</h1>
          <p>
            Этот экран собирает текущую ветку, commit, время Москвы, ссылку на Skaro и рабочие списки для планирования
            и исполнения. Если backend ещё не отдаёт статус, страница показывает локальный fallback без пустых блоков.
          </p>

          <div className="orchestration-page__actions">
            {quickActions.map((action) =>
              'href' in action ? (
                <a
                  key={action.label}
                  href={action.href}
                  target="_blank"
                  rel="noreferrer"
                  className="orchestration-page__primary-action"
                >
                  {action.label}
                </a>
              ) : (
                <Link key={action.label} to={action.to} className="orchestration-page__secondary-action">
                  {action.label}
                </Link>
              ),
            )}
            <button type="button" className="orchestration-page__secondary-action" onClick={handleCopyCommit}>
              {copyState === 'copied' ? 'Commit copied' : copyState === 'error' ? 'Copy failed' : 'Copy commit'}
            </button>
          </div>
        </div>

        <aside className="orchestration-page__signal-rail">
          <div className="orchestration-page__signal-row">
            <span>Branch</span>
            <strong title={status.branch}>{status.branch}</strong>
          </div>
          <div className="orchestration-page__signal-row">
            <span>Commit</span>
            <strong title={status.commit}>{status.commit.slice(0, 8)}</strong>
          </div>
          <div className="orchestration-page__signal-row">
            <span>Москва</span>
            <strong>{status.moscow_time || formatMoscowDateTime(moscowNow.toISOString())}</strong>
          </div>
          <div className="orchestration-page__signal-row">
            <span>Updated</span>
            <strong>{formatUpdatedAt(status.updated_at)}</strong>
          </div>
          <div className="orchestration-page__signal-row">
            <span>Endpoint</span>
            <strong>{status.api_endpoint}</strong>
          </div>
        </aside>
      </section>

      {isError ? (
        <div className="orchestration-page__notice">
          Backend endpoint пока недоступен. Показываем локальный fallback, чтобы control surface оставался рабочим.
        </div>
      ) : null}

      <section className="orchestration-page__bridge" id="bridge" aria-label="Связка Opus, Magic MCP и runtime">
        <div className="orchestration-page__section-head">
          <span className="orchestration-page__panel-kicker">Bridge map</span>
          <h2>Как связаны Opus, Magic MCP и локальный runtime</h2>
          <p>
            Локальный Claude Opus ключ — это слой reasoning: он собирает план, делает ревью и ведёт большие изменения.
            Magic MCP подключается только как визуальный и контекстный MCP для UI-паттернов, layout inspiration и
            проверки shell. Runtime и бизнес-логика живут отдельно в Go-сервисах MUDRO.
          </p>
        </div>

        <div className="orchestration-page__bridge-grid">
          {bridgeCards.map((card) => (
            <article key={card.title} className="orchestration-page__bridge-card">
              <span className="orchestration-page__panel-kicker">{card.title}</span>
              <p>{card.text}</p>
            </article>
          ))}
        </div>

        <div className="orchestration-page__bridge-flow" aria-label="Схема работы">
          {bridgeFlow.map((step, index) => (
            <span key={step} className={`orchestration-page__bridge-chip ${index === 0 ? 'orchestration-page__bridge-chip_accent' : ''}`}>
              {step}
              {index < bridgeFlow.length - 1 ? <span className="orchestration-page__bridge-arrow">→</span> : null}
            </span>
          ))}
        </div>
      </section>

      <section className="orchestration-page__panels" aria-label="Orchestration workspace">
        <motion.article
          className="orchestration-page__panel"
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.05 }}
        >
          <div className="orchestration-page__panel-head">
            <span className="orchestration-page__panel-kicker">Состояние</span>
            <h2>Текущий контракт работы</h2>
          </div>
          <ul className="orchestration-page__list">
            {status.state.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </motion.article>

        <motion.article
          className="orchestration-page__panel"
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
        >
          <div className="orchestration-page__panel-head">
            <span className="orchestration-page__panel-kicker">Todo</span>
            <h2>Что ещё надо соединить</h2>
          </div>
          <ol className="orchestration-page__list orchestration-page__list_ordered">
            {status.todo.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ol>
        </motion.article>

        <motion.article
          className="orchestration-page__panel"
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.15 }}
        >
          <div className="orchestration-page__panel-head">
            <span className="orchestration-page__panel-kicker">Done</span>
            <h2>Что уже закреплено</h2>
          </div>
          <ul className="orchestration-page__list">
            {status.done.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </motion.article>

        <motion.article
          className="orchestration-page__panel"
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
        >
          <div className="orchestration-page__panel-head">
            <span className="orchestration-page__panel-kicker">Сигналы</span>
            <h2>Статус и быстрые ссылки</h2>
          </div>
          <dl className="orchestration-page__status-grid">
            {status.status.map((item) => (
              <div key={`${item.label}-${item.value}`} className="orchestration-page__status-row">
                <dt>{item.label}</dt>
                <dd className={`orchestration-page__status-value orchestration-page__status-value_${item.tone ?? 'neutral'}`}>
                  {item.value}
                </dd>
              </div>
            ))}
          </dl>

          <div className="orchestration-page__footer-row">
            <span>Skaro dashboard</span>
            <a href={status.dashboard_url} target="_blank" rel="noreferrer">
              {status.dashboard_url}
            </a>
          </div>
        </motion.article>
      </section>
    </motion.main>
  )
}
