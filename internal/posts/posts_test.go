package posts

import (
	"testing"
	"time"

	"github.com/goritskimihail/mudro/pkg/models"
)

func TestParseMediaItems(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantLen  int
		wantItem func(*testing.T, []models.MediaItem)
	}{
		{
			name:    "empty input",
			input:   []byte(""),
			wantLen: 0,
		},
		{
			name:    "null input",
			input:   []byte("null"),
			wantLen: 0,
		},
		{
			name:    "empty array",
			input:   []byte("[]"),
			wantLen: 0,
		},
		{
			name: "single image item",
			input: []byte(`[{"kind": "image", "url": "https://example.com/img.jpg", "title": "Test Image", "width": 800, "height": 600, "position": 1}]`),
			wantLen: 1,
			wantItem: func(t *testing.T, items []models.MediaItem) {
				if items[0].Kind != "image" {
					t.Errorf("kind = %q, want \"image\"", items[0].Kind)
				}
				if items[0].URL != "https://example.com/img.jpg" {
					t.Errorf("url = %q, want %q", items[0].URL, "https://example.com/img.jpg")
				}
				if items[0].Title != "Test Image" {
					t.Errorf("title = %q, want \"Test Image\"", items[0].Title)
				}
				if items[0].Width != 800 {
					t.Errorf("width = %d, want 800", items[0].Width)
				}
				if items[0].Height != 600 {
					t.Errorf("height = %d, want 600", items[0].Height)
				}
				if items[0].Position != 1 {
					t.Errorf("position = %d, want 1", items[0].Position)
				}
			},
		},
		{
			name: "multiple items with mixed fields",
			input: []byte(`[
				{"kind": "video", "url": "https://example.com/vid.mp4", "position": 1},
				{"kind": "image", "preview_url": "https://example.com/thumb.jpg", "title": "Thumb", "position": 2}
			]`),
			wantLen: 2,
			wantItem: func(t *testing.T, items []models.MediaItem) {
				if items[0].Kind != "video" || items[1].Kind != "image" {
					t.Errorf("kinds = %q, %q, want video, image", items[0].Kind, items[1].Kind)
				}
				if items[0].URL != "https://example.com/vid.mp4" {
					t.Errorf("item[0].url = %q", items[0].URL)
				}
				if items[1].PreviewURL != "https://example.com/thumb.jpg" {
					t.Errorf("item[1].preview_url = %q", items[1].PreviewURL)
				}
			},
		},
		{
			name: "position auto-assignment when missing",
			input: []byte(`[
				{"kind": "image", "url": "a.jpg"},
				{"kind": "image", "url": "b.jpg"}
			]`),
			wantLen: 2,
			wantItem: func(t *testing.T, items []models.MediaItem) {
				if items[0].Position != 1 || items[1].Position != 2 {
					t.Errorf("positions = %d, %d, want 1, 2", items[0].Position, items[1].Position)
				}
			},
		},
		{
			name: "explicit positions preserved",
			input: []byte(`[
				{"kind": "image", "url": "a.jpg", "position": 5},
				{"kind": "image", "url": "b.jpg", "position": 3}
			]`),
			wantLen: 2,
			wantItem: func(t *testing.T, items []models.MediaItem) {
				if items[0].Position != 5 || items[1].Position != 3 {
					t.Errorf("positions = %d, %d, want 5, 3", items[0].Position, items[1].Position)
				}
			},
		},
		{
			name: "duplicate positions resolved",
			input: []byte(`[
				{"kind": "image", "url": "a.jpg", "position": 1},
				{"kind": "image", "url": "b.jpg", "position": 1},
				{"kind": "image", "url": "c.jpg", "position": 1}
			]`),
			wantLen: 3,
			wantItem: func(t *testing.T, items []models.MediaItem) {
				// Should auto-assign unique positions: 1, 2, 3
				positions := map[int]bool{}
				for _, item := range items {
					if positions[item.Position] {
						t.Errorf("duplicate position %d", item.Position)
					}
					positions[item.Position] = true
				}
				if len(positions) != 3 {
					t.Errorf("expected 3 unique positions, got %d", len(positions))
				}
			},
		},
		{
			name: "case-insensitive kind normalization",
			input: []byte(`[{"kind": "IMAGE", "url": "a.jpg"}, {"kind": "Video", "url": "b.mp4"}]`),
			wantLen: 2,
			wantItem: func(t *testing.T, items []models.MediaItem) {
				if items[0].Kind != "image" || items[1].Kind != "video" {
					t.Errorf("kinds = %q, %q, want image, video", items[0].Kind, items[1].Kind)
				}
			},
		},
		{
			name: "default kind to image when empty",
			input: []byte(`[{"url": "a.jpg"}]`),
			wantLen: 1,
			wantItem: func(t *testing.T, items []models.MediaItem) {
				if items[0].Kind != "image" {
					t.Errorf("kind = %q, want image (default)", items[0].Kind)
				}
			},
		},
		{
			name: "invalid json returns empty",
			input: []byte(`not valid json`),
			wantLen: 0,
		},
		{
			name: "skips empty rows",
			input: []byte(`[{}, {"kind": "image", "url": "a.jpg"}, {}]`),
			wantLen: 1,
			wantItem: func(t *testing.T, items []models.MediaItem) {
				if items[0].URL != "a.jpg" {
					t.Errorf("url = %q, want a.jpg", items[0].URL)
				}
			},
		},
		{
			name: "title fallback from extra.file_name",
			input: []byte(`[{"kind": "image", "url": "a.jpg", "extra": {"file_name": "my_photo.jpg"}}]`),
			wantLen: 1,
			wantItem: func(t *testing.T, items []models.MediaItem) {
				if items[0].Title != "my_photo.jpg" {
					t.Errorf("title = %q, want my_photo.jpg", items[0].Title)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			items := ParseMediaItems(tc.input)
			if len(items) != tc.wantLen {
				t.Errorf("ParseMediaItems() len = %d, want %d", len(items), tc.wantLen)
			}
			if tc.wantItem != nil {
				tc.wantItem(t, items)
			}
		})
	}
}

