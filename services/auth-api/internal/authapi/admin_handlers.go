package authapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/goritskimihail/mudro/internal/auth"
)

type AdminHandlers struct {
	authSvc *auth.Service
}

func NewAdminHandlers(authSvc *auth.Service) *AdminHandlers {
	return &AdminHandlers{authSvc: authSvc}
}

type userListItem struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Email    *string `json:"email"`
	Role     string  `json:"role"`
}

// HandleGetUsers returns a list of all users.
func (h *AdminHandlers) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.authSvc == nil {
		http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
		return
	}
	users, err := h.authSvc.ListUsers(r.Context())
	if err != nil {
		log.Printf("admin list users: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	items := make([]userListItem, 0, len(users))
	for _, u := range users {
		items = append(items, userListItem{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			Role:     u.Role,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"users":  items,
	})
}

// HandleGetStats returns basic system stats for admin.
func (h *AdminHandlers) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.authSvc == nil {
		http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
		return
	}
	totalUsers, err := h.authSvc.CountUsers(r.Context())
	if err != nil {
		log.Printf("admin stats total_users: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	activeSubscriptions, err := h.authSvc.CountActiveSubscriptions(r.Context())
	if err != nil {
		log.Printf("admin stats active_subscriptions: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"total_users":          totalUsers,
		"active_subscriptions": activeSubscriptions,
	})
}
