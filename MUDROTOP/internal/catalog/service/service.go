package service

import (
	"context"
	"strings"

	"mudrotop/internal/catalog/domain"
	"mudrotop/internal/catalog/repository"
)

type CatalogService struct {
	repository repository.Repository
}

func NewCatalogService(repo repository.Repository) *CatalogService {
	return &CatalogService{repository: repo}
}

func (s *CatalogService) ListMovies(ctx context.Context, query domain.MovieQuery) (domain.MoviePage, error) {
	normalized := normalizeQuery(query)
	return s.repository.ListMovies(ctx, normalized)
}

func (s *CatalogService) ListGenres(ctx context.Context) ([]domain.GenreOption, error) {
	return s.repository.ListGenres(ctx)
}

func normalizeQuery(query domain.MovieQuery) domain.MovieQuery {
	if query.Page < 1 {
		query.Page = 1
	}

	if query.PageSize < 1 {
		query.PageSize = 12
	}

	if query.PageSize > 100 {
		query.PageSize = 100
	}

	query.IncludeGenre = normalizeGenre(query.IncludeGenre)

	if len(query.ExcludeGenres) > 0 {
		seen := make(map[string]struct{}, len(query.ExcludeGenres))
		normalized := make([]string, 0, len(query.ExcludeGenres))

		for _, genre := range query.ExcludeGenres {
			value := normalizeGenre(genre)
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			normalized = append(normalized, value)
		}

		query.ExcludeGenres = normalized
	}

	return query
}

func normalizeGenre(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
