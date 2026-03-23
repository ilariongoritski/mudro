import { Button } from '@/shared/ui/Button'
import type { MoviePage } from '@/entities/movie/model/types'

type Props = {
  page?: MoviePage
  isLoading: boolean
  isError: boolean
  onPrevious: () => void
  onNext: () => void
}

export const MovieCatalog = ({ page, isLoading, isError, onPrevious, onNext }: Props) => {
  if (isLoading) {
    return <div className="surface state-box">Загружаем каталог фильмов...</div>
  }

  if (isError) {
    return <div className="surface state-box state-box--danger">Не удалось загрузить данные каталога.</div>
  }

  if (!page || page.items.length === 0) {
    return <div className="surface state-box">По этим фильтрам ничего не найдено.</div>
  }

  const totalPages = Math.max(1, Math.ceil(page.total / page.page_size))

  return (
    <section className="catalog-section">
      <div className="catalog-grid">
        {page.items.map((movie) => (
          <article key={movie.id} className="surface movie-card">
            <div className="movie-card__poster">
            {movie.poster_url ? (
              <img src={movie.poster_url} alt={movie.name} />
            ) : (
              <div className="movie-card__poster-fallback">Нет постера</div>
            )}
            </div>

            <div className="movie-card__body">
              <div className="movie-card__topline">
                <div>
                  <h2>{movie.name}</h2>
                  {movie.alternative_name ? <p>{movie.alternative_name}</p> : null}
                </div>
                {typeof movie.rating === 'number' ? <strong>{movie.rating.toFixed(1)}</strong> : null}
              </div>

              <div className="movie-card__meta">
                {movie.year ? <span>{movie.year}</span> : null}
                {movie.duration ? <span>{movie.duration} мин</span> : null}
              </div>

              <div className="movie-card__genres">
                {movie.genres.map((genre) => (
                  <span key={genre}>{genre}</span>
                ))}
              </div>

              {movie.description ? <p className="movie-card__description">{movie.description}</p> : null}

              {movie.kp_url ? (
                <a className="button button--link" href={movie.kp_url} target="_blank" rel="noreferrer">
                  Открыть карточку
                </a>
              ) : null}
            </div>
          </article>
        ))}
      </div>

      <footer className="surface catalog-pagination">
        <div>
          Страница {page.page} из {totalPages}. Всего фильмов: {page.total}
        </div>
        <div className="catalog-pagination__actions">
          <Button tone="neutral" type="button" onClick={onPrevious} disabled={page.page <= 1}>
            Назад
          </Button>
          <Button tone="neutral" type="button" onClick={onNext} disabled={page.page >= totalPages}>
            Вперёд
          </Button>
        </div>
      </footer>
    </section>
  )
}
