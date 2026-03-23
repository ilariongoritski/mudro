import { startTransition, useDeferredValue, useEffect, useMemo, useState } from 'react'

import { fetchMovieCatalog, fetchMovieGenres } from '@/entities/movie/api/movieCatalogApi'
import type { GenreOption, MoviePage, MovieQuery } from '@/entities/movie/model/types'
import { MovieFilters } from '@/features/movie-filters/ui/MovieFilters'
import { MovieCatalog } from '@/widgets/movie-catalog/ui/MovieCatalog'

const initialQuery: MovieQuery = {
  page: 1,
  pageSize: 12,
}

export const MovieCatalogPage = () => {
  const [query, setQuery] = useState<MovieQuery>(initialQuery)
  const [genres, setGenres] = useState<GenreOption[]>([])
  const [page, setPage] = useState<MoviePage>()
  const [isLoading, setIsLoading] = useState(true)
  const [isError, setIsError] = useState(false)

  const deferredQuery = useDeferredValue(query)

  useEffect(() => {
    const controller = new AbortController()

    fetchMovieGenres(controller.signal)
      .then((items) => {
        if (!controller.signal.aborted) {
          setGenres(items)
        }
      })
      .catch(() => {
        if (!controller.signal.aborted) {
          setGenres([])
        }
      })

    return () => {
      controller.abort()
    }
  }, [])

  useEffect(() => {
    const controller = new AbortController()

    setIsLoading(true)
    setIsError(false)

    fetchMovieCatalog(deferredQuery, controller.signal)
      .then((nextPage) => {
        if (!controller.signal.aborted) {
          setPage(nextPage)
        }
      })
      .catch(() => {
        if (!controller.signal.aborted) {
          setIsError(true)
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setIsLoading(false)
        }
      })

    return () => {
      controller.abort()
    }
  }, [deferredQuery])

  const summary = useMemo(() => {
    if (isLoading) {
      return 'Подгружаем экран и серверную выдачу...'
    }

    return 'Фильтры, карточки и пагинация уже работают через typed HTTP contract.'
  }, [isLoading])

  return (
    <main className="page-shell">
      <header className="page-hero">
        <p className="page-kicker">movie-catalog</p>
        <h1 className="page-title">Каталог фильмов</h1>
        <p className="page-subtitle">
          Подготовка к слиянию в `mudro`: нормальный серверный доступ к данным, понятная визуальная иерархия и те же
          хорошие сценарии кнопок.
        </p>
        <p className="page-subtitle">{summary}</p>
      </header>

      <MovieFilters
        genres={genres}
        value={query}
        onApply={(next) => {
          startTransition(() => {
            setQuery(next)
          })
        }}
        onReset={() => {
          startTransition(() => {
            setQuery(initialQuery)
          })
        }}
      />

      <MovieCatalog
        page={page}
        isLoading={isLoading}
        isError={isError}
        onPrevious={() =>
          startTransition(() => {
            setQuery((current) => ({ ...current, page: Math.max(1, current.page - 1) }))
          })
        }
        onNext={() =>
          startTransition(() => {
            setQuery((current) => ({ ...current, page: current.page + 1 }))
          })
        }
      />
    </main>
  )
}
