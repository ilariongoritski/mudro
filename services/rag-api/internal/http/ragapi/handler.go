package ragapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/goritskimihail/mudro/internal/rag"
	"github.com/goritskimihail/mudro/pkg/httputil"
)

type asker interface {
	Ask(context.Context, string) (rag.Answer, error)
}

type Handler struct{ service asker }

func NewHandler(service asker) http.Handler {
	h := &Handler{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /internal/rag/ask", h.ask)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "rag-api"})
}

func (h *Handler) ask(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Question string `json:"question"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON request"})
		return
	}
	request.Question = strings.TrimSpace(request.Question)
	if request.Question == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "question is required"})
		return
	}
	answer, err := h.service.Ask(r.Context(), request.Question)
	if errors.Is(err, rag.ErrInsufficientContext) {
		httputil.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "insufficient documentation context", "sources": answer.Sources})
		return
	}
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadGateway, map[string]string{"error": "RAG request failed"})
		return
	}
	httputil.WriteJSON(w, http.StatusOK, answer)
}
