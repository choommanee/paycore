package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func newClaims() Claims {
	return Claims{UserID: uuid.New(), MerchantID: uuid.New(), Email: "a@b.co"}
}

func TestIssueVerifyRoundTrip(t *testing.T) {
	m := NewManager("test-secret-at-least-32-bytes-long!!", time.Hour)
	c := newClaims()
	tok, err := m.Issue(c)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	got, err := m.Verify(tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got.UserID != c.UserID || got.MerchantID != c.MerchantID || got.Email != c.Email {
		t.Fatalf("claims mismatch: got %+v want %+v", got, c)
	}
}

func TestVerifyRejectsTamper(t *testing.T) {
	m := NewManager("test-secret-at-least-32-bytes-long!!", time.Hour)
	tok, _ := m.Issue(newClaims())
	// flip the last character of the payload
	bad := tok[:len(tok)-1] + "X"
	if _, err := m.Verify(bad); err == nil {
		t.Fatal("Verify accepted a tampered token")
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	tok, _ := NewManager("secret-one-secret-one-secret-one-xx", time.Hour).Issue(newClaims())
	if _, err := NewManager("secret-two-secret-two-secret-two-xx", time.Hour).Verify(tok); err == nil {
		t.Fatal("Verify accepted a token signed with a different secret")
	}
}

func TestVerifyRejectsExpired(t *testing.T) {
	m := NewManager("test-secret-at-least-32-bytes-long!!", -time.Minute) // already expired
	tok, _ := m.Issue(newClaims())
	if _, err := m.Verify(tok); err != ErrExpired {
		t.Fatalf("Verify err=%v want ErrExpired", err)
	}
}
