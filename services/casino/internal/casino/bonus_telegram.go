package casino

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type bonusVerificationResult struct {
	Status           string
	Message          string
	TelegramUserID   int64
	TelegramUsername string
	TelegramChannel  string
}

type telegramInitDataUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type telegramInitDataPayload struct {
	User telegramInitDataUser
}

type TelegramAuth struct {
	UserID     string
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	AuthDate   int64
	InitData   string
}

func verifyBonusSubscription(ctx context.Context, initData string) (bonusVerificationResult, error) {
	botToken := BonusTelegramBotToken()
	channel := BonusTelegramChannel()
	if botToken == "" || channel == "" {
		return bonusVerificationResult{
			Status:  "not_configured",
			Message: "telegram bonus verification is not configured",
		}, ErrBonusVerificationNotConfigured
	}

	initData = strings.TrimSpace(initData)
	if initData == "" {
		return bonusVerificationResult{
			Status:  "verification_required",
			Message: "telegram init data is required",
		}, ErrBonusVerificationRequired
	}

	payload, err := ValidateInitData(botToken, initData)
	if err != nil {
		return bonusVerificationResult{
			Status:  "denied",
			Message: err.Error(),
		}, ErrBonusVerificationDenied
	}

	ok, err := verifyTelegramChannelMembership(ctx, botToken, channel, payload.TelegramID)
	if err != nil {
		return bonusVerificationResult{
			Status:  "unavailable",
			Message: err.Error(),
		}, ErrBonusVerificationUnavailable
	}
	if !ok {
		return bonusVerificationResult{
			Status:           "denied",
			Message:          "telegram channel membership required",
			TelegramUserID:   payload.TelegramID,
			TelegramUsername: strings.TrimSpace(payload.Username),
			TelegramChannel:  channel,
		}, ErrBonusVerificationDenied
	}

	username := strings.TrimSpace(payload.Username)
	if username == "" {
		username = strings.TrimSpace(payload.FirstName)
	}

	return bonusVerificationResult{
		Status:           "verified",
		Message:          "telegram channel membership verified",
		TelegramUserID:   payload.TelegramID,
		TelegramUsername: username,
		TelegramChannel:  channel,
	}, nil
}

func verifyTelegramChannelMembership(ctx context.Context, botToken, channel string, telegramUserID int64) (bool, error) {
	if telegramUserID <= 0 {
		return false, errors.New("telegram user id is missing")
	}
	apiBase := BonusTelegramAPIBaseURL()
	endpoint := fmt.Sprintf("%s/bot%s/getChatMember", apiBase, botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, err
	}
	q := req.URL.Query()
	q.Set("chat_id", channel)
	q.Set("user_id", strconv.FormatInt(telegramUserID, 10))
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("telegram api returned %s", resp.Status)
	}

	var payload struct {
		OK     bool `json:"ok"`
		Result struct {
			Status   string `json:"status"`
			IsMember bool   `json:"is_member"`
		} `json:"result"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return false, err
	}
	if !payload.OK {
		if strings.TrimSpace(payload.Description) != "" {
			return false, errors.New(payload.Description)
		}
		return false, errors.New("telegram api rejected getChatMember request")
	}

	switch strings.ToLower(strings.TrimSpace(payload.Result.Status)) {
	case "creator", "administrator", "member":
		return true, nil
	case "restricted":
		return payload.Result.IsMember, nil
	default:
		return false, nil
	}
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

	var user telegramInitDataUser
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

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