func TestBuildPostsVisibilityWhere(t *testing.T) {
	// Test with tgVisiblePostIDs set
	sWithTG := &Service{tgVisiblePostIDs: []string{"1001", "1002"}}
	// Test without tgVisiblePostIDs
	sWithoutTG := &Service{tgVisiblePostIDs: nil}

	tests := []struct {
		name         string
		s            *Service
		source       string
		query        string
		wantWhere    string
		wantArgsLen  int
	}{
		{
			name:         "no filters, no tg visible posts",
			s:            sWithoutTG,
			source:       "",
			query:        "",
			wantWhere:    "",
			wantArgsLen:  0,
		},
		{
			name:         "source filter only, no tg visible posts",
			s:            sWithoutTG,
			source:       "tg",
			query:        "",
			wantWhere:    " where source = $1",
			wantArgsLen:  1,
		},
		{
			name:         "query filter only, no tg visible posts",
			s:            sWithoutTG,
			source:       "",
			query:        "test",
			wantWhere:    " where LOWER(text) LIKE $1",
			wantArgsLen:  1,
		},
		{
			name:         "source and query filters, no tg visible posts",
			s:            sWithoutTG,
			source:       "vk",
			query:        "hello",
			wantWhere:    " where source = $1 and LOWER(text) LIKE $2",
			wantArgsLen:  2,
		},
		{
			name:         "no filters, with tg visible posts",
			s:            sWithTG,
			source:       "",
			query:        "",
			wantWhere:    " where (source <> 'tg' or source_post_id = any($1))",
			wantArgsLen:  1,
		},
		{
			name:         "source filter only, with tg visible posts",
			s:            sWithTG,
			source:       "tg",
			query:        "",
			wantWhere:    " where source = $1 and source_post_id = any($2)",
			wantArgsLen:  2,
		},
		{
			name:         "query filter only, with tg visible posts",
			s:            sWithTG,
			source:       "",
			query:        "test",
			wantWhere:    " where LOWER(text) LIKE $1 and (source <> 'tg' or source_post_id = any($2))",
			wantArgsLen:  2,
		},
		{
			name:         "source and query filters, with tg visible posts",
			s:            sWithTG,
			source:       "vk",
			query:        "hello",
			wantWhere:    " where source = $1 and LOWER(text) LIKE $2",
			wantArgsLen:  2,
		},
		{
			name:         "tg source with visible post IDs",
			s:            sWithTG,
			source:       "tg",
			query:        "",
			wantWhere:    " where source = $1 and source_post_id = any($2)",
			wantArgsLen:  2,
		},
		{
			name:         "tg source with query and visible post IDs",
			s:            sWithTG,
			source:       "tg",
			query:        "search",
			wantWhere:    " where source = $1 and LOWER(text) LIKE $2 and source_post_id = any($3)",
			wantArgsLen:  3,
		},
		{
			name:         "empty source with query and visible post IDs",
			s:            sWithTG,
			source:       "",
			query:        "search",
			wantWhere:    " where LOWER(text) LIKE $1 and (source <> 'tg' or source_post_id = any($2))",
			wantArgsLen:  2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			whereSQL, args := tc.s.buildPostsVisibilityWhere(tc.source, tc.query, nil)
			if whereSQL != tc.wantWhere {
				t.Errorf("whereSQL = %q, want %q", whereSQL, tc.wantWhere)
			}
			if len(args) != tc.wantArgsLen {
				t.Errorf("args len = %d, want %d", len(args), tc.wantArgsLen)
			}
		})
	}
}

