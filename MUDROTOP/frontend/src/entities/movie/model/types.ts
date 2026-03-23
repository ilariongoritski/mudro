export interface MovieSummary {
  id: string
  name: string
  alternative_name?: string
  year?: number
  duration?: number
  rating?: number
  poster_url?: string
  description?: string
  kp_url?: string
  genres: string[]
}

export interface GenreOption {
  value: string
  label: string
}

export interface MoviePage {
  items: MovieSummary[]
  total: number
  page: number
  page_size: number
}

export interface MovieQuery {
  yearMin?: number
  durationMin?: number
  includeGenre?: string
  excludeGenres?: string[]
  page: number
  pageSize: number
}
