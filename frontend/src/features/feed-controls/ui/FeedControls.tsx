import type { FeedSort, FeedSource } from '@/entities/post/model/types'
import { setLimit, setSort, setSource } from '@/features/feed-controls/model/feedFiltersSlice'
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
  const { source, sort, limit } = useAppSelector((state) => state.feedFilters)

  return (
    <section className="feed-controls mudro-fade-up" aria-label="Контролы ленты">
      <div className="feed-toolbar">
        <div className="feed-toolbar__intro">
          <span className="feed-toolbar__eyebrow">Лента</span>
          <strong className="feed-toolbar__title">Один экран для чтения, media и обсуждений</strong>
          <p className="feed-toolbar__lead">
            Один рабочий surface архива: переключай источники, проверяй срез и раскрывай треды без второго технического экрана.
          </p>
        </div>

        <div className="feed-toolbar__stats" aria-label="Сводка ленты">
          <div className="feed-toolbar__stat">
            <span>Архив</span>
            <strong>{totalPosts}</strong>
          </div>
          <div className="feed-toolbar__stat">
            <span>VK snapshot</span>
            <strong>{vkPosts}</strong>
          </div>
          <div className="feed-toolbar__stat">
            <span>Telegram</span>
            <strong>{tgPosts}</strong>
          </div>
          <div className="feed-toolbar__stat">
            <span>Обновлено</span>
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
            }}
          >
            Сбросить фильтры
          </button>
        </div>
      </div>

      <div className="feed-controls__layout">
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
      </div>
    </section>
  )
}
