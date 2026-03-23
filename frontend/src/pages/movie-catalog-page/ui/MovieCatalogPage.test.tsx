import { render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { MovieCatalogPage } from '@/pages/movie-catalog-page/ui/MovieCatalogPage'

const fetchMovieCatalog = vi.fn()
const fetchMovieGenres = vi.fn()

vi.mock('@/entities/movie/api/movieCatalogApi', () => ({
  fetchMovieCatalog: (...args: unknown[]) => fetchMovieCatalog(...args),
  fetchMovieGenres: (...args: unknown[]) => fetchMovieGenres(...args),
}))

describe('MovieCatalogPage', () => {
  beforeEach(() => {
    fetchMovieCatalog.mockReset()
    fetchMovieGenres.mockReset()
  })

  it('renders fetched movie data after initial loading', async () => {
    fetchMovieGenres.mockResolvedValue([{ value: 'drama', label: 'Drama' }])
    fetchMovieCatalog.mockResolvedValue({
      items: [
        {
          id: 'arrival',
          name: 'Arrival',
          genres: ['drama', 'sci-fi'],
        },
      ],
      total: 1,
      page: 1,
      page_size: 12,
    })

    render(<MovieCatalogPage />)

    expect(screen.getByText('Подгружаем экран и серверную выдачу...')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByText('Arrival')).toBeInTheDocument()
    })
  })

  it('renders error state when movie request fails', async () => {
    fetchMovieGenres.mockResolvedValue([])
    fetchMovieCatalog.mockRejectedValue(new Error('boom'))

    render(<MovieCatalogPage />)

    await waitFor(() => {
      expect(screen.getByText('Не удалось загрузить данные каталога.')).toBeInTheDocument()
    })
  })
})
