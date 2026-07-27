package domain

import (
	"reflect"
	"testing"
)

func TestDisplayMethodsEmptyMeansAllSupported(t *testing.T) {
	got := DisplayMethods(nil)
	want := []string{"card", "promptpay", "mobile_banking", "truemoney", "shopeepay", "alipay", "wechat", "card_installment"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("empty allowed = %v want %v", got, want)
	}
}

func TestDisplayMethodsIntersectsAndKeepsSupportedOrder(t *testing.T) {
	// Link allows a wallet + promptpay + card, in a different order.
	got := DisplayMethods([]string{"truemoney", "promptpay", "card"})
	// All three are now supported; survivors keep CheckoutSupportedMethods order
	// (card, promptpay, then truemoney).
	want := []string{"card", "promptpay", "truemoney"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestDisplayMethodsCardOnly(t *testing.T) {
	got := DisplayMethods([]string{"card"})
	if !reflect.DeepEqual(got, []string{"card"}) {
		t.Fatalf("got %v want [card]", got)
	}
}

func TestDisplayMethodsWalletOnly(t *testing.T) {
	got := DisplayMethods([]string{"shopeepay", "alipay"})
	if !reflect.DeepEqual(got, []string{"shopeepay", "alipay"}) {
		t.Fatalf("got %v want [shopeepay alipay]", got)
	}
}

func TestIsWalletMethod(t *testing.T) {
	for _, m := range []string{"mobile_banking", "truemoney", "shopeepay", "alipay", "wechat", "card_installment"} {
		if !IsWalletMethod(m) {
			t.Fatalf("IsWalletMethod(%q) = false, want true", m)
		}
	}
	for _, m := range []string{"card", "promptpay", "", "paypal"} {
		if IsWalletMethod(m) {
			t.Fatalf("IsWalletMethod(%q) = true, want false", m)
		}
	}
}
