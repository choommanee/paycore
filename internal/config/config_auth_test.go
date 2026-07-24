package config

import (
	"testing"
	"time"
)

func TestSessionTTLDefault(t *testing.T) {
	c := &Config{SessionTTLHours: 0}
	if got := c.SessionTTL(); got != 168*time.Hour {
		t.Fatalf("SessionTTL()=%v want 168h (default)", got)
	}
}

func TestSessionTTLFromHours(t *testing.T) {
	c := &Config{SessionTTLHours: 24}
	if got := c.SessionTTL(); got != 24*time.Hour {
		t.Fatalf("SessionTTL()=%v want 24h", got)
	}
}
