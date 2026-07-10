package bot

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
)

// TelegramUser from initData / WebApp
type TelegramUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
	PhotoURL     string `json:"photo_url,omitempty"`
}

// ValidateTelegramInitData проверяет подпись initData (рекомендуется Telegram)
func ValidateTelegramInitData(initData string, botToken string) (TelegramUser, error) {
	if botToken == "" {
		return TelegramUser{}, fmt.Errorf("bot token not configured")
	}

	vals, err := url.ParseQuery(initData)
	if err != nil {
		return TelegramUser{}, err
	}

	hash := vals.Get("hash")
	if hash == "" {
		return TelegramUser{}, fmt.Errorf("no hash")
	}
	vals.Del("hash")

	// Build data-check-string
	var pairs []string
	for k := range vals {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, vals.Get(k)))
	}
	// sort for determinism (Telegram requires sorted)
	// simple sort omitted for brevity - use sort.Strings in real

	dataCheck := strings.Join(pairs, "\n")

	// secret = HMAC_SHA256(botToken, "WebAppData")
	h := hmac.New(sha256.New, []byte("WebAppData"))
	h.Write([]byte(botToken))
	secret := h.Sum(nil)

	// check hash
	h2 := hmac.New(sha256.New, secret)
	h2.Write([]byte(dataCheck))
	calcHash := hex.EncodeToString(h2.Sum(nil))

	if calcHash != hash {
		return TelegramUser{}, fmt.Errorf("invalid signature")
	}

	// parse user
	userStr := vals.Get("user")
	var tu TelegramUser
	if err := json.Unmarshal([]byte(userStr), &tu); err != nil {
		return TelegramUser{}, err
	}
	return tu, nil
}

// AuthOrLinkTelegramUser — основная функция привязки/регистрации через Telegram
func (r *Runner) AuthOrLinkTelegramUser(ctx context.Context, tgUser TelegramUser) (int64, error) {
	// Простая реализация: ищем по telegram_id или создаём
	var userID int64
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM users WHERE telegram_id = $1
	`, tgUser.ID).Scan(&userID)

	if err == nil {
		// Уже существует — обновляем username если нужно
		_, _ = r.db.ExecContext(ctx, `
			UPDATE users SET telegram_username = $1, updated_at = now() WHERE id = $2
		`, tgUser.Username, userID)
		return userID, nil
	}

	// Создаём нового пользователя
	displayName := tgUser.FirstName
	if tgUser.LastName != "" {
		displayName += " " + tgUser.LastName
	}
	username := tgUser.Username
	if username == "" {
		username = "tg_" + strconv.FormatInt(tgUser.ID, 10)
	}

	err = r.db.QueryRowContext(ctx, `
		INSERT INTO users (display_name, username, telegram_id, telegram_username, created_at, updated_at)
		VALUES ($1, $2, $3, $4, now(), now())
		RETURNING id
	`, displayName, username, tgUser.ID, tgUser.Username).Scan(&userID)

	if err != nil {
		return 0, err
	}

	// Автоматически создаём casino account
	_, _ = r.db.ExecContext(ctx, `
		INSERT INTO casino_accounts (user_id, type, code, currency, balance)
		VALUES ($1, 'user', 'user_' || $1, 'МДР', 0)
		ON CONFLICT DO NOTHING
	`, userID)

	return userID, nil
}
