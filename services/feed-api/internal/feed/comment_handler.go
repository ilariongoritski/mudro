package feed

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/goritskimihail/mudro/internal/auth"
	"github.com/goritskimihail/mudro/services/feed-api/internal/feed/contracts"
)


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

	var req contracts.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		http.Error(w, `{"error":"text is required"}`, http.StatusBadRequest)
		return
	}

	user, err := s.authSvc.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	authorName := user.Username
	if user.TelegramName != nil && *user.TelegramName != "" {
		authorName = *user.TelegramName
	}

	commentID, now, err := s.postsSvc.AddComment(r.Context(), postID, authorName, req.Text, req.ParentCommentID)
	if err != nil {
		http.Error(w, `{"error":"failed to create comment"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(contracts.CommentResponse{
		ID:          commentID,
		PostID:      postID,
		AuthorName:  authorName,
		Text:        req.Text,
		PublishedAt: now,
	})
}
