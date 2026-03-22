package casino

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TelegramAuth struct {
	UserID     string
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	AuthDate   int64
	InitData   string
}

type telegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func ValidateInitData(botToken, rawInitData string) (*TelegramAuth, error) {
	if rawInitData == "" {
		return nil, errors.New("empty initData")
	}

	values, err := url.ParseQuery(rawInitData)
	if err != nil {
		return nil, fmt.Errorf("parse initData: %w", err)
	}

	receivedHash := values.Get("hash")
	if receivedHash == "" {
		return nil, errors.New("missing hash")
	}

	// Build data-check-string
	pairs := make([]string, 0)
	for key := range values {
		if key == "hash" {
			continue
		}
		pairs = append(pairs, key+"="+values.Get(key))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// HMAC-SHA256
	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	expectedHash := hex.EncodeToString(hmacSHA256(secret, []byte(dataCheckString)))

	if !hmac.Equal([]byte(expectedHash), []byte(receivedHash)) {
		return nil, errors.New("invalid hash")
	}

	// Parse auth_date
	authDateStr := values.Get("auth_date")
	authDate, _ := strconv.ParseInt(authDateStr, 10, 64)

	// Parse user
	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, errors.New("missing user field")
	}

	var user telegramUser
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, fmt.Errorf("parse user: %w", err)
	}

	return &TelegramAuth{
		UserID:     fmt.Sprintf("tg_%d", user.ID),
		TelegramID: user.ID,
		Username:   user.Username,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		AuthDate:   authDate,
		InitData:   rawInitData,
	}, nil
}

func DevInitData(botToken string, telegramID int64) string {
	user := telegramUser{
		ID:        telegramID,
		Username:  "demo_player",
		FirstName: "Demo",
		LastName:  "Player",
	}
	userJSON, _ := json.Marshal(user)

	params := url.Values{}
	params.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("query_id", "dev-"+randomHex(16))
	params.Set("user", string(userJSON))

	pairs := make([]string, 0)
	for key := range params {
		pairs = append(pairs, key+"="+params.Get(key))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	hash := hex.EncodeToString(hmacSHA256(secret, []byte(dataCheckString)))

	params.Set("hash", hash)
	return params.Encode()
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
