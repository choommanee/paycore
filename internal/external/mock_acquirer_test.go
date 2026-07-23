package external

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
)

func TestMockAcquirerAuthorizeApproved(t *testing.T) {
	a := NewMockAcquirer()
	res, err := a.Authorize(context.Background(), AuthRequest{
		PAN:      "4111111111111111",
		Amount:   decimal.RequireFromString("10.00"),
		Currency: "THB",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Approved {
		t.Fatalf("expected approval, got decline %q", res.DeclineCode)
	}
	if res.AuthCode == "" || res.AcquirerRef == "" {
		t.Fatal("approved auth must carry auth code and acquirer ref")
	}
	if res.CardBrand != "visa" {
		t.Fatalf("brand=%s want visa", res.CardBrand)
	}
	if res.Last4 != "1111" {
		t.Fatalf("last4=%s want 1111", res.Last4)
	}
}

func TestMockAcquirerDeclineInsufficientFunds(t *testing.T) {
	a := NewMockAcquirer()
	res, err := a.Authorize(context.Background(), AuthRequest{PAN: "4000000000000000"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Approved {
		t.Fatal("expected decline")
	}
	if res.DeclineCode != DeclineInsufficientFunds {
		t.Fatalf("decline code=%s want %s", res.DeclineCode, DeclineInsufficientFunds)
	}
}

func TestMockAcquirer3DSRequiredThenSatisfied(t *testing.T) {
	a := NewMockAcquirer()
	// Without a cryptogram: 3DS required.
	res, err := a.Authorize(context.Background(), AuthRequest{PAN: "4000000000000119"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Approved || res.DeclineCode != Decline3DSRequired {
		t.Fatalf("expected 3ds_required decline, got approved=%v code=%s", res.Approved, res.DeclineCode)
	}
	// With a cryptogram present: approved.
	res2, err := a.Authorize(context.Background(), AuthRequest{PAN: "4000000000000119", ThreeDSCryptogram: "abc123"})
	if err != nil {
		t.Fatal(err)
	}
	if !res2.Approved {
		t.Fatal("expected approval once cryptogram supplied")
	}
}

func TestMockAcquirerBrandDetection(t *testing.T) {
	a := NewMockAcquirer()
	cases := map[string]string{
		"4111111111111111": "visa",
		"5555555555554444": "mastercard",
		"2221000000000009": "mastercard",
		"378282246310005":  "amex",
		"6011000000000004": "unknown",
	}
	for pan, want := range cases {
		res, err := a.Authorize(context.Background(), AuthRequest{PAN: pan})
		if err != nil {
			t.Fatal(err)
		}
		if res.CardBrand != want {
			t.Fatalf("pan %s brand=%s want %s", pan, res.CardBrand, want)
		}
	}
}

func TestMockAcquirerCaptureRefundVoid(t *testing.T) {
	a := NewMockAcquirer()
	ctx := context.Background()
	amt := decimal.RequireFromString("5.00")

	if _, err := a.Capture(ctx, "", amt); err == nil {
		t.Fatal("capture with empty ref should error")
	}
	if r, err := a.Capture(ctx, "ref-1", amt); err != nil || !r.Approved {
		t.Fatalf("capture failed: %v approved=%v", err, r != nil && r.Approved)
	}
	if r, err := a.Refund(ctx, "ref-1", amt); err != nil || !r.Approved {
		t.Fatalf("refund failed: %v", err)
	}
	if r, err := a.Void(ctx, "ref-1"); err != nil || !r.Approved {
		t.Fatalf("void failed: %v", err)
	}
	if _, err := a.Refund(ctx, "", amt); err == nil {
		t.Fatal("refund with empty ref should error")
	}
	if _, err := a.Void(ctx, ""); err == nil {
		t.Fatal("void with empty ref should error")
	}
}

func TestMockThreeDSFrictionless(t *testing.T) {
	tds := NewMockThreeDS("")
	url, crypto, err := tds.Authenticate(context.Background(), "4111111111111111",
		decimal.RequireFromString("10.00"), "THB", "https://return")
	if err != nil {
		t.Fatal(err)
	}
	if url != "" {
		t.Fatalf("frictionless flow should not return challenge url, got %s", url)
	}
	if crypto == "" {
		t.Fatal("frictionless flow must return a cryptogram")
	}
	// Cryptogram is deterministic.
	again := tds.Cryptogram("4111111111111111", decimal.RequireFromString("10.00"), "THB")
	if again != crypto {
		t.Fatal("cryptogram must be deterministic")
	}
}

func TestMockThreeDSChallengeAndFailure(t *testing.T) {
	tds := NewMockThreeDS("")
	ctx := context.Background()
	amt := decimal.RequireFromString("10.00")

	url, crypto, err := tds.Authenticate(ctx, "4000000000000119", amt, "THB", "https://ret")
	if err != nil {
		t.Fatal(err)
	}
	if url == "" || crypto != "" {
		t.Fatalf("challenge flow should return url and empty cryptogram, got url=%q crypto=%q", url, crypto)
	}

	if _, _, err := tds.Authenticate(ctx, "4000000000000000", amt, "THB", ""); err == nil {
		t.Fatal("expected authentication failure for not-enrolled card")
	}
}
