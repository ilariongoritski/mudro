package ragapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goritskimihail/mudro/internal/rag"
)

type stubAsker struct {
	answer rag.Answer
	err    error
}

func (s stubAsker) Ask(context.Context, string) (rag.Answer, error) { return s.answer, s.err }

func TestHealth(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewHandler(stubAsker{}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"service":"rag-api"`) {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
}
func TestAskRejectsEmptyQuestion(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewHandler(stubAsker{}).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/internal/rag/ask", strings.NewReader(`{"question":" "}`)))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
}
func TestAskReturnsSources(t *testing.T) {
	service := stubAsker{answer: rag.Answer{Answer: "Ответ [docs/rag-mvp.md]", Grounded: true, Sources: []rag.Source{{Path: "docs/rag-mvp.md", Excerpt: "source"}}}}
	recorder := httptest.NewRecorder()
	NewHandler(service).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/internal/rag/ask", strings.NewReader(`{"question":"Как работает RAG?"}`)))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"grounded":true`) || !strings.Contains(recorder.Body.String(), `docs/rag-mvp.md`) {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
}
func TestAskReportsInsufficientContext(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewHandler(stubAsker{answer: rag.Answer{Sources: []rag.Source{}}, err: rag.ErrInsufficientContext}).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/internal/rag/ask", strings.NewReader(`{"question":"нет в документации"}`)))
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", recorder.Code)
	}
}
