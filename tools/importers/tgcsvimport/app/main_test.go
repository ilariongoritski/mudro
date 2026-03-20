package app

import (
	"testing"
	"time"
)

func TestNormalizeCSVMessage(t *testing.T) {
	if got := normalizeCSVMessage("MediaMessage"); got != "" {
		t.Fatalf("normalizeCSVMessage(MediaMessage) = %q, want empty", got)
	}
	if got := normalizeCSVMessage("  Привет  "); got != "Привет" {
		t.Fatalf("normalizeCSVMessage(trim) = %q", got)
	}
}

func TestParseCSVReactions(t *testing.T) {
	reactions := parseCSVReactions("{':pill:': 1, 'ReactionCustomEmoji(document_id=1)': 3, 'ReactionPaid()': 2}")

	if reactions["emoji::pill:"] != 1 {
		t.Fatalf("pill reaction missing: %#v", reactions)
	}
	if reactions["custom:ReactionCustomEmoji(document_id=1)"] != 3 {
		t.Fatalf("custom reaction missing: %#v", reactions)
	}
	if reactions["unknown:ReactionPaid()"] != 2 {
		t.Fatalf("paid reaction missing: %#v", reactions)
	}
}

func TestParseCSVDate(t *testing.T) {
	ts, err := parseCSVDate("2026-03-13 17:45:56+00:00")
	if err != nil {
		t.Fatalf("parseCSVDate returned error: %v", err)
	}
	if got := ts.UTC().Format(time.RFC3339); got != "2026-03-13T17:45:56Z" {
		t.Fatalf("parsed ts = %s", got)
	}
}
