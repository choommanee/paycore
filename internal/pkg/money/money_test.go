package money

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestToMinor(t *testing.T) {
	tests := []struct {
		name     string
		amount   string
		currency string
		want     int64
		wantErr  bool
	}{
		{"thb whole", "100", "THB", 10000, false},
		{"thb satang", "100.50", "THB", 10050, false},
		{"thb one satang", "0.01", "THB", 1, false},
		{"usd cents", "1.99", "USD", 199, false},
		{"jpy zero-decimal", "500", "JPY", 500, false},
		{"jpy fractional rejected", "500.5", "JPY", 0, true},
		{"thb too much precision", "1.005", "THB", 0, true},
		{"unsupported currency", "1.00", "XAU", 0, true},
		{"zero", "0", "THB", 0, false},
		{"large", "99999999.99", "THB", 9999999999, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			amt := decimal.RequireFromString(tc.amount)
			got, err := ToMinor(amt, tc.currency)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got minor=%d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("ToMinor(%s,%s)=%d, want %d", tc.amount, tc.currency, got, tc.want)
			}
		})
	}
}

func TestFromMinor(t *testing.T) {
	tests := []struct {
		minor    int64
		currency string
		want     string
		wantErr  bool
	}{
		{10000, "THB", "100", false},
		{10050, "THB", "100.5", false},
		{1, "THB", "0.01", false},
		{199, "USD", "1.99", false},
		{500, "JPY", "500", false},
		{100, "XAU", "", true},
	}
	for _, tc := range tests {
		got, err := FromMinor(tc.minor, tc.currency)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("FromMinor(%d,%s): expected error", tc.minor, tc.currency)
			}
			continue
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.Equal(decimal.RequireFromString(tc.want)) {
			t.Fatalf("FromMinor(%d,%s)=%s, want %s", tc.minor, tc.currency, got, tc.want)
		}
	}
}

// TestRoundTrip ensures ToMinor -> FromMinor -> ToMinor is stable across
// currencies and precisions.
func TestRoundTrip(t *testing.T) {
	cases := []struct {
		amount   string
		currency string
	}{
		{"100.50", "THB"},
		{"0.01", "USD"},
		{"1234.56", "EUR"},
		{"7", "JPY"},
		{"0.99", "SGD"},
	}
	for _, tc := range cases {
		amt := decimal.RequireFromString(tc.amount)
		minor, err := ToMinor(amt, tc.currency)
		if err != nil {
			t.Fatalf("ToMinor(%s,%s): %v", tc.amount, tc.currency, err)
		}
		back, err := FromMinor(minor, tc.currency)
		if err != nil {
			t.Fatalf("FromMinor: %v", err)
		}
		if !back.Equal(amt) {
			t.Fatalf("round-trip mismatch: %s -> %d -> %s", tc.amount, minor, back)
		}
		again, err := ToMinor(back, tc.currency)
		if err != nil || again != minor {
			t.Fatalf("second ToMinor mismatch: %d != %d (err=%v)", again, minor, err)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		amount   string
		currency string
		wantErr  bool
	}{
		{"1.00", "THB", false},
		{"0", "THB", true},     // not positive
		{"-5", "THB", true},    // negative
		{"1.005", "THB", true}, // too much precision
		{"1.00", "XAU", true},  // unsupported
	}
	for _, tc := range tests {
		err := Validate(decimal.RequireFromString(tc.amount), tc.currency)
		if (err != nil) != tc.wantErr {
			t.Fatalf("Validate(%s,%s) err=%v wantErr=%v", tc.amount, tc.currency, err, tc.wantErr)
		}
	}
}
