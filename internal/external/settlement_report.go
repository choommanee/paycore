package external

import (
	"context"
	"time"
)

// SettlementLine is one merchant/currency total from an acquirer's daily
// settlement report. Amounts are int64 minor units to match the ledger.
type SettlementLine struct {
	MerchantID  string
	Currency    string
	AmountMinor int64
}

// SettlementReporter fetches the acquirer's settlement report for a given date.
// Reconciliation compares these figures against the gateway's own ledger totals.
type SettlementReporter interface {
	SettlementReport(ctx context.Context, date time.Time) ([]SettlementLine, error)
}

// MockSettlementReporter returns a caller-supplied report. Tests and the local
// reconciliation worker inject lines (or leave it empty) to exercise the
// mismatch-detection path without a real acquirer connection.
type MockSettlementReporter struct {
	// Lines is returned verbatim for any date. Swap it in tests to simulate
	// matches and deltas.
	Lines []SettlementLine
}

// NewMockSettlementReporter returns a reporter that always yields lines.
func NewMockSettlementReporter(lines ...SettlementLine) *MockSettlementReporter {
	return &MockSettlementReporter{Lines: lines}
}

var _ SettlementReporter = (*MockSettlementReporter)(nil)

// SettlementReport returns the configured lines regardless of date.
func (m *MockSettlementReporter) SettlementReport(_ context.Context, _ time.Time) ([]SettlementLine, error) {
	return m.Lines, nil
}
