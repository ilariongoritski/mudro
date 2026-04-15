package casino

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func setTestEnv(t *testing.T, key, value string) {
	t.Helper()
	oldValue, hadValue := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("set env %s: %v", key, err)
	}
	t.Cleanup(func() {
		if !hadValue {
			_ = os.Unsetenv(key)
			return
		}
		_ = os.Setenv(key, oldValue)
	})
}

func TestVerifyBonusSubscriptionRequiresConfiguration(t *testing.T) {
	setTestEnv(t, "CASINO_BONUS_TELEGRAM_BOT_TOKEN", "")
	setTestEnv(t, "TELEGRAM_BOT_TOKEN", "")
	setTestEnv(t, "CASINO_BOT_TOKEN", "")
	setTestEnv(t, "CASINO_BONUS_TELEGRAM_CHANNEL", "")
	setTestEnv(t, "CASINO_BONUS_CHANNEL", "")

	result, err := verifyBonusSubscription(context.Background(), "init-data")
	if err != ErrBonusVerificationNotConfigured {
		t.Fatalf("verifyBonusSubscription() error = %v, want %v", err, ErrBonusVerificationNotConfigured)
	}
	if result.Status != "not_configured" {
		t.Fatalf("status = %q, want not_configured", result.Status)
	}
}

func TestVerifyBonusSubscriptionRequiresInitData(t *testing.T) {
	setTestEnv(t, "CASINO_BONUS_TELEGRAM_BOT_TOKEN", "test-bot-token")
	setTestEnv(t, "CASINO_BONUS_TELEGRAM_CHANNEL", "@mudro_bonus")

	result, err := verifyBonusSubscription(context.Background(), "")
	if err != ErrBonusVerificationRequired {
		t.Fatalf("verifyBonusSubscription() error = %v, want %v", err, ErrBonusVerificationRequired)
	}
	if result.Status != "verification_required" {
		t.Fatalf("status = %q, want verification_required", result.Status)
	}
}

func TestVerifyTelegramChannelMembership(t *testing.T) {
	t.Run("accepts member status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"ok":true,"result":{"status":"member","is_member":true}}`))
		}))
		defer server.Close()

		setTestEnv(t, "CASINO_BONUS_TELEGRAM_API_BASE", server.URL)

		ok, err := verifyTelegramChannelMembership(context.Background(), "token", "@mudro_bonus", 42)
		if err != nil {
			t.Fatalf("verifyTelegramChannelMembership() error = %v", err)
		}
		if !ok {
			t.Fatal("expected membership check to pass")
		}
	})

	t.Run("rejects left status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"ok":true,"result":{"status":"left","is_member":false}}`))
		}))
		defer server.Close()

		setTestEnv(t, "CASINO_BONUS_TELEGRAM_API_BASE", server.URL)

		ok, err := verifyTelegramChannelMembership(context.Background(), "token", "@mudro_bonus", 42)
		if err != nil {
			t.Fatalf("verifyTelegramChannelMembership() error = %v", err)
		}
		if ok {
			t.Fatal("expected membership check to fail")
		}
	})
}