func TestBuildPostsQuery(t *testing.T) {
	s := &Service{tgVisiblePostIDs: []string{"1001"}}

	beforeTS := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	beforeID := int64(100)
	limit := 20
	page := 2

	tests := []struct {
		name          string
		source        string
		query         string
		page          *int
		beforeTS      *time.Time
		beforeID      *int64
		sortOrder     SortOrder
		wantCursor    bool
		wantContains  []string
		wantNotContain []string
	}{
		{
			name:          "first page, no cursor, desc order",
			source:        "tg",
			query:         "",
			page:          nil,
			beforeTS:      nil,
			beforeID:      nil,
			sortOrder:     SortDesc,
			wantCursor:    false, // useCursor only when both beforeTS and beforeID provided
			wantContains:  []string{"order by published_at desc, id desc limit", "source_post_id = any"},
			wantNotContain: []string{"offset"},
		},
		{
			name:          "page-based pagination",
			source:        "vk",
			query:         "test",
			page:          &page,
			beforeTS:      nil,
			beforeID:      nil,
			sortOrder:     SortDesc,
			wantCursor:    false,
			wantContains:  []string{"offset", "limit", "LOWER(text) LIKE"},
			wantNotContain: []string{"(published_at, id)"},
		},
		{
			name:          "keyset pagination with cursor",
			source:        "",
			query:         "",
			page:          nil,
			beforeTS:      &beforeTS,
			beforeID:      &beforeID,
			sortOrder:     SortDesc,
			wantCursor:    true,
			wantContains:  []string{"(published_at, id) <", "limit $"},
			wantNotContain: []string{"offset"},
		},
		{
			name:          "ascending order",
			source:        "tg",
			query:         "",
			page:          nil,
			beforeTS:      &beforeTS,
			beforeID:      &beforeID,
			sortOrder:     SortAsc,
			wantCursor:    true,
			wantContains:  []string{"order by published_at asc, id asc", "(published_at, id) >"},
			wantNotContain: []string{"desc"},
		},
		{
			name:          "empty source and query, with tg visible posts",
			source:        "",
			query:         "",
			page:          nil,
			beforeTS:      nil,
			beforeID:      nil,
			sortOrder:     SortDesc,
			wantCursor:    false,
			wantContains:  []string{"select id, source", "from posts", "(source <> 'tg' or source_post_id = any"},
			wantNotContain: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			comparator := "<"
			if tc.sortOrder == SortAsc {
				comparator = ">"
			}
			q, args, useCursor := s.buildPostsQuery(tc.source, tc.query, tc.page, tc.beforeTS, tc.beforeID, limit, string(tc.sortOrder), comparator)
			if useCursor != tc.wantCursor {
				t.Errorf("useCursor = %v, want %v", useCursor, tc.wantCursor)
			}
			for _, want := range tc.wantContains {
				if !contains(q, want) {
					t.Errorf("query missing %q:\n%s", want, q)
				}
			}
			for _, notWant := range tc.wantNotContain {
				if contains(q, notWant) {
					t.Errorf("query unexpectedly contains %q:\n%s", notWant, q)
				}
			}
			if len(args) == 0 && tc.wantCursor && (tc.beforeTS != nil || tc.page != nil) {
				t.Errorf("expected args for pagination, got none")
			}
		})
	}
}

