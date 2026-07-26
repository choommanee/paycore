package middleware

import "testing"

// TestRedactSecretPath asserts the checkout session token is stripped from a
// path before it is logged (it is a bearer credential carried in the URL).
func TestRedactSecretPath(t *testing.T) {
	cases := map[string]string{
		"/v1/checkout/sessions/cs_deadbeef":     "/v1/checkout/sessions/<redacted>",
		"/v1/checkout/sessions/cs_deadbeef/pay": "/v1/checkout/sessions/<redacted>/pay",
		"/v1/payment-links":                     "/v1/payment-links",
		"/v1/checkout/sessions":                 "/v1/checkout/sessions",
		"/healthz":                              "/healthz",
	}
	for in, want := range cases {
		if got := redactSecretPath(in); got != want {
			t.Errorf("redactSecretPath(%q) = %q, want %q", in, got, want)
		}
	}
}
