package server

import (
	"testing"
	"time"
)

func TestRateLimiterBlocksAfterMaxAttempts(t *testing.T) {
	rl := newRateLimiter()
	defer rl.Close()

	ip := "203.0.113.10"
	for i := 0; i < maxAuthAttempts; i++ {
		if !rl.isAllowed(ip) {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}

	if rl.isAllowed(ip) {
		t.Fatal("attempt beyond maxAuthAttempts should be blocked")
	}
}

func TestRateLimiterUnblocksAfterBlockDuration(t *testing.T) {
	rl := newRateLimiter()
	defer rl.Close()

	ip := "198.51.100.7"
	for i := 0; i <= maxAuthAttempts; i++ {
		_ = rl.isAllowed(ip)
	}

	rl.mu.Lock()
	info := rl.attempts[ip]
	info.firstTime = time.Now().Add(-blockDuration - time.Second)
	rl.mu.Unlock()

	if !rl.isAllowed(ip) {
		t.Fatal("IP should be unblocked after blockDuration")
	}
}
