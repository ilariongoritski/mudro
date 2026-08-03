package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIP_XRealIPValid(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Real-IP", "203.0.113.5")
	r.RemoteAddr = "10.0.0.1:54321"
	if got := clientIP(r); got != "203.0.113.5" {
		t.Fatalf("expected 203.0.113.5, got %q", got)
	}
}

func TestClientIP_XRealIPInvalidFallsBackToRemoteAddr(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Real-IP", "not-an-ip")
	r.RemoteAddr = "10.0.0.2:54321"
	if got := clientIP(r); got != "10.0.0.2" {
		t.Fatalf("expected fallback to 10.0.0.2, got %q", got)
	}
}

func TestClientIP_RemoteAddrStripsPort(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "192.0.2.10:8080"
	if got := clientIP(r); got != "192.0.2.10" {
		t.Fatalf("expected 192.0.2.10, got %q", got)
	}
}

func TestClientIP_RemoteAddrNoPort(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "192.0.2.20"
	if got := clientIP(r); got != "192.0.2.20" {
		t.Fatalf("expected 192.0.2.20, got %q", got)
	}
}

func TestClientIP_EmptyRemoteAddr(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = ""
	if got := clientIP(r); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
