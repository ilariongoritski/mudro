package app

import (
	"testing"
	"time"
)

func TestBuildPlanMatchesUniqueDiscussionRootPost(t *testing.T) {
	base := time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC)
	posts := []postState{
		{ID: 1, SourcePostID: "1200", PublishedAt: base, Text: "Hello world"},
		{ID: 2, SourcePostID: "1305", PublishedAt: base.Add(3 * time.Second), Text: "Hello world", ActualComments: 4},
	}
	roots := map[int64]rootRow{
		1305: {MessageID: 1305, Date: base.Add(3 * time.Second), Text: "hello world"},
	}

	plan := buildPlan(posts, roots, 1299)
	if plan.MatchedGroups != 1 {
		t.Fatalf("MatchedGroups = %d, want 1", plan.MatchedGroups)
	}
	if plan.RowsToRemove != 1 {
		t.Fatalf("RowsToRemove = %d, want 1", plan.RowsToRemove)
	}
	if plan.MovedComments != 4 {
		t.Fatalf("MovedComments = %d, want 4", plan.MovedComments)
	}
	if len(plan.Merges) != 1 {
		t.Fatalf("len(Merges) = %d, want 1", len(plan.Merges))
	}
	if plan.Merges[0].CanonicalID != 1 || plan.Merges[0].DuplicateID != 2 {
		t.Fatalf("merge = %+v, want canonical=1 duplicate=2", plan.Merges[0])
	}
}

func TestBuildPlanSkipsAmbiguousEmptyText(t *testing.T) {
	base := time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC)
	posts := []postState{
		{ID: 1, SourcePostID: "1200", PublishedAt: base, Text: ""},
		{ID: 2, SourcePostID: "1201", PublishedAt: base.Add(10 * time.Second), Text: ""},
		{ID: 3, SourcePostID: "1305", PublishedAt: base.Add(5 * time.Second), Text: "", ActualComments: 2},
	}
	roots := map[int64]rootRow{
		1305: {MessageID: 1305, Date: base.Add(5 * time.Second), Text: ""},
	}

	plan := buildPlan(posts, roots, 1299)
	if plan.MatchedGroups != 0 {
		t.Fatalf("MatchedGroups = %d, want 0", plan.MatchedGroups)
	}
	if plan.SkippedAmbiguous != 1 {
		t.Fatalf("SkippedAmbiguous = %d, want 1", plan.SkippedAmbiguous)
	}
}

func TestBuildPlanFallsBackToCommentsCountComplement(t *testing.T) {
	base := time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC)
	posts := []postState{
		{
			ID:            1,
			SourcePostID:  "1183",
			PublishedAt:   base,
			Text:          "",
			CommentsCount: intPtr(8),
		},
		{
			ID:           3,
			SourcePostID: "1184",
			PublishedAt:  base,
			Text:         "",
		},
		{
			ID:             2,
			SourcePostID:   "2898",
			PublishedAt:    base.Add(4 * time.Second),
			Text:           "",
			ActualComments: 8,
		},
	}
	roots := map[int64]rootRow{
		2898: {MessageID: 2898, Date: base.Add(4 * time.Second), Text: ""},
	}

	plan := buildPlan(posts, roots, 1299)
	if plan.MatchedGroups != 1 {
		t.Fatalf("MatchedGroups = %d, want 1", plan.MatchedGroups)
	}
	if plan.Merges[0].Reason != "comments-count-complement" {
		t.Fatalf("reason = %q, want comments-count-complement", plan.Merges[0].Reason)
	}
}

func intPtr(v int) *int {
	return &v
}
