package repository

import (
	"context"

	"mudrotop/internal/catalog/domain"
)

type Repository interface {
	ListMovies(ctx context.Context, query domain.MovieQuery) (domain.MoviePage, error)
	ListGenres(ctx context.Context) ([]domain.GenreOption, error)
}
