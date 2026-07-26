package domain

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

// TestCreatePaymentLinkRejectsUnsafeInput locks two hardening rules from the
// Phase 2 review: image_url must be an http(s) URL (a javascript: scheme is a
// stored-XSS vector once Phase 3 renders it), and currency must be alphabetic
// (a non-alpha 3-char code passes len=3 but crashes Intl.NumberFormat client-side).
func TestCreatePaymentLinkRejectsUnsafeInput(t *testing.T) {
	v := validator.New()

	cases := []struct {
		name    string
		req     CreatePaymentLinkRequest
		wantErr bool
	}{
		{"javascript image_url", CreatePaymentLinkRequest{Title: "x", AmountMinor: 1, ImageURL: "javascript:alert(1)"}, true},
		{"https image_url ok", CreatePaymentLinkRequest{Title: "x", AmountMinor: 1, ImageURL: "https://cdn/x.png"}, false},
		{"non-alpha currency", CreatePaymentLinkRequest{Title: "x", AmountMinor: 1, Currency: "1$@"}, true},
		{"THB currency ok", CreatePaymentLinkRequest{Title: "x", AmountMinor: 1, Currency: "THB"}, false},
		{"no optional fields", CreatePaymentLinkRequest{Title: "x", AmountMinor: 1}, false},
	}
	for _, tc := range cases {
		err := v.Struct(tc.req)
		if tc.wantErr && err == nil {
			t.Errorf("%s: expected validation error, got none", tc.name)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("%s: expected no error, got %v", tc.name, err)
		}
	}
}
