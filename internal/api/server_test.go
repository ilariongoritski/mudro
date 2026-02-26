package api

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	s := NewServer(nil)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestHandleHealth(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	s.Router().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("content-type = %q, want application/json", ct)
	}
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Fatalf("body = %q, want status ok", w.Body.String())
	}
}

func TestHandlePostsInvalidParams(t *testing.T) {
	s := &Server{}

	req := httptest.NewRequest("GET", "/api/posts?page=0", nil)
	w := httptest.NewRecorder()
	s.Router().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("page invalid status = %d, want 400", w.Code)
	}

	req = httptest.NewRequest("GET", "/api/posts?before_ts=2026-02-24T12:00:00Z", nil)
	w = httptest.NewRecorder()
	s.Router().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("cursor invalid status = %d, want 400", w.Code)
	}
}

func TestHandleFeedInvalidCursor(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest("GET", "/feed?before_ts=bad&before_id=1", nil)
	w := httptest.NewRecorder()

	s.Router().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestParseLimit(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 50},
		{"10", 10},
		{"0", 50},
		{"-1", 50},
		{"x", 50},
		{"1000", 200},
	}

	for _, tc := range cases {
		if got := parseLimit(tc.in); got != tc.want {
			t.Fatalf("parseLimit(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestParsePage(t *testing.T) {
	p, err := parsePage("")
	if err != nil || p != nil {
		t.Fatalf("empty page: p=%v err=%v", p, err)
	}

	p, err = parsePage("2")
	if err != nil || p == nil || *p != 2 {
		t.Fatalf("valid page: p=%v err=%v", p, err)
	}

	if _, err := parsePage("0"); err == nil {
		t.Fatal("expected error for page=0")
	}
}

func TestParseCursor(t *testing.T) {
	ts := "2026-02-24T12:00:00Z"
	id := "123"
	gotTS, gotID, err := parseCursor(ts, id)
	if err != nil {
		t.Fatalf("parseCursor valid: %v", err)
	}
	if gotTS == nil || gotTS.Format(time.RFC3339) != ts {
		t.Fatalf("timestamp mismatch: got=%v", gotTS)
	}
	if gotID == nil || *gotID != 123 {
		t.Fatalf("id mismatch: got=%v", gotID)
	}

	if _, _, err := parseCursor(ts, ""); err == nil {
		t.Fatal("expected error when before_id missing")
	}
	if _, _, err := parseCursor("bad", id); err == nil {
		t.Fatal("expected error for bad ts")
	}
}

func TestCompactText(t *testing.T) {
	if got := compactText(nil); got != "" {
		t.Fatalf("compactText(nil) = %q", got)
	}
	s := "  hello  "
	if got := compactText(&s); got != "hello" {
		t.Fatalf("compactText = %q, want hello", got)
	}
}
