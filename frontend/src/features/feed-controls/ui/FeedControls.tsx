import type { FeedSort, FeedSource } from '@/entities/post/model/types'
import { setLimit, setSort, setSource } from '@/features/feed-controls/model/feedFiltersSlice'
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

export const FeedControls = () => {
  const dispatch = useAppDispatch()
  const { source, sort, limit } = useAppSelector((state) => state.feedFilters)

  return (
    <section className="feed-controls mudro-fade-up" aria-label="Контролы ленты">
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
