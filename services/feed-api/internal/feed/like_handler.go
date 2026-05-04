package feed

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/goritskimihail/mudro/internal/auth"
	"github.com/goritskimihail/mudro/services/feed-api/internal/feed/contracts"
)

func (s *Server) handleToggleLike(w http.ResponseWriter, r *http.Request) {
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

	liked, count, err := s.postsSvc.ToggleLike(r.Context(), postID, userID)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(contracts.LikeResponse{Liked: liked, LikesCount: count})
}
