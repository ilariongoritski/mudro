import { useEffect, useState } from 'react'

import { Button } from '@/shared/ui/Button'
import type { GenreOption, MovieQuery } from '@/entities/movie/model/types'

type Props = {
  genres: GenreOption[]
  value: MovieQuery
  onApply: (next: MovieQuery) => void
  onReset: () => void
}

type DraftState = {
  yearMin: string
  durationMin: string
  includeGenre: string
  excludeGenres: string
}

function toDraft(value: MovieQuery): DraftState {
  return {
    yearMin: value.yearMin ? String(value.yearMin) : '',
    durationMin: value.durationMin ? String(value.durationMin) : '',
    includeGenre: value.includeGenre ?? '',
    excludeGenres: value.excludeGenres?.join(', ') ?? '',
  }
}

function parseExcludeGenres(value: string): string[] {
  return value
    .split(',')
    .map((chunk) => chunk.trim().toLowerCase())
    .filter(Boolean)
}

export const MovieFilters = ({ genres, value, onApply, onReset }: Props) => {
  const [draft, setDraft] = useState<DraftState>(() => toDraft(value))

  useEffect(() => {
    setDraft(toDraft(value))
  }, [value])

  return (
    <form
      className="surface filters-panel"
      onSubmit={(event) => {
        event.preventDefault()
        onApply({
          yearMin: draft.yearMin ? Number(draft.yearMin) : undefined,
          durationMin: draft.durationMin ? Number(draft.durationMin) : undefined,
          includeGenre: draft.includeGenre || undefined,
          excludeGenres: parseExcludeGenres(draft.excludeGenres),
          page: 1,
          pageSize: value.pageSize,
        })
      }}
    >
      <div className="filters-grid">
        <label className="filters-field">
          <span>Минимальный год</span>
          <input
            value={draft.yearMin}
            onChange={(event) => setDraft((current) => ({ ...current, yearMin: event.target.value }))}
            placeholder="2010"
            inputMode="numeric"
          />
        </label>

        <label className="filters-field">
          <span>Минимальная длина</span>
          <input
            value={draft.durationMin}
            onChange={(event) => setDraft((current) => ({ ...current, durationMin: event.target.value }))}
            placeholder="90"
            inputMode="numeric"
          />
        </label>

        <label className="filters-field">
          <span>Жанр</span>
          <select
            value={draft.includeGenre}
            onChange={(event) => setDraft((current) => ({ ...current, includeGenre: event.target.value }))}
          >
            <option value="">Любой</option>
            {genres.map((genre) => (
              <option key={genre.value} value={genre.value}>
                {genre.label}
              </option>
            ))}
          </select>
        </label>

        <label className="filters-field">
          <span>Исключить жанры</span>
          <input
            value={draft.excludeGenres}
            onChange={(event) => setDraft((current) => ({ ...current, excludeGenres: event.target.value }))}
            placeholder="comedy, horror"
          />
        </label>
      </div>

      <div className="filters-actions">
        <Button tone="primary" type="submit">
          Применить
        </Button>
        <Button
          tone="secondary"
          type="button"
          onClick={() => {
            setDraft(toDraft({ page: 1, pageSize: value.pageSize }))
            onReset()
          }}
        >
          Сбросить
        </Button>
      </div>
    </form>
  )
}
