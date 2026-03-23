package service

import (
	"context"
	"testing"

	"mudrotop/internal/catalog/domain"
	"mudrotop/internal/catalog/repository"
)

type captureRepository struct {
	lastQuery domain.MovieQuery
}

func (c *captureRepository) ListMovies(_ context.Context, query domain.MovieQuery) (domain.MoviePage, error) {
	c.lastQuery = query
	return domain.MoviePage{Page: query.Page, PageSize: query.PageSize}, nil
}

func (c *captureRepository) ListGenres(_ context.Context) ([]domain.GenreOption, error) {
	return nil, nil
}

var _ repository.Repository = (*captureRepository)(nil)

func TestListMoviesNormalizesQuery(t *testing.T) {
	repo := &captureRepository{}
	service := NewCatalogService(repo)

	_, err := service.ListMovies(context.Background(), domain.MovieQuery{
		IncludeGenre:  " Drama ",
		ExcludeGenres: []string{" horror ", "", "HORROR", " comedy "},
		Page:          0,
		PageSize:      1000,
	})
	if err != nil {
		t.Fatalf("ListMovies returned error: %v", err)
	}

	if repo.lastQuery.IncludeGenre != "drama" {
		t.Fatalf("include genre = %q", repo.lastQuery.IncludeGenre)
	}

	if repo.lastQuery.Page != 1 {
		t.Fatalf("page = %d", repo.lastQuery.Page)
	}

	if repo.lastQuery.PageSize != 100 {
		t.Fatalf("pageSize = %d", repo.lastQuery.PageSize)
	}

	if len(repo.lastQuery.ExcludeGenres) != 2 || repo.lastQuery.ExcludeGenres[0] != "horror" || repo.lastQuery.ExcludeGenres[1] != "comedy" {
		t.Fatalf("excludeGenres = %#v", repo.lastQuery.ExcludeGenres)
	}
}
