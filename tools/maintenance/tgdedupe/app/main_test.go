package app

import (
	"testing"
	"time"
)

func TestClassifyGroupSafeForNonEmptyPair(t *testing.T) {
	group := []postState{
		{ID: 1, Text: "hello", PublishedAt: time.Unix(100, 0)},
		{ID: 2, Text: "hello", PublishedAt: time.Unix(103, 0)},
	}

	reason, ok := classifyGroup(group)
	if !ok || reason != "same-text-near-time" {
		t.Fatalf("classifyGroup() = (%q, %v), want same-text-near-time, true", reason, ok)
	}
}

func TestClassifyGroupSkipsIdenticalEmptyPair(t *testing.T) {
	group := []postState{
		{ID: 1, PublishedAt: time.Unix(100, 0)},
		{ID: 2, PublishedAt: time.Unix(103, 0)},
	}

	_, ok := classifyGroup(group)
	if ok {
		t.Fatal("expected identical empty pair to be skipped")
	}
}

func TestCanonicalLessPrefersVisible(t *testing.T) {
	left := postState{ID: 1, Visible: true, PublishedAt: time.Unix(100, 0)}
	right := postState{ID: 2, Visible: false, ActualComments: 10, PublishedAt: time.Unix(101, 0)}

	if !canonicalLess(left, right) {
		t.Fatal("expected visible post to win canonical rank")
	}
}
