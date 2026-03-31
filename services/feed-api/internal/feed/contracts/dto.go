package contracts

import "time"

type CreateCommentRequest struct {
	Text            string `json:"text"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
}

type CommentResponse struct {
	ID          int64     `json:"id"`
	PostID      int64     `json:"post_id"`
	AuthorName  string    `json:"author_name"`
	Text        string    `json:"text"`
	PublishedAt time.Time `json:"published_at"`
}

type LikeResponse struct {
	Liked      bool `json:"liked"`
	LikesCount int  `json:"likes_count"`
}
