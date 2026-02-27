package main

import (
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

