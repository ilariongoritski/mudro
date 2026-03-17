package api

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	s := NewServer(nil, nil)
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

func TestHandleHealthCORS(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest("GET", "/healthz", nil)
	req.Header.Set("Origin", "http://127.0.0.1:4174")
	w := httptest.NewRecorder()

	s.Router().ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", got)
	}
}

func TestOptionsCORSPreflight(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest("OPTIONS", "/api/front", nil)
	req.Header.Set("Origin", "http://127.0.0.1:4174")
	w := httptest.NewRecorder()

	s.Router().ServeHTTP(w, req)

	if w.Code != 204 {
		t.Fatalf("status = %d, want 204", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("expected Access-Control-Allow-Methods header")
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

func TestReactionLabel(t *testing.T) {
	cases := map[string]string{
		"emoji:🔥":  "🔥",
		"custom:x": "✨",
		"unknown:": "?",
		"":         "?",
		"plain":    "plain",
	}
	for in, want := range cases {
		if got := reactionLabel(in); got != want {
			t.Fatalf("reactionLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildFeedReactions(t *testing.T) {
	got := buildFeedReactions(map[string]int{
		"emoji:🔥":   3,
		"emoji:❤️":  5,
		"custom:id": 2,
		"emoji:0":   0,
	})
	want := []feedReaction{
		{Label: "❤️", Count: 5, Raw: "emoji:❤️"},
		{Label: "🔥", Count: 3, Raw: "emoji:🔥"},
		{Label: "✨", Count: 2, Raw: "custom:id"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildFeedReactions() = %#v, want %#v", got, want)
	}
}

func TestParseMediaItemsDetectsKindsAndTitles(t *testing.T) {
	raw := json.RawMessage(`[
	  {"kind":"file","url":"photos/video_1.mp4","extra":{"media_type":"video_file","file_name":"video_1.mp4"}},
	  {"kind":"file","url":"https://cdn.example.com/pic.png"},
	  {"kind":"gif","url":"stickers/sticker.webp","preview_url":"stickers/sticker.webp_thumb.jpg","extra":{"media_type":"sticker","file_name":"sticker.webp"}},
	  {"kind":"audio","title":"Artist - Track"},
	  {"kind":"doc","url":"https://cdn.example.com/file.pdf"},
	  {"kind":"link","url":"https://example.com/readme"}
	]`)

	got := parseMediaItems(raw)
	if len(got) != 6 {
		t.Fatalf("len=%d, want 6", len(got))
	}

	if !got[0].IsVideo || got[0].URL != "/media/photos/video_1.mp4" || got[0].Title != "video_1.mp4" {
		t.Fatalf("item0 unexpected: %+v", got[0])
	}
	if !got[1].IsImage || got[1].URL == "" {
		t.Fatalf("item1 unexpected: %+v", got[1])
	}
	if !got[2].IsImage || got[2].URL != "/media/stickers/sticker.webp" {
		t.Fatalf("item2 unexpected: %+v", got[2])
	}
	if !got[3].IsAudio || got[3].Title == "" {
		t.Fatalf("item3 unexpected: %+v", got[3])
	}
	if !got[4].IsDocument {
		t.Fatalf("item4 unexpected: %+v", got[4])
	}
	if !got[5].IsLink || got[5].URL == "" {
		t.Fatalf("item5 unexpected: %+v", got[5])
	}
}

func TestNormalizeMediaURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "https://example.com/a.jpg", want: "https://example.com/a.jpg"},
		{in: "http://example.com/a.jpg", want: "http://example.com/a.jpg"},
		{in: "missing://file.mp4", want: ""},
		{in: "photos/file.jpg", want: "/media/photos/file.jpg"},
	}
	for _, tc := range cases {
		if got := normalizeMediaURL(tc.in); got != tc.want {
			t.Fatalf("normalizeMediaURL(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizePostMediaJSON(t *testing.T) {
	raw := json.RawMessage(`[
	  {"kind":"photo","url":"photos/file.jpg","extra":{"file_name":"file.jpg"}},
	  {"kind":"video","url":"https://cdn.example.com/video.mp4"}
	]`)

	got := normalizePostMediaJSON(raw)
	if len(got) == 0 {
		t.Fatal("normalizePostMediaJSON returned empty payload")
	}
	if !strings.Contains(string(got), `"url":"/media/photos/file.jpg"`) {
		t.Fatalf("normalized payload = %s, want lowercase url key", string(got))
	}

	var decoded []feedMediaItem
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("unmarshal normalized media: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("len(decoded) = %d, want 2", len(decoded))
	}
	if decoded[0].URL != "/media/photos/file.jpg" {
		t.Fatalf("decoded[0].URL = %q, want /media/photos/file.jpg", decoded[0].URL)
	}
	if decoded[1].URL != "https://cdn.example.com/video.mp4" {
		t.Fatalf("decoded[1].URL = %q, want original https URL", decoded[1].URL)
	}
}

func TestBuildPostsVisibilityWhereSkipsTGFilterForVK(t *testing.T) {
	s := &Server{tgVisiblePostIDs: []string{"1", "2"}}

	where, args := s.buildPostsVisibilityWhere("vk", nil)

	if where != " where source = $1" {
		t.Fatalf("where = %q, want only source filter", where)
	}
	if len(args) != 1 || args[0] != "vk" {
		t.Fatalf("args = %#v, want only vk source", args)
	}
}

func TestBuildPostsVisibilityWhereKeepsTGFilterForTG(t *testing.T) {
	s := &Server{tgVisiblePostIDs: []string{"1", "2"}}

	where, args := s.buildPostsVisibilityWhere("tg", nil)

	if where != " where source = $1 and source_post_id = any($2)" {
		t.Fatalf("where = %q", where)
	}
	if len(args) != 2 {
		t.Fatalf("args len = %d, want 2", len(args))
	}
}
