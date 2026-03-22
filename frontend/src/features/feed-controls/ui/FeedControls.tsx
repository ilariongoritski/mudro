import type { FeedSort, FeedSource } from '@/entities/post/model/types'
import { setLimit, setSort, setSource, setQuery } from '../model/feedFiltersSlice'
import { formatDateTime } from '@/shared/lib/format/date'
import { useAppDispatch, useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { cn } from '@/shared/lib/utils'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { Card, CardContent } from '@/shared/ui/card'

const sourceOptions: Array<{ value: FeedSource; label: string; badge: string; variant?: 'vk' | 'tg' }> = [
  { value: 'all', label: 'Общая', badge: 'ALL' },
  { value: 'vk', label: 'VK', badge: 'VK', variant: 'vk' },
  { value: 'tg', label: 'Telegram', badge: 'TG', variant: 'tg' },
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
                <span className="feed-user-badge__email">{user.email}</span>
                {user.isPremium && (
                  <span className="feed-user-badge__premium">PREMIUM</span>
                )}
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

        <div className="flex flex-wrap gap-3">
          <div className="flex flex-col items-center px-3 py-1.5 rounded-lg bg-slate-50">
            <span className="text-xs text-slate-500">Архив</span>
            <strong className="text-sm font-semibold">{totalPosts}</strong>
          </div>
          <div className="flex flex-col items-center px-3 py-1.5 rounded-lg bg-slate-50">
            <span className="text-xs text-slate-500">VK</span>
            <strong className="text-sm font-semibold text-vk">{vkPosts}</strong>
          </div>
          <div className="flex flex-col items-center px-3 py-1.5 rounded-lg bg-slate-50">
            <span className="text-xs text-slate-500">Telegram</span>
            <strong className="text-sm font-semibold text-tg">{tgPosts}</strong>
          </div>
          <div className="flex flex-col items-center px-3 py-1.5 rounded-lg bg-slate-50">
            <span className="text-xs text-slate-500">Обновлено</span>
            <strong className="text-sm font-semibold">{lastSyncAt ? formatDateTime(lastSyncAt) : '—'}</strong>
          </div>
        </div>

        <div className="flex flex-wrap items-end gap-4">
          <div className="space-y-1.5">
            <span className="text-xs font-medium text-slate-500">Источник</span>
            <div className="flex gap-1.5">
              {sourceOptions.map((option) => (
                <Button
                  key={option.value}
                  variant={source === option.value ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => dispatch(setSource(option.value))}
                  className={cn(source === option.value && option.variant === 'vk' && 'bg-vk hover:bg-vk/90', source === option.value && option.variant === 'tg' && 'bg-tg hover:bg-tg/90')}
                >
                  <Badge variant={option.variant ?? 'default'} className="text-[10px] px-1.5 py-0">
                    {option.badge}
                  </Badge>
                  {option.label}
                </Button>
              ))}
            </div>
          </div>

          <div className="space-y-1.5">
            <span className="text-xs font-medium text-slate-500">Сортировка</span>
            <div className="flex gap-1.5">
              {sortOptions.map((option) => (
                <Button
                  key={option.value}
                  variant={sort === option.value ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => dispatch(setSort(option.value))}
                >
                  {option.label}
                </Button>
              ))}
            </div>
          </div>

          <div className="space-y-1.5">
            <span className="text-xs font-medium text-slate-500">На экран</span>
            <select
              value={limit}
              onChange={(event) => dispatch(setLimit(Number(event.target.value)))}
              className="flex h-8 rounded-lg border border-slate-200 bg-white px-2.5 text-xs font-medium focus:outline-none focus:ring-2 focus:ring-sky-500"
            >
              {limits.map((value) => (
                <option key={value} value={value}>{value}</option>
              ))}
            </select>
          </div>

          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              dispatch(setSource('all'))
              dispatch(setSort('desc'))
              dispatch(setLimit(12))
            }}
          >
            Сбросить
          </Button>
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
