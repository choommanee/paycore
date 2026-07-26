package domain

import (
	"reflect"
	"testing"
)

func TestDisplayMethodsEmptyMeansAllSupported(t *testing.T) {
	got := DisplayMethods(nil)
	if !reflect.DeepEqual(got, []string{"card", "promptpay"}) {
		t.Fatalf("empty allowed = %v want [card promptpay]", got)
	}
}

func TestDisplayMethodsIntersectsAndKeepsSupportedOrder(t *testing.T) {
	// Link allows a Phase-4 wallet + promptpay + card, in a different order.
	got := DisplayMethods([]string{"truemoney", "promptpay", "card"})
	// Only supported methods survive, in CheckoutSupportedMethods order.
	if !reflect.DeepEqual(got, []string{"card", "promptpay"}) {
		t.Fatalf("got %v want [card promptpay]", got)
	}
}

func TestDisplayMethodsCardOnly(t *testing.T) {
	got := DisplayMethods([]string{"card"})
	if !reflect.DeepEqual(got, []string{"card"}) {
		t.Fatalf("got %v want [card]", got)
	}
}
