package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/goritskimihail/mudro/internal/auth"
)

type likeResponse struct {
	Liked      bool `json:"liked"`
	LikesCount int  `json:"likes_count"`
}

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

	// try insert; if already liked, delete instead
	var liked bool
	tag, err := s.pool.Exec(r.Context(),
		`insert into post_user_likes (post_id, user_id) values ($1, $2) on conflict do nothing`,
		postID, userID,
	)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	if tag.RowsAffected() == 0 {
		// already liked — remove
		_, err = s.pool.Exec(r.Context(),
			`delete from post_user_likes where post_id = $1 and user_id = $2`,
			postID, userID,
		)
		if err != nil {
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}
		liked = false
	} else {
		liked = true
	}

	var count int
	_ = s.pool.QueryRow(r.Context(),
		`select likes_count from posts where id = $1`, postID,
	).Scan(&count)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(likeResponse{Liked: liked, LikesCount: count})
}