func TestBuildPostsQuery_InvalidOrder(t *testing.T) {
	s := &Service{}
	q, args, useCursor := s.buildPostsQuery("", "", nil, nil, nil, 10, "invalid", "<")
	if q != "" || args != nil || useCursor {
		t.Errorf("invalid order should return empty: q=%q, args=%v, useCursor=%v", q, args, useCursor)
	}
}

func TestBuildPostsQuery_InvalidComparator(t *testing.T) {
	s := &Service{}
	q, args, useCursor := s.buildPostsQuery("", "", nil, nil, nil, 10, "desc", "invalid")
	if q != "" || args != nil || useCursor {
		t.Errorf("invalid comparator should return empty: q=%q, args=%v, useCursor=%v", q, args, useCursor)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestSortOrderConstants verifies SortOrder constants match models
func TestSortOrderConstants(t *testing.T) {
	if SortDesc != models.SortDesc {
		t.Errorf("SortDesc = %q, want %q", SortDesc, models.SortDesc)
	}
	if SortAsc != models.SortAsc {
		t.Errorf("SortAsc = %q, want %q", SortAsc, models.SortAsc)
	}
}

func TestCursorTypeAlias(t *testing.T) {
	var c Cursor = models.Cursor{BeforeTS: time.Now(), BeforeID: 123}
	if c.BeforeID != 123 {
		t.Errorf("Cursor type alias failed")
	}
}

func TestPostTypeAlias(t *testing.T) {
	var p Post = models.Post{ID: 1, Source: "tg"}
	if p.ID != 1 || p.Source != "tg" {
		t.Errorf("Post type alias failed")
	}
}

func TestCommentTypeAlias(t *testing.T) {
	var c Comment = models.Comment{SourceCommentID: "123", AuthorName: "Test"}
	if c.SourceCommentID != "123" || c.AuthorName != "Test" {
		t.Errorf("Comment type alias failed")
	}
}

func TestMediaItemTypeAlias(t *testing.T) {
	var m MediaItem = models.MediaItem{Kind: "image", URL: "test.jpg"}
	if m.Kind != "image" || m.URL != "test.jpg" {
		t.Errorf("MediaItem type alias failed")
	}
}

func TestSourceStatTypeAlias(t *testing.T) {
	var s SourceStat = models.SourceStat{Source: "tg", Posts: 42}
	if s.Source != "tg" || s.Posts != 42 {
		t.Errorf("SourceStat type alias failed")
	}
}

func TestReactionTypeAlias(t *testing.T) {
	var r Reaction = models.Reaction{Label: "👍", Count: 10, Raw: "raw"}
	if r.Label != "👍" || r.Count != 10 || r.Raw != "raw" {
		t.Errorf("Reaction type alias failed")
	}
}