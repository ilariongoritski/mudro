package moviecatalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/goritskimihail/mudro/internal/catalog/domain"
	"github.com/goritskimihail/mudro/internal/catalog/service"
)

type Handler struct {
	service *service.CatalogService
	ping    func(context.Context) error
}

func NewHandler(catalogService *service.CatalogService, ping func(context.Context) error) http.Handler {
	handler := &Handler{service: catalogService, ping: ping}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.handleHealth)
	mux.HandleFunc("/api/movie-catalog/genres", handler.handleGenres)
	mux.HandleFunc("/api/movie-catalog/movies", handler.handleMovies)

	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if h.ping != nil {
		if err := h.ping(r.Context()); err != nil {
			writeError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleGenres(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	genres, err := h.service.ListGenres(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load genres")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": genres})
}

func (h *Handler) handleMovies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query, err := decodeQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	page, err := h.service.ListMovies(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load movies")
		return
	}

	writeJSON(w, http.StatusOK, page)
}

func decodeQuery(r *http.Request) (domain.MovieQuery, error) {
	values := r.URL.Query()
	query := domain.MovieQuery{
		IncludeGenre:  strings.TrimSpace(values.Get("include_genre")),
		ExcludeGenres: parseList(values["exclude_genres"]),
		Page:          1,
		PageSize:      12,
	}

	if raw := values.Get("year_min"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return domain.MovieQuery{}, errors.New("year_min must be an integer")
		}
		query.YearMin = &value
	}

	if raw := values.Get("duration_min"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return domain.MovieQuery{}, errors.New("duration_min must be an integer")
		}
		query.DurationMin = &value
	}

	if raw := values.Get("page"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return domain.MovieQuery{}, errors.New("page must be a positive integer")
		}
		query.Page = value
	}

	if raw := values.Get("page_size"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return domain.MovieQuery{}, errors.New("page_size must be a positive integer")
		}
		if value > 100 {
			return domain.MovieQuery{}, errors.New("page_size must not exceed 100")
		}
		query.PageSize = value
	}

	return query, nil
}

func parseList(chunks []string) []string {
	if len(chunks) == 0 {
		return nil
	}

	items := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		for _, item := range strings.Split(chunk, ",") {
			value := strings.TrimSpace(item)
			if value == "" {
				continue
			}
			items = append(items, value)
		}
	}

	return items
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
