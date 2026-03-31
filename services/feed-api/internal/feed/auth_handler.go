package feed

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// handleAuthLogin POST /api/auth/login
func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Login    string `json:"login"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	login := strings.TrimSpace(body.Login)
	if login == "" {
		login = strings.TrimSpace(body.Email)
	}
	if login == "" || strings.TrimSpace(body.Password) == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}

	user, token, err := s.authSvc.Login(r.Context(), login, body.Password)
	if err != nil {
		log.Printf("auth: login failed for %q: %v", login, err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"token": token,
		"user": map[string]any{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"role":      user.Role,
			"isPremium": user.IsPremium,
		},
	})
}

// handleAuthRegister POST /api/auth/register
func (s *Server) handleAuthRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Username string `json:"username"`
		Login    string `json:"login"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(body.Username)
	if username == "" {
		username = strings.TrimSpace(body.Login)
	}
	email := strings.TrimSpace(body.Email)
	password := strings.TrimSpace(body.Password)

	if username == "" || password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	user, err := s.authSvc.Register(r.Context(), username, email, password)
	if err != nil {
		log.Printf("auth: register failed for %q: %v", username, err)
		http.Error(w, "registration failed", http.StatusConflict)
		return
	}

	token, err := s.authSvc.IssueToken(user)
	if err != nil {
		http.Error(w, "token issue failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"token": token,
		"user": map[string]any{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"role":      user.Role,
			"isPremium": user.IsPremium,
		},
	})
}

// handleAuthMe GET /api/auth/me
func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimSpace(header[7:])

	claims, err := s.authSvc.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var userID int64
	switch v := claims["sub"].(type) {
	case float64:
		userID = int64(v)
	default:
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	user, err := s.authSvc.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"role":      user.Role,
		"isPremium": user.IsPremium,
	})
}
