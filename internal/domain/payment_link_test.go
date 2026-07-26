package domain

import "testing"

func TestIsValidMethod(t *testing.T) {
	for _, m := range []string{"card", "promptpay", "truemoney"} {
		if !IsValidMethod(m) {
			t.Fatalf("%q should be valid", m)
		}
	}
	for _, m := range []string{"", "bitcoin", "CARD"} {
		if IsValidMethod(m) {
			t.Fatalf("%q should be invalid", m)
		}
	}
}
