package tgauth

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestValidateInitData_Valid_FakeHash(t *testing.T) {
	const botToken = "test_bot_token"
	const rawData = "user=%7B%22id%22%3A12345%7D&auth_date=1234567890&hash=fakehash"

	_, err := ValidateInitData(botToken, rawData)
	if err == nil {
		t.Fatal("expected error for fake hash")
	}
}

func TestValidateInitData_EmptyInitData(t *testing.T) {
	_, err := ValidateInitData("test_token", "")
	if err == nil {
		t.Fatal("expected error for empty initData")
	}
	if err.Error() != "empty initData" {
		t.Fatalf("expected 'empty initData', got %q", err.Error())
	}
}

func TestValidateInitData_InvalidQuery(t *testing.T) {
	_, err := ValidateInitData("test_token", "%%invalid")
	if err == nil {
		t.Fatal("expected error for invalid query")
	}
	if !strings.Contains(err.Error(), "parse initData") {
		t.Fatalf("expected parse error, got %q", err.Error())
	}
}

func TestValidateInitData_MissingHash(t *testing.T) {
	_, err := ValidateInitData("test_token", "user=%7B%22id%22%3A1%7D&auth_date=1234567890")
	if err == nil {
		t.Fatal("expected error for missing hash")
	}
	if err.Error() != "missing hash" {
		t.Fatalf("expected 'missing hash', got %q", err.Error())
	}
}

func TestValidateInitData_MissingUser(t *testing.T) {
	botToken := "test_bot_token"
	authDate := "1234567890"

	// Build hash WITHOUT user field (only auth_date)
	pairs := []string{
		"auth_date=" + authDate,
	}
	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	dataCheckString := strings.Join(pairs, "\n")
	hashBytes := hmacSHA256(secret, []byte(dataCheckString))
	expectedHash := hex.EncodeToString(hashBytes)

	// Use valid hash for query without user
	rawData := "auth_date=" + authDate + "&hash=" + expectedHash
	_, err := ValidateInitData(botToken, rawData)
	if err == nil {
		t.Fatal("expected error for missing user")
	}
	if err.Error() != "missing user field" {
		t.Fatalf("expected 'missing user field', got %q", err.Error())
	}
}

func TestValidateInitData_InvalidUserJSON(t *testing.T) {
	botToken := "test_bot_token"
	authDate := "1234567890"

	// Build hash WITH invalid user JSON
	userStr := "notjson"
	pairs := []string{
		"auth_date=" + authDate,
		"user=" + userStr,
	}
	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	dataCheckString := strings.Join(pairs, "\n")
	hashBytes := hmacSHA256(secret, []byte(dataCheckString))
	expectedHash := hex.EncodeToString(hashBytes)

	// Use valid hash with invalid user JSON
	rawData := "user=" + userStr + "&auth_date=" + authDate + "&hash=" + expectedHash
	_, err := ValidateInitData(botToken, rawData)
	if err == nil {
		t.Fatal("expected error for invalid user JSON")
	}
	if !strings.Contains(err.Error(), "parse user") {
		t.Fatalf("expected parse user error, got %q", err.Error())
	}
}

func TestValidateInitData_InvalidHash(t *testing.T) {
	rawData := "user=%7B%22id%22%3A12345%7D&auth_date=1234567890&hash=invalidhash"
	_, err := ValidateInitData("test_bot_token", rawData)
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}
	if err.Error() != "invalid hash" {
		t.Fatalf("expected 'invalid hash', got %q", err.Error())
	}
}

func TestValidateInitData_ValidHash(t *testing.T) {
	botToken := "test_bot_token"
	userJSON := `{"id":12345,"username":"testuser","first_name":"Test"}`
	authDate := "1234567890"

	// Build hash using decoded values (as the code does after url.ParseQuery)
	pairs := []string{
		"auth_date=" + authDate,
		"user=" + userJSON,
	}
	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	dataCheckString := strings.Join(pairs, "\n")
	hashBytes := hmacSHA256(secret, []byte(dataCheckString))
	expectedHash := hex.EncodeToString(hashBytes)

	// URL-encode the user JSON for the raw data
	rawData := "user=%7B%22id%22%3A12345%2C%22username%22%3A%22testuser%22%2C%22first_name%22%3A%22Test%22%7D&auth_date=" + authDate + "&hash=" + expectedHash

	result, err := ValidateInitData(botToken, rawData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TelegramID != 12345 {
		t.Fatalf("expected TelegramID 12345, got %d", result.TelegramID)
	}
	if result.Username != "testuser" {
		t.Fatalf("expected username 'testuser', got %q", result.Username)
	}
	if result.FirstName != "Test" {
		t.Fatalf("expected first_name 'Test', got %q", result.FirstName)
	}
	if result.UserID != "tg_12345" {
		t.Fatalf("expected UserID 'tg_12345', got %q", result.UserID)
	}
	if result.AuthDate != 1234567890 {
		t.Fatalf("expected AuthDate 1234567890, got %d", result.AuthDate)
	}
}

func TestValidateInitData_WrongBotToken(t *testing.T) {
	botToken := "test_bot_token"
	userJSON := `{"id":12345,"username":"testuser","first_name":"Test"}`
	authDate := "1234567890"

	pairs := []string{
		"auth_date=" + authDate,
		"user=" + userJSON,
	}
	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	dataCheckString := strings.Join(pairs, "\n")
	hashBytes := hmacSHA256(secret, []byte(dataCheckString))
	expectedHash := hex.EncodeToString(hashBytes)

	rawData := "user=%7B%22id%22%3A12345%2C%22username%22%3A%22testuser%22%2C%22first_name%22%3A%22Test%22%7D&auth_date=" + authDate + "&hash=" + expectedHash

	_, err := ValidateInitData("WRONG_TOKEN:AAA", rawData)
	if err == nil {
		t.Fatal("expected error for wrong bot token")
	}
	if err.Error() != "invalid hash" {
		t.Fatalf("expected 'invalid hash', got %q", err.Error())
	}
}
