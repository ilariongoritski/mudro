package profile

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin" // assume gin or use net/http; keep simple
)

// Handler for profile endpoints. Use with existing router.
type Handler struct {
	svc *Service
}

// NewHandler creates handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers profile routes on existing router.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/api/profile/:id", h.GetProfile)
	r.PUT("/api/profile/me", h.UpdateProfile)
	r.POST("/api/profile/avatar", h.UploadAvatar)
	r.GET("/api/profile/me/casino", h.GetCasinoStats)
	r.GET("/api/profile/:id/activities", h.GetActivities)
	r.POST("/api/profile/:id/message", h.StartMessage)
}

// GetProfile public view
func (h *Handler) GetProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	p, err := h.svc.GetProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// UpdateProfile (me only - add auth check in middleware)
func (h *Handler) UpdateProfile(c *gin.Context) {
	// TODO: get userID from JWT/context
	userID := int64(1) // placeholder
	var req struct {
		DisplayName string            `json:"display_name"`
		Username    string            `json:"username"`
		Email       *string           `json:"email"`
		Age         *int              `json:"age"`
		Bio         *string           `json:"bio"`
		SocialLinks map[string]string `json:"social_links"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// simple update query (expand in real impl)
	_, _ = h.svc.db.ExecContext(c.Request.Context(), `
		UPDATE users SET display_name=$1, username=$2, email=$3, age=$4, bio=$5, social_links=$6, updated_at=now()
		WHERE id=$7
	`, req.DisplayName, req.Username, req.Email, req.Age, req.Bio, req.SocialLinks, userID)

	// recalc completion & rating
	p, _ := h.svc.GetProfile(c.Request.Context(), userID)
	comp := h.svc.CalculateCompletion(p)
	_ = h.svc.UpdateRating(c.Request.Context(), userID, 20, 30) // base 20 + activity example

	c.JSON(http.StatusOK, gin.H{"completion": comp})
}

// UploadAvatar - simple placeholder (real: save to VPS, validate)
func (h *Handler) UploadAvatar(c *gin.Context) {
	// TODO: multipart, save to /var/avatars/{id}.jpg, update avatar_url
	c.JSON(http.StatusOK, gin.H{"avatar_url": "/avatars/1.jpg"})
}

// GetCasinoStats - join casino tables
func (h *Handler) GetCasinoStats(c *gin.Context) {
	// TODO: real query with casino_accounts + casino_rounds
	c.JSON(http.StatusOK, gin.H{
		"balance":     "1000.00",
		"games_count": 42,
		"max_win":     "250.00",
	})
}

// GetActivities
func (h *Handler) GetActivities(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	// simple select from user_activities
	rows, _ := h.svc.db.QueryContext(c.Request.Context(), `
		SELECT id, type, ref_id, metadata, created_at FROM user_activities 
		WHERE user_id = $1 ORDER BY created_at DESC LIMIT 20
	`, id)
	defer rows.Close()
	var acts []Activity
	for rows.Next() {
		var a Activity
		rows.Scan(&a.ID, &a.Type, &a.RefID, &a.Metadata, &a.CreatedAt)
		acts = append(acts, a)
	}
	c.JSON(http.StatusOK, acts)
}

// StartMessage - link to chat
func (h *Handler) StartMessage(c *gin.Context) {
	targetID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"chat_id": "new_or_existing", "target": targetID})
}
