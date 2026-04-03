package feed

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/auth"
	internalcasino "github.com/goritskimihail/mudro/internal/casino"
)

type authSessionUser struct {
	ID              int64   `json:"id"`
	Username        string  `json:"username"`
	Email           *string `json:"email,omitempty"`
	Role            string  `json:"role"`
	IsPremium       bool    `json:"isPremium"`
	IsPremiumLegacy bool    `json:"is_premium"`
}

func buildAuthSessionUser(user *auth.User) authSessionUser {
	return authSessionUser{
		ID:              user.ID,
		Username:        user.Username,
		Email:           user.Email,
		Role:            user.Role,
		IsPremium:       user.IsPremium,
		IsPremiumLegacy: user.IsPremium,
	}
}

func writeAuthSession(w http.ResponseWriter, status int, user *auth.User, token string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"token": token,
		"user":  buildAuthSessionUser(user),
	})
}

func writeAuthUser(w http.ResponseWriter, user *auth.User) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(buildAuthSessionUser(user))
}

func extractBearerToken(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return ""
	}
	return strings.TrimSpace(header[7:])
}

func subjectFromClaims(claims map[string]any) (int64, error) {
	switch v := claims["sub"].(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case json.Number:
		return v.Int64()
	case string:
		return strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	default:
		return 0, auth.ErrInvalidToken
	}
}

func (s *Server) authenticatedUserFromRequest(r *http.Request) (*auth.User, error) {
	token := extractBearerToken(r)
	if token == "" {
		return nil, auth.ErrInvalidToken
	}
	claims, err := s.authSvc.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	userID, err := subjectFromClaims(claims)
	if err != nil {
		return nil, err
	}
	return s.authSvc.GetUserByID(r.Context(), userID)
}

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

	writeAuthSession(w, http.StatusOK, user, token)
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

	writeAuthSession(w, http.StatusCreated, user, token)
}

// handleAuthMe GET /api/auth/me
func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := s.authenticatedUserFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	writeAuthUser(w, user)
}

// handleAuthRefresh POST /api/auth/refresh
func (s *Server) handleAuthRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := s.authenticatedUserFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token, err := s.authSvc.IssueToken(user)
	if err != nil {
		log.Printf("auth: refresh token failed for %d: %v", user.ID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeAuthSession(w, http.StatusOK, user, token)
}

// handleAuthTelegram POST /api/auth/telegram
func (s *Server) handleAuthTelegram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		InitData string `json:"initData"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	initData := strings.TrimSpace(body.InitData)
	if initData == "" {
		http.Error(w, "initData is required", http.StatusBadRequest)
		return
	}

	botToken := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if botToken == "" {
		botToken = strings.TrimSpace(os.Getenv("CASINO_BOT_TOKEN"))
	}
	if botToken == "" {
		http.Error(w, "telegram auth is not configured", http.StatusServiceUnavailable)
		return
	}

	tgAuth, err := internalcasino.ValidateInitData(botToken, initData)
	if err != nil {
		http.Error(w, "invalid telegram initData", http.StatusUnauthorized)
		return
	}
	if tgAuth.AuthDate > 0 && time.Since(time.Unix(tgAuth.AuthDate, 0)) > 24*time.Hour {
		http.Error(w, "expired telegram session", http.StatusUnauthorized)
		return
	}

	user, err := s.authSvc.FindOrCreateTelegramUser(r.Context(), tgAuth.TelegramID, tgAuth.Username)
	if err != nil {
		log.Printf("auth: telegram bootstrap failed for %d: %v", tgAuth.TelegramID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	token, err := s.authSvc.IssueToken(user)
	if err != nil {
		log.Printf("auth: telegram token failed for %d: %v", user.ID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeAuthSession(w, http.StatusOK, user, token)
}
