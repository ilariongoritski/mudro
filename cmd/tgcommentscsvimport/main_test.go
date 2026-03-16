package main

import (
	"testing"
	"time"
)

func TestParseCSVReactions(t *testing.T) {
	got := parseCSVReactions("{':zany_face:': 2, 'ReactionCustomEmoji(document_id=5314692789492002556)': 3}")

	if got["emoji::zany_face:"] != 2 {
		t.Fatalf("expected zany_face=2, got %#v", got)
	}
	if got["custom:ReactionCustomEmoji(document_id=5314692789492002556)"] != 3 {
		t.Fatalf("expected custom reaction=3, got %#v", got)
	}
}

func TestBuildCommentAuthor(t *testing.T) {
	if got := buildCommentAuthor(csvCommentRow{Sender: "Иларион"}); got != "Иларион" {
		t.Fatalf("expected sender name, got %#v", got)
	}
	if got := buildCommentAuthor(csvCommentRow{SenderID: "424169621"}); got != "Участник #424169621" {
		t.Fatalf("expected sender id fallback, got %#v", got)
	}
	if got := buildCommentAuthor(csvCommentRow{SenderID: "-1"}); got != nil {
		t.Fatalf("expected nil author for root sender, got %#v", got)
	}
}

func TestMatchDiscussionRoot(t *testing.T) {
	row := csvCommentRow{
		MessageID:    3174,
		Message:      "Все больше думаю о организации своей секты, думаю основа уже подготовлена..",
		IsThreadRoot: true,
	}
	post, ok := matchDiscussionRoot(row, []tgPostCandidate{
		{PostID: 10, PublishedAt: mustTS("2026-03-13T17:45:56Z"), NormalizedText: normalizeLookupText(row.Message)},
	})
	if !ok || post.PostID != 10 {
		t.Fatalf("expected root to match post 10, got ok=%v post=%#v", ok, post)
	}
}

func mustTS(raw string) (outTime time.Time) {
	outTime, err := parseCSVDate(raw)
	if err != nil {
		panic(err)
	}
	return outTime
}
