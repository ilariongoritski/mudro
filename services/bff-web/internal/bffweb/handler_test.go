package bffweb

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goritskimihail/mudro/internal/posts"
)

type fakeTimeline struct {
	items []posts.Post
	next  *posts.Cursor
	err   error
}

func (f fakeTimeline) LoadPosts(_ context.Context, _ *time.Time, _ *int64, _ *int, _ int, _ string, _ posts.SortOrder, _ string) ([]posts.Post, *posts.Cursor, error) {
	return f.items, f.next, f.err
}

func TestTimeline(t *testing.T) {
	handler := NewHandler(fakeTimeline{
		items: []posts.Post{
			{
				ID:           1,
				Source:       "tg",
				SourcePostID: "42",
				PublishedAt:  time.Unix(1710000000, 0).UTC(),
				LikesCount:   5,
			},
		},
	}, "")

	req := httptest.NewRequest(http.MethodGet, "/api/bff/web/v1/timeline?limit=5", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload struct {
		Total int `json:"total"`
		Items []struct {
			ID int64 `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if payload.Total != 1 || len(payload.Items) != 1 || payload.Items[0].ID != 1 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestOrchestrationProxy(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orchestration/status" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"branch":"main","commit":"abc123"}`))
	}))
	defer upstream.Close()

	handler := NewHandler(fakeTimeline{}, upstream.URL)
	req := httptest.NewRequest(http.MethodGet, "/api/bff/web/v1/orchestration/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"branch":"main"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestCasinoWidget(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/casino/balance" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"balance":2500,"rtp":95.5,"currency":"credits"}`))
	}))
	defer upstream.Close()

	handler := NewHandler(fakeTimeline{}, upstream.URL)
	req := httptest.NewRequest(http.MethodGet, "/api/bff/web/v1/casino/widget", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"balance":2500`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"open-casino"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}
