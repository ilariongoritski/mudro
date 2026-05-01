import { useCallback, useEffect, useState } from 'react'
import { setLimit, setQuery, setSort, setSource } from '../model/feedFiltersSlice'
import type { FeedSort, FeedSource } from '@/entities/post/model/types'
import { useAppDispatch, useAppSelector } from '@/shared/lib/hooks/storeHooks'

import './FeedControls.css'

const sourceOptions: Array<{ value: FeedSource; label: string }> = [
  { value: 'all', label: 'Все' },
  { value: 'vk',  label: 'VK' },
  { value: 'tg',  label: 'TG' },
]

const sortOptions: Array<{ value: FeedSort; label: string }> = [
  { value: 'desc', label: '↓ Новые' },
  { value: 'asc',  label: '↑ Старые' },
]

const limitOptions = [12, 24, 48]

interface FeedControlsProps {
  totalPosts?: number
  vkPosts?: number
  tgPosts?: number
}

export const FeedControls = ({ totalPosts = 0 }: FeedControlsProps) => {
  const dispatch = useAppDispatch()
  const { source, sort, limit, query } = useAppSelector((state) => state.feedFilters)

  // Локальный стейт для debounce поиска
  const [searchInput, setSearchInput] = useState(query ?? '')

  // Debounce 400ms
  useEffect(() => {
    const timer = setTimeout(() => {
      dispatch(setQuery(searchInput.trim() || ''))
    }, 400)
    return () => clearTimeout(timer)
  }, [searchInput, dispatch])

  const handleReset = useCallback(() => {
    dispatch(setSource('all'))
    dispatch(setSort('desc'))
    dispatch(setLimit(12))
    dispatch(setQuery(''))
    setSearchInput('')
  }, [dispatch])

  const isFiltered =
    source !== 'all' || sort !== 'desc' || limit !== 12 || (query?.trim().length ?? 0) > 0

  return (
    <div className="feed-controls">
      <div className="feed-controls__bar">
        {/* Источник */}
        <div className="feed-controls__group" role="group" aria-label="Источник">
          {sourceOptions.map((opt) => (
            <button
              key={opt.value}
              type="button"
              className={`feed-pill ${source === opt.value ? 'feed-pill_active' : ''}`}
              aria-pressed={source === opt.value}
              onClick={() => dispatch(setSource(opt.value))}
            >
              {opt.label}
            </button>
          ))}
        </div>

        <div className="feed-controls__sep" aria-hidden="true" />

        {/* Сортировка */}
        <div className="feed-controls__group" role="group" aria-label="Сортировка">
          {sortOptions.map((opt) => (
            <button
              key={opt.value}
              type="button"
              className={`feed-pill ${sort === opt.value ? 'feed-pill_active' : ''}`}
              aria-pressed={sort === opt.value}
              onClick={() => dispatch(setSort(opt.value))}
            >
              {opt.label}
            </button>
          ))}
        </div>

        <div className="feed-controls__sep" aria-hidden="true" />

        {/* Лимит */}
        <div className="feed-controls__group" role="group" aria-label="Постов на экран">
          {limitOptions.map((val) => (
            <button
              key={val}
              type="button"
              className={`feed-pill ${limit === val ? 'feed-pill_active' : ''}`}
              aria-pressed={limit === val}
              onClick={() => dispatch(setLimit(val))}
            >
              {val}
            </button>
          ))}
        </div>

        {/* Поиск */}
        <input
          type="search"
          className="feed-controls__search"
          placeholder="Поиск…"
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          aria-label="Поиск по тексту"
        />

        {/* Счётчик */}
        {totalPosts > 0 && (
          <span className="feed-controls__count">{totalPosts}</span>
        )}

        {/* Сброс — только если есть активные фильтры */}
        {isFiltered && (
          <button
            type="button"
            className="feed-pill feed-pill_reset"
            onClick={handleReset}
            title="Сбросить фильтры"
            aria-label="Сбросить все фильтры"
          >
            ✕
          </button>
        )}
      </div>
    </div>
  )
}
