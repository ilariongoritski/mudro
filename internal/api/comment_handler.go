package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/auth"
)

type createCommentRequest struct {
	Text            string `json:"text"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
}

type commentResponse struct {
	ID          int64     `json:"id"`
	PostID      int64     `json:"post_id"`
	AuthorName  string    `json:"author_name"`
	Text        string    `json:"text"`
	PublishedAt time.Time `json:"published_at"`
}

func (s *Server) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.ContextUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	postID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid post id"}`, http.StatusBadRequest)
		return
	}

	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		http.Error(w, `{"error":"text is required"}`, http.StatusBadRequest)
		return
	}

	// get username for display
	var authorName string
	err = s.pool.QueryRow(r.Context(),
		`select coalesce(display_name, username) from users where id = $1`, userID,
	).Scan(&authorName)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	now := time.Now()
	var commentID int64
	err = s.pool.QueryRow(r.Context(),
		`insert into post_comments (post_id, source, source_comment_id, author_name, text, published_at, parent_comment_id)
		 values ($1, 'local', 'local-' || nextval('post_comments_id_seq')::text, $2, $3, $4, $5)
		 returning id`,
		postID, authorName, req.Text, now, req.ParentCommentID,
	).Scan(&commentID)
	if err != nil {
		http.Error(w, `{"error":"failed to create comment"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(commentResponse{
		ID:          commentID,
		PostID:      postID,
		AuthorName:  authorName,
		Text:        req.Text,
		PublishedAt: now,
	})
}
