package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseTGMessageID(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{in: "1269", want: 1269, ok: true},
		{in: "tg/1269", want: 1269, ok: true},
		{in: "tg:1269", want: 1269, ok: true},
		{in: "  tg/42  ", want: 42, ok: true},
		{in: "tg/", want: 0, ok: false},
		{in: "abc", want: 0, ok: false},
	}
	for _, tc := range cases {
		got, ok := parseTGMessageID(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("parseTGMessageID(%q) = (%d,%v), want (%d,%v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestResolvePostLinkRejectsCommentBeforePost(t *testing.T) {
	t.Parallel()
	postTime := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	commentTime := time.Date(2024, 3, 13, 11, 15, 43, 0, time.UTC)

	postByMsgID := map[int64]tgPostRef{
		1269: {PostID: 241, PublishedAt: postTime},
	}
	msgs := map[int64]htmlMessage{}

	postID, parent, ok := resolvePostLink(msgs, postByMsgID, 1269, commentTime)
	if ok {
		t.Fatalf("expected unresolved link, got postID=%d parent=%v", postID, parent)
	}
}

func TestMergeJSONMessagesAddsCommentMedia(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "result.json")
	payload := `{
	  "messages": [
	    {
	      "id": 5,
	      "type": "message",
	      "date_unixtime": "1700000000",
	      "from": "Катя",
	      "reply_to_message_id": 3,
	      "text": "",
	      "media_type": "sticker",
	      "file": "stickers/sticker.webp",
	      "thumbnail": "stickers/sticker.webp_thumb.jpg",
	      "sticker_emoji": "🔥"
	    }
	  ]
	}`
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("write result.json: %v", err)
	}

	msgs, ids, err := mergeJSONMessages(path, map[int64]htmlMessage{}, nil)
	if err != nil {
		t.Fatalf("mergeJSONMessages: %v", err)
	}
	if len(ids) != 1 || ids[0] != 5 {
		t.Fatalf("ids = %v, want [5]", ids)
	}
	msg, ok := msgs[5]
	if !ok {
		t.Fatal("message 5 not found")
	}
	if msg.ReplyToID != 3 {
		t.Fatalf("ReplyToID = %d, want 3", msg.ReplyToID)
	}
	if len(msg.Media) != 1 {
		t.Fatalf("len(media) = %d, want 1", len(msg.Media))
	}
	if msg.Media[0].URL != "stickers/sticker.webp" {
		t.Fatalf("media url = %q, want stickers/sticker.webp", msg.Media[0].URL)
	}
	if msg.Media[0].Kind != "gif" {
		t.Fatalf("media kind = %q, want gif", msg.Media[0].Kind)
	}
}
