import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { MovieCatalog } from '@/widgets/movie-catalog/ui/MovieCatalog'

describe('MovieCatalog', () => {
  it('renders paging buttons and movie card link', () => {
    const onPrevious = vi.fn()
    const onNext = vi.fn()

    render(
      <MovieCatalog
        page={{
          items: [
            {
              id: 'arrival',
              name: 'Arrival',
              year: 2016,
              duration: 116,
              rating: 7.9,
              kp_url: 'https://example.com/arrival',
              genres: ['drama', 'sci-fi'],
            },
          ],
          total: 30,
          page: 2,
          page_size: 12,
        }}
        isLoading={false}
        isError={false}
        onPrevious={onPrevious}
        onNext={onNext}
      />,
    )

    expect(screen.getByText('Arrival')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Открыть карточку' })).toHaveAttribute(
      'href',
      'https://example.com/arrival',
    )

    fireEvent.click(screen.getByRole('button', { name: 'Назад' }))
    fireEvent.click(screen.getByRole('button', { name: 'Вперёд' }))

    expect(onPrevious).toHaveBeenCalledTimes(1)
    expect(onNext).toHaveBeenCalledTimes(1)
  })
})
