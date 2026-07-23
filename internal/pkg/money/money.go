package money

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// minorUnits maps ISO-4217 currency to its number of minor-unit digits.
// Money is stored and transmitted as integer minor units (satang/cents) to
// avoid floating point errors; decimal.Decimal is used for arithmetic.
var minorUnits = map[string]int32{
	"THB": 2, "USD": 2, "EUR": 2, "SGD": 2, "JPY": 0,
}

// ToMinor converts a decimal amount to integer minor units for a currency.
func ToMinor(amount decimal.Decimal, currency string) (int64, error) {
	exp, ok := minorUnits[currency]
	if !ok {
		return 0, fmt.Errorf("unsupported currency %q", currency)
	}
	scaled := amount.Shift(exp)
	if !scaled.Equal(scaled.Truncate(0)) {
		return 0, fmt.Errorf("amount %s has more precision than %s allows", amount, currency)
	}
	return scaled.IntPart(), nil
}

// FromMinor converts integer minor units back to a decimal amount.
func FromMinor(minor int64, currency string) (decimal.Decimal, error) {
	exp, ok := minorUnits[currency]
	if !ok {
		return decimal.Zero, fmt.Errorf("unsupported currency %q", currency)
	}
	return decimal.New(minor, -exp), nil
}

// Validate checks the amount is positive and within currency precision.
func Validate(amount decimal.Decimal, currency string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be positive")
	}
	_, err := ToMinor(amount, currency)
	return err
}
