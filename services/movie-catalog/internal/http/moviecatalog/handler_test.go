package moviecatalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goritskimihail/mudro/internal/catalog/domain"
	"github.com/goritskimihail/mudro/internal/catalog/repository"
	"github.com/goritskimihail/mudro/internal/catalog/service"
)

type stubRepository struct{}

func (stubRepository) ListGenres(_ context.Context) ([]domain.GenreOption, error) {
	return []domain.GenreOption{
		{Value: "drama", Label: "Drama"},
		{Value: "sci-fi", Label: "Sci-fi"},
	}, nil
}

func (stubRepository) ListMovies(_ context.Context, query domain.MovieQuery) (domain.MoviePage, error) {
	return domain.MoviePage{
		Items: []domain.MovieSummary{
			{
				ID:     "arrival",
				Name:   "Arrival",
				Genres: []string{"Drama", "Sci-fi"},
			},
		},
		Total:    1,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

var _ repository.Repository = stubRepository{}

func TestHealth(t *testing.T) {
	handler := NewHandler(service.NewCatalogService(stubRepository{}), func(context.Context) error { return nil })
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestMoviesRejectBadPage(t *testing.T) {
	handler := NewHandler(service.NewCatalogService(stubRepository{}), nil)
	req := httptest.NewRequest(http.MethodGet, "/api/movie-catalog/movies?page=0", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestMoviesReturnPage(t *testing.T) {
	handler := NewHandler(service.NewCatalogService(stubRepository{}), nil)
	req := httptest.NewRequest(http.MethodGet, "/api/movie-catalog/movies?page=2&page_size=24", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"name":"Arrival"`) || !strings.Contains(body, `"page":2`) {
		t.Fatalf("body = %q", body)
	}
}

func TestHealthReturnsUnavailableWhenPingFails(t *testing.T) {
	handler := NewHandler(service.NewCatalogService(stubRepository{}), func(context.Context) error {
		return context.DeadlineExceeded
	})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}
