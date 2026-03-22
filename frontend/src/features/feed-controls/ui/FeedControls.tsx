import { setLimit, setQuery, setSort, setSource } from '../model/feedFiltersSlice'
import type { FeedSort, FeedSource } from '@/entities/post/model/types'
import { formatDateTime } from '@/shared/lib/format/date'
import { useAppDispatch, useAppSelector } from '@/shared/lib/hooks/storeHooks'

import './FeedControls.css'

const sourceOptions: Array<{ value: FeedSource; label: string; badge: string }> = [
  { value: 'all', label: 'Общая', badge: 'ALL' },
  { value: 'vk', label: 'VK', badge: 'VK' },
  { value: 'tg', label: 'Telegram', badge: 'TG' },
]

const sortOptions: Array<{ value: FeedSort; label: string }> = [
  { value: 'desc', label: 'Новые сверху' },
  { value: 'asc', label: 'Старые сверху' },
]

const limits = [12, 24, 48]

interface FeedControlsProps {
  totalPosts?: number
  vkPosts?: number
  tgPosts?: number
  lastSyncAt?: string
}

const sourceStateLabel: Record<FeedSource, string> = {
  all: 'Все источники',
  vk: 'Только VK',
  tg: 'Только Telegram',
}

const sortStateLabel: Record<FeedSort, string> = {
  desc: 'Сначала новые',
  asc: 'Сначала старые',
}

export const FeedControls = ({ totalPosts = 0, vkPosts = 0, tgPosts = 0, lastSyncAt }: FeedControlsProps) => {
  const dispatch = useAppDispatch()
  const { source, sort, limit, query } = useAppSelector((state) => state.feedFilters)
  const user = useAppSelector((state) => state.session.user)

  return (
    <section className="feed-controls mudro-fade-up" aria-label="Контролы ленты">
      <div className="feed-toolbar">
        <div className="feed-toolbar__intro">
          <div className="feed-toolbar__user-status">
            {user ? (
              <div className="feed-user-badge">
                <span className="feed-user-badge__email">{user.email ?? user.username}</span>
                {user.isPremium && <span className="feed-user-badge__premium">PREMIUM</span>}
                <span className="feed-user-badge__role">{user.role?.toUpperCase()}</span>
              </div>
            ) : (
              <span className="feed-toolbar__eyebrow">Архив</span>
            )}
          </div>
          <strong className="feed-toolbar__title">Живая лента с реальными постами, вложениями и обсуждениями</strong>
          <p className="feed-toolbar__lead">
            Переключай источники, меняй порядок и быстро открывай посты, комментарии и media без ухода со страницы.
          </p>
        </div>

        <div className="feed-toolbar__stats" aria-label="Сводка ленты">
          <div className="feed-toolbar__stat">
            <span>Всего постов</span>
            <strong>{totalPosts}</strong>
          </div>
          <div className="feed-toolbar__stat">
            <span>VK</span>
            <strong>{vkPosts}</strong>
          </div>
          <div className="feed-toolbar__stat">
            <span>Telegram</span>
            <strong>{tgPosts}</strong>
          </div>
          <div className="feed-toolbar__stat">
            <span>Последний sync</span>
            <strong>{lastSyncAt ? formatDateTime(lastSyncAt) : '—'}</strong>
          </div>
        </div>

        <div className="feed-toolbar__status">
          <span className="feed-toolbar__signal">Источник: {sourceStateLabel[source]}</span>
          <span className="feed-toolbar__signal">Порядок: {sortStateLabel[sort]}</span>
          <span className="feed-toolbar__signal">На экран: {limit}</span>
          <button
            type="button"
            className="feed-toolbar__reset"
            onClick={() => {
              dispatch(setSource('all'))
              dispatch(setSort('desc'))
              dispatch(setLimit(12))
              dispatch(setQuery(''))
            }}
          >
            Сбросить фильтры
          </button>
        </div>
      </div>

      <div className="feed-controls__group">
        <span className="feed-controls__label">Поиск по тексту</span>
        <div className="feed-controls__search-box">
          <input
            type="text"
            placeholder="Что ищем?.."
            value={query ?? ''}
            onChange={(e) => dispatch(setQuery(e.target.value))}
            className="feed-controls__search-input"
          />
        </div>
      </div>

      <div className="feed-controls__group">
        <span className="feed-controls__label">Источник</span>
        <div className="feed-controls__row">
          {sourceOptions.map((option) => (
            <button
              key={option.value}
              type="button"
              onClick={() => dispatch(setSource(option.value))}
              className={`feed-pill ${source === option.value ? 'feed-pill_active' : ''}`}
            >
              <span className="feed-pill__badge">{option.badge}</span>
              {option.label}
            </button>
          ))}
        </div>
      </div>

      <div className="feed-controls__group">
        <span className="feed-controls__label">Сортировка</span>
        <div className="feed-controls__row">
          {sortOptions.map((option) => (
            <button
              key={option.value}
              type="button"
              onClick={() => dispatch(setSort(option.value))}
              className={`feed-pill ${sort === option.value ? 'feed-pill_active' : ''}`}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      <label className="feed-controls__limit">
        <span className="feed-controls__label">Постов на экран</span>
        <select value={limit} onChange={(event) => dispatch(setLimit(Number(event.target.value)))}>
          {limits.map((value) => (
            <option key={value} value={value}>
              {value}
            </option>
          ))}
        </select>
      </label>
    </section>
  )
}
