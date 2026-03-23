package postgrescatalog

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goritskimihail/mudro/internal/catalog/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListGenres(ctx context.Context) ([]domain.GenreOption, error) {
	rows, err := r.pool.Query(ctx, `
		select slug, label
		from movie_catalog.genres
		order by label asc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	genres := make([]domain.GenreOption, 0)
	for rows.Next() {
		var item domain.GenreOption
		if err := rows.Scan(&item.Value, &item.Label); err != nil {
			return nil, err
		}
		genres = append(genres, item)
	}

	return genres, rows.Err()
}

func (r *Repository) ListMovies(ctx context.Context, query domain.MovieQuery) (domain.MoviePage, error) {
	excludeGenres := query.ExcludeGenres
	if excludeGenres == nil {
		excludeGenres = []string{}
	}

	total, err := r.countMovies(ctx, query)
	if err != nil {
		return domain.MoviePage{}, err
	}

	rows, err := r.pool.Query(ctx, `
		select
			m.id,
			m.name,
			coalesce(m.alternative_name, ''),
			m.year,
			m.duration_minutes,
			m.rating_kp,
			coalesce(m.poster_url, ''),
			coalesce(m.description, ''),
			coalesce(m.kp_url, ''),
			coalesce(array_agg(g.label order by g.label) filter (where g.label is not null), '{}')
		from movie_catalog.movies m
		left join movie_catalog.movie_genres mg on mg.movie_id = m.id
		left join movie_catalog.genres g on g.slug = mg.genre_slug
		where ($1::int is null or m.year >= $1)
		  and ($2::int is null or m.duration_minutes >= $2)
		  and ($3::text = '' or exists (
			select 1
			from movie_catalog.movie_genres mg2
			where mg2.movie_id = m.id and mg2.genre_slug = $3
		  ))
		  and (
			coalesce(cardinality($4::text[]), 0) = 0 or not exists (
				select 1
				from movie_catalog.movie_genres mg3
				where mg3.movie_id = m.id and mg3.genre_slug = any($4::text[])
			)
		  )
		group by m.id, m.name, m.alternative_name, m.year, m.duration_minutes, m.rating_kp, m.poster_url, m.description, m.kp_url
		order by m.rating_kp desc nulls last, m.year desc nulls last, m.name asc
		limit $5 offset $6
	`, query.YearMin, query.DurationMin, query.IncludeGenre, excludeGenres, query.PageSize, (query.Page-1)*query.PageSize)
	if err != nil {
		return domain.MoviePage{}, err
	}
	defer rows.Close()

	items := make([]domain.MovieSummary, 0, query.PageSize)
	for rows.Next() {
		var (
			item            domain.MovieSummary
			alternativeName string
			year            sql.NullInt32
			duration        sql.NullInt32
			rating          sql.NullFloat64
			posterURL       string
			description     string
			kpURL           string
			genres          []string
		)

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&alternativeName,
			&year,
			&duration,
			&rating,
			&posterURL,
			&description,
			&kpURL,
			&genres,
		); err != nil {
			return domain.MoviePage{}, err
		}

		item.AlternativeName = alternativeName
		if year.Valid {
			value := int(year.Int32)
			item.Year = &value
		}
		if duration.Valid {
			value := int(duration.Int32)
			item.Duration = &value
		}
		if rating.Valid {
			value := rating.Float64
			item.Rating = &value
		}
		item.PosterURL = posterURL
		item.Description = description
		item.KPURL = kpURL
		item.Genres = genres
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return domain.MoviePage{}, err
	}

	return domain.MoviePage{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (r *Repository) countMovies(ctx context.Context, query domain.MovieQuery) (int, error) {
	excludeGenres := query.ExcludeGenres
	if excludeGenres == nil {
		excludeGenres = []string{}
	}

	var total int
	err := r.pool.QueryRow(ctx, `
		select count(*)
		from movie_catalog.movies m
		where ($1::int is null or m.year >= $1)
		  and ($2::int is null or m.duration_minutes >= $2)
		  and ($3::text = '' or exists (
			select 1
			from movie_catalog.movie_genres mg2
			where mg2.movie_id = m.id and mg2.genre_slug = $3
		  ))
		  and (
			coalesce(cardinality($4::text[]), 0) = 0 or not exists (
				select 1
				from movie_catalog.movie_genres mg3
				where mg3.movie_id = m.id and mg3.genre_slug = any($4::text[])
			)
		  )
	`, query.YearMin, query.DurationMin, query.IncludeGenre, excludeGenres).Scan(&total)

	return total, err
}
