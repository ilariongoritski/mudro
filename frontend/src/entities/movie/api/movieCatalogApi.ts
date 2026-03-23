import { env } from '@/shared/config/env'
import { getJSON } from '@/shared/api/http'
import { buildSearchParams } from '@/shared/lib/query'
import type { GenreOption, MoviePage, MovieQuery } from '@/entities/movie/model/types'

type GenreResponse = {
  items: GenreOption[]
}

export async function fetchMovieCatalog(query: MovieQuery, signal?: AbortSignal): Promise<MoviePage> {
  const search = buildSearchParams({
    year_min: query.yearMin,
    duration_min: query.durationMin,
    include_genre: query.includeGenre,
    exclude_genres: query.excludeGenres,
    page: query.page,
    page_size: query.pageSize,
  })

  return getJSON<MoviePage>(`${env.apiBaseUrl}/movie-catalog/movies${search}`, signal)
}

export async function fetchMovieGenres(signal?: AbortSignal): Promise<GenreOption[]> {
  const data = await getJSON<GenreResponse>(`${env.apiBaseUrl}/movie-catalog/genres`, signal)
  return data.items
}
