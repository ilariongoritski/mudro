import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { MovieFilters } from '@/features/movie-filters/ui/MovieFilters'

describe('MovieFilters', () => {
  it('applies explicit button-driven filters', () => {
    const onApply = vi.fn()
    const onReset = vi.fn()

    render(
      <MovieFilters
        genres={[
          { value: 'drama', label: 'Drama' },
          { value: 'sci-fi', label: 'Sci-fi' },
        ]}
        value={{ page: 1, pageSize: 12 }}
        onApply={onApply}
        onReset={onReset}
      />,
    )

    fireEvent.change(screen.getByPlaceholderText('2010'), { target: { value: '2015' } })
    fireEvent.change(screen.getByPlaceholderText('90'), { target: { value: '120' } })
    fireEvent.change(screen.getByDisplayValue('Любой'), { target: { value: 'drama' } })
    fireEvent.change(screen.getByPlaceholderText('comedy, horror'), { target: { value: 'comedy, horror' } })
    fireEvent.click(screen.getByRole('button', { name: 'Применить' }))

    expect(onApply).toHaveBeenCalledWith({
      yearMin: 2015,
      durationMin: 120,
      includeGenre: 'drama',
      excludeGenres: ['comedy', 'horror'],
      page: 1,
      pageSize: 12,
    })
  })

  it('resets only through explicit reset button', () => {
    const onApply = vi.fn()
    const onReset = vi.fn()

    render(
      <MovieFilters
        genres={[]}
        value={{ page: 2, pageSize: 24, yearMin: 2010 }}
        onApply={onApply}
        onReset={onReset}
      />,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Сбросить' }))

    expect(onReset).toHaveBeenCalledTimes(1)
  })
})
