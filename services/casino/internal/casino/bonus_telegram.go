package casino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goritskimihail/mudro/pkg/tgauth"
)

var telegramHTTPClient struct {
	once sync.Once
	c    *http.Client
}

func getTelegramHTTPClient() *http.Client {
	telegramHTTPClient.once.Do(func() {
		telegramHTTPClient.c = &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     60 * time.Second,
				MaxIdleConnsPerHost: 5,
			},
		}
	})
	return telegramHTTPClient.c
}

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

	payload, err := tgauth.ValidateInitData(botToken, initData)
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

	client := getTelegramHTTPClient()
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
