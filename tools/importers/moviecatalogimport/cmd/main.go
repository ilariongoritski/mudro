package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type movieRecord struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	AlternativeName string   `json:"alternative_name,omitempty"`
	Year            *int     `json:"year,omitempty"`
	Duration        *int     `json:"duration,omitempty"`
	Rating          *float64 `json:"rating,omitempty"`
	PosterURL       string   `json:"poster_url,omitempty"`
	Description     string   `json:"description,omitempty"`
	KPURL           string   `json:"kp_url,omitempty"`
	Genres          []string `json:"genres"`
}

type dataset struct {
	Movies []movieRecord `json:"movies"`
}

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	dsn := getenv("MOVIE_CATALOG_DB_DSN", "postgres://postgres:postgres@localhost:5434/movie_catalog?sslmode=disable")
	input := getenv("MOVIE_CATALOG_IMPORT_FILE", "out/movie-catalog.slim.json")

	raw, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("read import file: %w", err)
	}

	var payload dataset
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("decode import file: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	importedIDs := make([]string, 0, len(payload.Movies))

	for _, movie := range payload.Movies {
		if strings.TrimSpace(movie.ID) == "" || strings.TrimSpace(movie.Name) == "" {
			return errors.New("movie id and name are required")
		}

		importedIDs = append(importedIDs, movie.ID)

		_, err := tx.Exec(ctx, `
			insert into movie_catalog.movies (
				id, name, alternative_name, year, duration_minutes, rating_kp, poster_url, description, kp_url
			) values ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			on conflict (id) do update set
				name = excluded.name,
				alternative_name = excluded.alternative_name,
				year = excluded.year,
				duration_minutes = excluded.duration_minutes,
				rating_kp = excluded.rating_kp,
				poster_url = excluded.poster_url,
				description = excluded.description,
				kp_url = excluded.kp_url,
				updated_at = now()
		`, movie.ID, movie.Name, nullIfEmpty(movie.AlternativeName), movie.Year, movie.Duration, movie.Rating, nullIfEmpty(movie.PosterURL), nullIfEmpty(movie.Description), nullIfEmpty(movie.KPURL))
		if err != nil {
			return fmt.Errorf("upsert movie %s: %w", movie.ID, err)
		}

		_, err = tx.Exec(ctx, `delete from movie_catalog.movie_genres where movie_id = $1`, movie.ID)
		if err != nil {
			return fmt.Errorf("clear movie genres %s: %w", movie.ID, err)
		}

		for _, genre := range movie.Genres {
			slug := strings.ToLower(strings.TrimSpace(genre))
			if slug == "" {
				continue
			}
			label := genreLabel(slug)

			_, err = tx.Exec(ctx, `
				insert into movie_catalog.genres (slug, label)
				values ($1, $2)
				on conflict (slug) do update set label = excluded.label
			`, slug, label)
			if err != nil {
				return fmt.Errorf("upsert genre %s: %w", slug, err)
			}

			_, err = tx.Exec(ctx, `
				insert into movie_catalog.movie_genres (movie_id, genre_slug)
				values ($1, $2)
				on conflict do nothing
			`, movie.ID, slug)
			if err != nil {
				return fmt.Errorf("bind genre %s to movie %s: %w", slug, movie.ID, err)
			}
		}
	}

	if len(importedIDs) == 0 {
		if _, err := tx.Exec(ctx, `delete from movie_catalog.movie_genres`); err != nil {
			return fmt.Errorf("clear movie genres: %w", err)
		}
		if _, err := tx.Exec(ctx, `delete from movie_catalog.genres`); err != nil {
			return fmt.Errorf("clear genres: %w", err)
		}
		if _, err := tx.Exec(ctx, `delete from movie_catalog.movies`); err != nil {
			return fmt.Errorf("clear movies: %w", err)
		}
	} else {
		if _, err := tx.Exec(ctx, `
			delete from movie_catalog.movies
			where not (id = any($1::text[]))
		`, importedIDs); err != nil {
			return fmt.Errorf("delete stale movies: %w", err)
		}

		if _, err := tx.Exec(ctx, `
			delete from movie_catalog.genres g
			where not exists (
				select 1
				from movie_catalog.movie_genres mg
				where mg.genre_slug = g.slug
			)
		`); err != nil {
			return fmt.Errorf("delete orphan genres: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	fmt.Printf("imported %d movies\n", len(payload.Movies))
	return nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	return value
}

func genreLabel(slug string) string {
	if slug == "" {
		return ""
	}

	return strings.ToUpper(slug[:1]) + slug[1:]
}
