package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// mockRepo implements UserRepository for testing.
type mockRepo struct {
	users           map[int64]*User
	usersByLogin    map[string]*User
	usersByTG       map[int64]*User
	subs            map[int64]bool
	createErr       error
	findByLoginErr  error
	findByIDErr     error
	findByTGErr     error
	countResult     int64
	countSubsResult int64
	subErr          error
	hasSubErr       error
	updateTGNameErr error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users:        make(map[int64]*User),
		usersByLogin: make(map[string]*User),
		usersByTG:    make(map[int64]*User),
		subs:         make(map[int64]bool),
	}
}

func (m *mockRepo) FindByLogin(ctx context.Context, login string) (*User, error) {
	if m.findByLoginErr != nil {
		return nil, m.findByLoginErr
	}
	if u, ok := m.usersByLogin[login]; ok {
		return u, nil
	}
	return nil, ErrInvalidCredentials
}

func (m *mockRepo) FindByID(ctx context.Context, id int64) (*User, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, ErrNoSession
}

func (m *mockRepo) FindByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	if m.findByTGErr != nil {
		return nil, m.findByTGErr
	}
	if u, ok := m.usersByTG[telegramID]; ok {
		return u, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) Create(ctx context.Context, username, email, passwordHash string) (*User, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	id := int64(len(m.users) + 1)
	u := &User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         "user",
		CreatedAt:    time.Now(),
	}
	if email != "" {
		u.Email = &email
	}
	m.users[id] = u
	m.usersByLogin[username] = u
	if email != "" {
		m.usersByLogin[email] = u
	}
	return u, nil
}

func (m *mockRepo) CreateFromTelegram(ctx context.Context, params TelegramUserParams) (*User, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if _, exists := m.usersByTG[params.TelegramID]; exists {
		return nil, ErrTelegramIDConflict
	}
	for _, u := range m.users {
		if u.Username == params.Username {
			return nil, ErrUserExists
		}
	}
	id := int64(len(m.users) + 1)
	u := &User{
		ID:           id,
		Username:     params.Username,
		PasswordHash: params.PasswordHash,
		Role:         "user",
		TelegramID:   &params.TelegramID,
		TelegramName: params.TelegramName,
		CreatedAt:    time.Now(),
	}
	m.users[id] = u
	m.usersByLogin[params.Username] = u
	m.usersByTG[params.TelegramID] = u
	return u, nil
}

func (m *mockRepo) UpdateTelegramName(ctx context.Context, id int64, name string) error {
	return m.updateTGNameErr
}

func (m *mockRepo) HasActiveSubscription(ctx context.Context, userID int64) (bool, error) {
	if m.hasSubErr != nil {
		return false, m.hasSubErr
	}
	return m.subs[userID], nil
}

func (m *mockRepo) ListAll(ctx context.Context) ([]User, error) {
	users := make([]User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, *u)
	}
	return users, nil
}

func (m *mockRepo) Count(ctx context.Context) (int64, error) {
	return m.countResult, nil
}

func (m *mockRepo) CountActiveSubscriptions(ctx context.Context) (int64, error) {
	return m.countSubsResult, nil
}

func (m *mockRepo) AddSubscription(ctx context.Context, userID int64, planID string, duration time.Duration) error {
	if m.subErr != nil {
		return m.subErr
	}
	m.subs[userID] = true
	return nil
}

func TestService_Register_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	user, err := svc.Register(context.Background(), "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "testuser" {
		t.Fatalf("expected username 'testuser', got %q", user.Username)
	}
	if user.PasswordHash == "" {
		t.Fatal("expected password hash to be set")
	}
}

