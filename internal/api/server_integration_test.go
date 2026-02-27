package api

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testDBServer(t *testing.T) *Server {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, config.DSN())
	if err != nil {
		t.Skipf("skip integration test: db connect: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("skip integration test: db ping: %v", err)
	}

	_, err = pool.Exec(ctx, `truncate table post_reactions, posts restart identity cascade`)
	if err != nil {
		t.Skipf("skip integration test: truncate: %v", err)
	}

	_, err = pool.Exec(ctx, `
		insert into posts (source, source_post_id, published_at, text, likes_count, views_count, comments_count)
		values
		('vk', '100', now() - interval '1 hour', 'post a', 10, 100, 1),
		('tg', '200', now() - interval '2 hour', 'post b', 20, 200, 2)
	`)
	if err != nil {
		t.Fatalf("seed posts: %v", err)
	}
	_, err = pool.Exec(ctx, `insert into post_reactions (post_id, emoji, count) values (1, '👍', 3)`)
	if err != nil {
		t.Fatalf("seed reactions: %v", err)
	}

	return NewServer(pool)
}

func TestLoadPostsAndFrontHandlersIntegration(t *testing.T) {
	s := testDBServer(t)
	ctx := context.Background()

	posts, next, err := s.loadPosts(ctx, nil, nil, nil, 10, "", "desc")
	if err != nil {
		t.Fatalf("loadPosts: %v", err)
	}
	if len(posts) != 2 || next == nil {
		t.Fatalf("unexpected posts len=%d next=%v", len(posts), next)
	}
	if posts[0].Reactions["👍"] != 3 {
		t.Fatalf("expected reactions on first post: %+v", posts[0].Reactions)
	}

	src, err := s.loadSourceStats(ctx)
	if err != nil {
		t.Fatalf("loadSourceStats: %v", err)
	}
	if len(src) != 2 {
		t.Fatalf("source stats len=%d", len(src))
	}

	req := httptest.NewRequest("GET", "/api/front?limit=10", nil)
	w := httptest.NewRecorder()
	s.Router().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("front status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"total_posts": 2`) {
		t.Fatalf("front body=%s", w.Body.String())
	}
}

func TestHandleFeedIntegration(t *testing.T) {
	s := testDBServer(t)

	req := httptest.NewRequest("GET", "/feed?limit=10", nil)
	w := httptest.NewRecorder()
	s.Router().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("feed status=%d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "Mudro feed") || !strings.Contains(body, "post a") {
		t.Fatalf("feed body=%s", body)
	}
}
