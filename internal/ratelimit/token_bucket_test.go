package ratelimit

import (
	"testing"
	"time"
)

func TestTokenBucketAllow(t *testing.T) {
	b := NewTokenBucket(2, 2)
	now := time.Now()

	if !b.Allow(now) {
		t.Fatal("first allow should pass")
	}
	if !b.Allow(now) {
		t.Fatal("second allow should pass")
	}
	if b.Allow(now) {
		t.Fatal("third allow should be limited")
	}
	if !b.Allow(now.Add(600 * time.Millisecond)) {
		t.Fatal("should refill after time")
	}
}

func TestTokenBucketDisabledRate(t *testing.T) {
	b := NewTokenBucket(0, 1)
	if b.Allow(time.Now()) {
		t.Fatal("disabled rate should not allow")
	}
}
