package models

import (
	"encoding/json"
	"time"
)

type SortOrder string

const (
	SortDesc SortOrder = "desc"
	SortAsc  SortOrder = "asc"
)

type Cursor struct {
	BeforeTS time.Time `json:"before_ts"`
	BeforeID int64     `json:"before_id"`
}

type Post struct {
	ID            int64           `json:"id"`
	Source        string          `json:"source"`
	SourcePostID  string          `json:"source_post_id"`
	PublishedAt   time.Time       `json:"published_at"`
	Text          *string         `json:"text,omitempty"`
	Media         json.RawMessage `json:"media,omitempty"`
	LikesCount    int             `json:"likes_count"`
	ViewsCount    *int            `json:"views_count,omitempty"`
	CommentsCount *int            `json:"comments_count,omitempty"`
	Reactions     map[string]int  `json:"reactions,omitempty"`
	Comments      []Comment       `json:"comments,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type Comment struct {
	SourceCommentID string          `json:"source_comment_id"`
	ParentCommentID string          `json:"parent_comment_id,omitempty"`
	AuthorName      string          `json:"author_name,omitempty"`
	PublishedAt     string          `json:"published_at"`
	Text            string          `json:"text,omitempty"`
	Reactions       map[string]int  `json:"reactions,omitempty"`
	Media           json.RawMessage `json:"media,omitempty"`
}

type MediaItem struct {
	Kind       string `json:"kind"`
	URL        string `json:"url,omitempty"`
	PreviewURL string `json:"preview_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Position   int    `json:"position,omitempty"`
}

type SourceStat struct {
	Source string `json:"source"`
	Posts  int64  `json:"posts"`
}

type Reaction struct {
	Label string `json:"label"`
	Count int    `json:"count"`
	Raw   string `json:"raw"`
}