func TestService_Register_DBError(t *testing.T) {
	repo := newMockRepo()
	repo.createErr = errors.New("db error")
	svc := NewService(repo, "test-secret")

	_, err := svc.Register(context.Background(), "testuser", "test@example.com", "password123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_Login_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	_, err := svc.Register(context.Background(), "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	user, token, err := svc.Login(context.Background(), "testuser", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "testuser" {
		t.Fatalf("expected username 'testuser', got %q", user.Username)
	}
	if token == "" {
		t.Fatal("expected token to be returned")
	}
}

func TestService_Login_InvalidCredentials(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	_, _, err := svc.Login(context.Background(), "nonexistent", "password123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_Login_WrongPassword(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	_, err := svc.Register(context.Background(), "testuser", "", "correctpassword")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, _, err = svc.Login(context.Background(), "testuser", "wrongpassword")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestService_IssueToken_NilUser(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	_, err := svc.IssueToken(nil)
	if err == nil {
		t.Fatal("expected error for nil user")
	}
}

func TestService_IssueToken_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	user := &User{ID: 42, Username: "testuser", Role: "admin"}
	token, err := svc.IssueToken(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
}

func TestService_ValidateToken_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	user := &User{ID: 42, Username: "testuser", Role: "user"}
	token, err := svc.IssueToken(user)
	if err != nil {
		t.Fatalf("issue failed: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	sub, ok := claims["sub"].(float64)
	if !ok || int64(sub) != 42 {
		t.Fatalf("expected sub 42, got %v", claims["sub"])
	}
	role, ok := claims["role"].(string)
	if !ok || role != "user" {
		t.Fatalf("expected role 'user', got %v", claims["role"])
	}
}

func TestService_ValidateToken_WrongSecret(t *testing.T) {
	repo := newMockRepo()
	svc1 := NewService(repo, "secret-one")
	svc2 := NewService(repo, "secret-two")

	user := &User{ID: 42, Role: "user"}
	token, err := svc1.IssueToken(user)
	if err != nil {
		t.Fatalf("issue failed: %v", err)
	}

	_, err = svc2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestService_ValidateToken_Tampered(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	user := &User{ID: 42, Role: "user"}
	token, err := svc.IssueToken(user)
	if err != nil {
		t.Fatalf("issue failed: %v", err)
	}

	_, err = svc.ValidateToken(token + "tampered")
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestService_GetUserByID(t *testing.T) {
	repo := newMockRepo()
	repo.users[1] = &User{ID: 1, Username: "testuser"}
	repo.usersByLogin["testuser"] = repo.users[1]
	repo.subs[1] = true
	svc := NewService(repo, "test-secret")

	user, err := svc.GetUserByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "testuser" {
		t.Fatalf("expected 'testuser', got %q", user.Username)
	}
	if !user.IsPremium {
		t.Fatal("expected IsPremium to be true")
	}
}

func TestService_ListUsers(t *testing.T) {
	repo := newMockRepo()
	repo.users[1] = &User{ID: 1, Username: "user1"}
	repo.users[2] = &User{ID: 2, Username: "user2"}
	svc := NewService(repo, "test-secret")

	users, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestService_CountUsers(t *testing.T) {
	repo := newMockRepo()
	repo.countResult = 42
	svc := NewService(repo, "test-secret")

	count, err := svc.CountUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 42 {
		t.Fatalf("expected 42, got %d", count)
	}
}

func TestService_Subscriptions(t *testing.T) {
	repo := newMockRepo()
	repo.countSubsResult = 5
	svc := NewService(repo, "test-secret")

	count, err := svc.CountActiveSubscriptions(context.Background())
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected 5, got %d", count)
	}

	err = svc.AddSubscription(context.Background(), 1, "plan1", 30*24*time.Hour)
	if err != nil {
		t.Fatalf("add sub failed: %v", err)
	}

	has, err := svc.HasActiveSubscription(context.Background(), 1)
	if err != nil {
		t.Fatalf("has sub failed: %v", err)
	}
	if !has {
		t.Fatal("expected active subscription")
	}
}

func TestService_FindOrCreateTelegramUser_NewUser(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	user, err := svc.FindOrCreateTelegramUser(context.Background(), 12345, "testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username == "" {
		t.Fatal("expected username to be set")
	}
	if user.TelegramID == nil || *user.TelegramID != 12345 {
		t.Fatalf("expected TelegramID 12345, got %v", user.TelegramID)
	}
}

func TestService_FindOrCreateTelegramUser_ExistingUser(t *testing.T) {
	repo := newMockRepo()
	tgID := int64(12345)
	name := "existing"
	repo.users[1] = &User{ID: 1, Username: "existing", TelegramID: &tgID, TelegramName: &name}
	repo.usersByTG[tgID] = repo.users[1]
	repo.subs[1] = true
	svc := NewService(repo, "test-secret")

	user, err := svc.FindOrCreateTelegramUser(context.Background(), 12345, "newname")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "existing" {
		t.Fatalf("expected 'existing', got %q", user.Username)
	}
	if !user.IsPremium {
		t.Fatal("expected IsPremium")
	}
}

func TestService_FindOrCreateTelegramUser_InvalidID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	_, err := svc.FindOrCreateTelegramUser(context.Background(), 0, "testuser")
	if err == nil {
		t.Fatal("expected error for invalid telegram ID")
	}
}

func TestService_SetTokenExpiry(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	svc.SetTokenExpiry(24 * time.Hour)
	user := &User{ID: 1, Role: "user"}
	token, err := svc.IssueToken(user)
	if err != nil {
		t.Fatalf("issue failed: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal("expected exp claim")
	}
	now := float64(time.Now().Unix())
	if exp < now+23*3600 || exp > now+25*3600 {
		t.Fatalf("expected expiry around 24h from now, got %v (now=%v)", exp, now)
	}
}

func TestService_SetTokenExpiry_Zero(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	svc.SetTokenExpiry(0)
	user := &User{ID: 1, Role: "user"}
	token, err := svc.IssueToken(user)
	if err != nil {
		t.Fatalf("issue failed: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal("expected exp claim")
	}
	now := float64(time.Now().Unix())
	// Default is 168h (7 days)
	if exp < now+167*3600 || exp > now+169*3600 {
		t.Fatalf("expected expiry around 168h from now, got %v (now=%v)", exp, now)
	}
}

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected hash to be non-empty")
	}
}

func TestCheckPassword(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if !CheckPassword(hash, "mypassword") {
		t.Fatal("expected password to match")
	}
	if CheckPassword(hash, "wrongpassword") {
		t.Fatal("expected password to not match")
	}
}

func TestGenerateToken(t *testing.T) {
	SetSecret("test-secret-12345")
	token, err := GenerateToken(42, "testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
}

func TestGenerateToken_NoSecret(t *testing.T) {
	jwtSecret = nil
	_, err := GenerateToken(42, "testuser")
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestParseToken(t *testing.T) {
	SetSecret("test-secret-12345")
	token, err := GenerateToken(42, "testuser")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if claims.UserID != 42 {
		t.Fatalf("expected UserID 42, got %d", claims.UserID)
	}
	if claims.Username != "testuser" {
		t.Fatalf("expected username 'testuser', got %q", claims.Username)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	SetSecret("secret-one")
	token, err := GenerateToken(42, "testuser")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	SetSecret("secret-two")
	_, err = ParseToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestParseToken_NoSecret(t *testing.T) {
	jwtSecret = nil
	_, err := ParseToken("some.token.here")
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestContextUserID(t *testing.T) {
	ctx := WithUserID(context.Background(), 42)
	id, ok := ContextUserID(ctx)
	if !ok {
		t.Fatal("expected ok")
	}
	if id != 42 {
		t.Fatalf("expected 42, got %d", id)
	}
}

func TestContextUserID_Missing(t *testing.T) {
	_, ok := ContextUserID(context.Background())
	if ok {
		t.Fatal("expected not ok")
	}
}

func TestNormalizeTelegramUsername(t *testing.T) {
	tests := []struct {
		input    string
		tgID     int64
		expected string
	}{
		{"TestUser_123", 1, "testuser_123"},
		{"", 42, "tg_42"},
		{"   ", 42, "tg_42"},
		{"Hello@World!", 1, "helloworld"},
		{"a_very_long_username_that_exceeds_32_chars", 1, "a_very_long_username_that_exceed"},
	}

	for _, tt := range tests {
		result := normalizeTelegramUsername(tt.input, tt.tgID)
		if result != tt.expected {
			t.Errorf("normalizeTelegramUsername(%q, %d) = %q, want %q", tt.input, tt.tgID, result, tt.expected)
		}
	}
}

func TestNullableTrimmed(t *testing.T) {
	result := nullableTrimmed("  hello  ")
	if result == nil || *result != "hello" {
		t.Fatalf("expected 'hello', got %v", result)
	}

	result = nullableTrimmed("   ")
	if result != nil {
		t.Fatalf("expected nil for empty string, got %v", result)
	}

	result = nullableTrimmed("")
	if result != nil {
		t.Fatalf("expected nil for empty string, got %v", result)
	}
}
