package casino

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtractInitDataFromBodyPreservesRequestBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/casino/game/bet", strings.NewReader(`{"initData":"abc","betAmount":10}`))

	initData, err := extractInitDataFromBody(req)
	if err != nil {
		t.Fatalf("extractInitDataFromBody: %v", err)
	}
	if initData != "abc" {
		t.Fatalf("unexpected initData: %q", initData)
	}

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("ReadAll(req.Body): %v", err)
	}

	var payload struct {
		InitData  string  `json:"initData"`
		BetAmount float64 `json:"betAmount"`
	}
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		t.Fatalf("json.Unmarshal(body): %v", err)
	}
	if payload.InitData != "abc" || payload.BetAmount != 10 {
		t.Fatalf("body was not preserved: %+v", payload)
	}
}
