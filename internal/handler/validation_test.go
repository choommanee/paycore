package handler

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestToWireName(t *testing.T) {
	cases := map[string]string{
		"SettlementCurrency": "settlement_currency",
		"Name":               "name",
		"MCC":                "m_c_c", // acronym: acceptable, still snake-ish and stable
		"Amount":             "amount",
	}
	for in, want := range cases {
		if got := toWireName(in); got != want {
			t.Fatalf("toWireName(%q)=%q want %q", in, got, want)
		}
	}
}

// sample mirrors the shape validated by handlers to drive real
// validator.ValidationErrors through validationDetails.
type sample struct {
	Name               string `validate:"required"`
	SettlementCurrency string `validate:"required,len=3"`
}

func TestValidationDetailsFieldOriented(t *testing.T) {
	v := validator.New()
	err := v.Struct(sample{SettlementCurrency: "TH"}) // Name missing, currency wrong len
	if err == nil {
		t.Fatal("expected validation error")
	}
	details := validationDetails(err)
	if len(details) != 2 {
		t.Fatalf("want 2 field errors, got %d: %+v", len(details), details)
	}
	byField := map[string]FieldError{}
	for _, d := range details {
		byField[d.Field] = d
	}
	if fe, ok := byField["name"]; !ok || fe.Code != "required" {
		t.Fatalf("name/required missing: %+v", details)
	}
	if fe, ok := byField["settlement_currency"]; !ok || fe.Code != "len" {
		t.Fatalf("settlement_currency/len missing: %+v", details)
	}
	if byField["settlement_currency"].Message == "" {
		t.Fatal("expected a human message for len failure")
	}
}

func TestValidationDetailsNonValidationError(t *testing.T) {
	details := validationDetails(errors.New("boom"))
	if len(details) != 1 || details[0].Code != "invalid" {
		t.Fatalf("non-validation error should yield one generic entry: %+v", details)
	}
}
