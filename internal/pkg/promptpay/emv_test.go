package promptpay

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

// TestCRC16CheckValue verifies the CRC-16/CCITT-FALSE implementation against the
// canonical check value: CRC of the ASCII string "123456789" is 0x29B1.
func TestCRC16CheckValue(t *testing.T) {
	if got := crc16("123456789"); got != "29B1" {
		t.Fatalf("crc16(\"123456789\")=%s, want 29B1", got)
	}
}

// TestTLV verifies the tag-length-value encoding with 2-digit zero-padded length.
func TestTLV(t *testing.T) {
	cases := []struct {
		id, value, want string
	}{
		{"00", "01", "000201"},
		{"53", "764", "5303764"},
		{"58", "TH", "5802TH"},
	}
	for _, tc := range cases {
		if got := tlv(tc.id, tc.value); got != tc.want {
			t.Fatalf("tlv(%s,%s)=%s want %s", tc.id, tc.value, got, tc.want)
		}
	}
}

func TestNormalizeMobile(t *testing.T) {
	// 0812345678 -> strip leading 0 -> 812345678 -> prefix 66 -> 66812345678
	// then zero-padded to 13 digits -> 0066812345678
	got := normalizeMobile("0812345678")
	if got != "0066812345678" {
		t.Fatalf("normalizeMobile=%s want 0066812345678", got)
	}
	if got := normalizeMobile("08-1234-5678"); got != "0066812345678" {
		t.Fatalf("normalizeMobile with dashes=%s want 0066812345678", got)
	}
}

// TestBuildStatic verifies a static (reusable, no amount) PromptPay payload and
// that its trailing CRC validates.
func TestBuildStatic(t *testing.T) {
	payload, err := Build(Target{MobileNo: "0812345678"}, decimal.Zero, false, "")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	assertValidPayload(t, payload)

	// Static QR: point of initiation "11", no transaction amount tag (54).
	if !strings.Contains(payload, tlv("01", "11")) {
		t.Fatalf("static payload missing POI 11: %s", payload)
	}
	if strings.Contains(payload, "5406") || strings.Contains(payload, tlv("54", "")) {
		t.Fatalf("static payload should not contain amount tag: %s", payload)
	}
	// AID + mobile proxy present.
	if !strings.Contains(payload, "A000000677010111") {
		t.Fatalf("payload missing PromptPay AID: %s", payload)
	}
	if !strings.Contains(payload, "0066812345678") {
		t.Fatalf("payload missing normalized mobile: %s", payload)
	}
}

// TestBuildDynamic verifies a dynamic (one-time) payload carries the amount and
// POI "12", with a valid CRC.
func TestBuildDynamic(t *testing.T) {
	amount := decimal.RequireFromString("100.50")
	payload, err := Build(Target{MobileNo: "0812345678"}, amount, true, "INV-1")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	assertValidPayload(t, payload)

	if !strings.Contains(payload, tlv("01", "12")) {
		t.Fatalf("dynamic payload missing POI 12: %s", payload)
	}
	if !strings.Contains(payload, tlv("54", "100.50")) {
		t.Fatalf("dynamic payload missing amount 100.50: %s", payload)
	}
	// Additional data (tag 62 -> subtag 01 ref).
	if !strings.Contains(payload, tlv("62", tlv("01", "INV-1"))) {
		t.Fatalf("dynamic payload missing ref: %s", payload)
	}
}

func TestBuildNationalIDAndEWallet(t *testing.T) {
	p1, err := Build(Target{NationalID: "1234567890123"}, decimal.Zero, false, "")
	if err != nil {
		t.Fatalf("national id build: %v", err)
	}
	assertValidPayload(t, p1)
	if !strings.Contains(p1, tlv("02", "1234567890123")) {
		t.Fatalf("national id proxy missing: %s", p1)
	}

	p2, err := Build(Target{EWalletID: "123456789012345"}, decimal.Zero, false, "")
	if err != nil {
		t.Fatalf("ewallet build: %v", err)
	}
	assertValidPayload(t, p2)
	if !strings.Contains(p2, tlv("03", "123456789012345")) {
		t.Fatalf("ewallet proxy missing: %s", p2)
	}
}

func TestBuildNoTarget(t *testing.T) {
	if _, err := Build(Target{}, decimal.Zero, false, ""); err == nil {
		t.Fatal("expected error with no target proxy")
	}
}

// assertValidPayload recomputes the CRC over everything up to (and including)
// the "6304" tag header and asserts it matches the trailing 4 hex chars, exactly
// as a wallet would verify a scanned QR.
func assertValidPayload(t *testing.T, payload string) {
	t.Helper()
	if len(payload) < 8 {
		t.Fatalf("payload too short: %q", payload)
	}
	body := payload[:len(payload)-4]
	got := payload[len(payload)-4:]
	// body ends with "6304" (tag 63, len 04). CRC is computed over the whole
	// string including that header.
	if !strings.HasSuffix(body, "6304") {
		t.Fatalf("payload does not end with CRC tag header 6304: %q", body)
	}
	want := crc16(body)
	if got != want {
		t.Fatalf("CRC mismatch: got %s want %s (payload=%s)", got, want, payload)
	}
}
