package api

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
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// HandleGetUsers returns a list of all users.
func (h *AdminHandlers) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.authSvc.ListUsers(r.Context())
	if err != nil {
		log.Printf("admin list users: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	items := make([]userListItem, 0, len(users))
	for _, u := range users {
		items = append(items, userListItem{
			ID:    u.ID,
			Email: u.Email,
			Role:  u.Role,
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
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"total_users": 0, // Placeholder
		"active_subscriptions": 0, // Placeholder
	})
}
