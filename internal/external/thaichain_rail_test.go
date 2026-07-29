package external

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func newRail() *ThaiChainRail { return NewThaiChainRailSandbox("", "") }

func TestThaiChainRail_ImplementsPaymentRail(t *testing.T) {
	var _ PaymentRail = newRail() // compile-time already, keep an explicit runtime check
	if got := newRail().Name(); got != "thaichain" {
		t.Fatalf("Name() = %q, want thaichain", got)
	}
}

func TestThaiChainRail_CreateChargeInstructions(t *testing.T) {
	r := newRail()
	res, err := r.CreateCharge(context.Background(), RailChargeRequest{
		PaymentID: "pay_abc123",
		Amount:    decimal.RequireFromString("49.90"),
		Asset:     "USDC",
	})
	if err != nil {
		t.Fatalf("CreateCharge: %v", err)
	}
	if res.ProviderRef == "" {
		t.Fatal("want a non-empty ProviderRef")
	}
	in := res.Instructions
	if in.Memo != "pay_abc123" {
		t.Errorf("memo = %q, want the PaymentID (reconciliation key)", in.Memo)
	}
	if in.ChainID != ThaiChainID {
		t.Errorf("chainID = %d, want %d", in.ChainID, ThaiChainID)
	}
	if !in.FeeSponsored {
		t.Error("want FeeSponsored=true (gas sponsored via feePayer)")
	}
	if in.Address == "" || in.AssetAddress == "" {
		t.Error("want a deposit address and TIP-20 asset address")
	}
	if in.Asset != "USDC" {
		t.Errorf("asset = %q, want USDC (from request)", in.Asset)
	}
	if in.URI == "" {
		t.Error("want a wallet URI")
	}
	if res.ExpiresAtUnix <= time.Now().Unix() {
		t.Error("want a future expiry")
	}
}

func TestThaiChainRail_ProviderRefIsPaymentIDAndIdempotent(t *testing.T) {
	r := newRail()
	req := RailChargeRequest{PaymentID: "sess_42", Amount: decimal.RequireFromString("10.00")}

	first, err := r.CreateCharge(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCharge: %v", err)
	}
	// ProviderRef is the PaymentID: a caller that persists only the PaymentID can
	// Confirm/SimulateDeposit without storing a separate rail handle.
	if first.ProviderRef != "sess_42" {
		t.Fatalf("ProviderRef = %q, want the PaymentID sess_42", first.ProviderRef)
	}

	// Settle it, then re-create with the same PaymentID: idempotent — the existing
	// (settled) charge is returned, not reset.
	if err := r.SimulateDeposit("sess_42", decimal.RequireFromString("10.00")); err != nil {
		t.Fatalf("SimulateDeposit: %v", err)
	}
	if _, err := r.CreateCharge(context.Background(), req); err != nil {
		t.Fatalf("idempotent CreateCharge: %v", err)
	}
	c, err := r.Confirm(context.Background(), "sess_42")
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if c.Status != RailConfirmed {
		t.Fatalf("status = %q, want confirmed (idempotent re-create must not reset settlement)", c.Status)
	}
}

func TestThaiChainRail_CreateChargeValidates(t *testing.T) {
	r := newRail()
	if _, err := r.CreateCharge(context.Background(), RailChargeRequest{
		Amount: decimal.NewFromInt(10),
	}); err == nil {
		t.Error("want error when PaymentID is empty")
	}
	if _, err := r.CreateCharge(context.Background(), RailChargeRequest{
		PaymentID: "pay_1",
		Amount:    decimal.Zero,
	}); err == nil {
		t.Error("want error when amount is not positive")
	}
}

func TestThaiChainRail_PendingThenConfirmed(t *testing.T) {
	r := newRail()
	amount := decimal.RequireFromString("100.00")
	res, err := r.CreateCharge(context.Background(), RailChargeRequest{
		PaymentID: "pay_confirm",
		Amount:    amount,
	})
	if err != nil {
		t.Fatalf("CreateCharge: %v", err)
	}

	// Before any deposit: pending, no tx, required confirmations advertised.
	c, err := r.Confirm(context.Background(), res.ProviderRef)
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if c.Status != RailPending {
		t.Fatalf("status = %q, want pending", c.Status)
	}
	if c.RequiredConfirmations != thaiChainConfirmations {
		t.Errorf("RequiredConfirmations = %d, want %d", c.RequiredConfirmations, thaiChainConfirmations)
	}
	if c.TxRef != "" {
		t.Error("want empty TxRef before settlement")
	}

	// Simulate an exact-amount on-chain deposit.
	if err := r.SimulateDeposit(res.ProviderRef, amount); err != nil {
		t.Fatalf("SimulateDeposit: %v", err)
	}

	c, err = r.Confirm(context.Background(), res.ProviderRef)
	if err != nil {
		t.Fatalf("Confirm after deposit: %v", err)
	}
	if c.Status != RailConfirmed {
		t.Fatalf("status = %q, want confirmed", c.Status)
	}
	if c.TxRef == "" {
		t.Error("want a tx hash after settlement")
	}
	if !c.PaidAmount.Equal(amount) {
		t.Errorf("PaidAmount = %s, want %s", c.PaidAmount, amount)
	}
	if c.Confirmations < c.RequiredConfirmations {
		t.Errorf("Confirmations = %d, want >= %d", c.Confirmations, c.RequiredConfirmations)
	}
	if c.SettledAtUnix == 0 {
		t.Error("want a settledAt once confirmed")
	}
}

func TestThaiChainRail_Underpaid(t *testing.T) {
	r := newRail()
	res, err := r.CreateCharge(context.Background(), RailChargeRequest{
		PaymentID: "pay_under",
		Amount:    decimal.RequireFromString("100.00"),
	})
	if err != nil {
		t.Fatalf("CreateCharge: %v", err)
	}
	if err := r.SimulateDeposit(res.ProviderRef, decimal.RequireFromString("60.00")); err != nil {
		t.Fatalf("SimulateDeposit: %v", err)
	}
	c, err := r.Confirm(context.Background(), res.ProviderRef)
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if c.Status != RailUnderpaid {
		t.Fatalf("status = %q, want underpaid", c.Status)
	}
	if !c.PaidAmount.Equal(decimal.RequireFromString("60.00")) {
		t.Errorf("PaidAmount = %s, want 60.00", c.PaidAmount)
	}
}

func TestThaiChainRail_Expired(t *testing.T) {
	r := newRail()
	res, err := r.CreateCharge(context.Background(), RailChargeRequest{
		PaymentID: "pay_exp",
		Amount:    decimal.NewFromInt(10),
		ExpirySec: 60,
	})
	if err != nil {
		t.Fatalf("CreateCharge: %v", err)
	}
	// Confirm well past the expiry window via the injected clock.
	c, err := r.confirmAt(res.ProviderRef, time.Now().Add(2*time.Minute))
	if err != nil {
		t.Fatalf("confirmAt: %v", err)
	}
	if c.Status != RailExpired {
		t.Fatalf("status = %q, want expired", c.Status)
	}
}

func TestThaiChainRail_SimulateDepositIdempotentAndUnknownRef(t *testing.T) {
	r := newRail()
	res, err := r.CreateCharge(context.Background(), RailChargeRequest{
		PaymentID: "pay_idem",
		Amount:    decimal.NewFromInt(10),
	})
	if err != nil {
		t.Fatalf("CreateCharge: %v", err)
	}
	if err := r.SimulateDeposit(res.ProviderRef, decimal.NewFromInt(10)); err != nil {
		t.Fatalf("first SimulateDeposit: %v", err)
	}
	// Second deposit is a no-op (does not error, does not change the amount).
	if err := r.SimulateDeposit(res.ProviderRef, decimal.NewFromInt(999)); err != nil {
		t.Fatalf("idempotent SimulateDeposit: %v", err)
	}
	c, _ := r.Confirm(context.Background(), res.ProviderRef)
	if !c.PaidAmount.Equal(decimal.NewFromInt(10)) {
		t.Errorf("PaidAmount = %s, want the first deposit amount 10", c.PaidAmount)
	}

	if err := r.SimulateDeposit("nope", decimal.NewFromInt(1)); err == nil {
		t.Error("want error for unknown providerRef")
	}
	if _, err := r.Confirm(context.Background(), "nope"); err == nil {
		t.Error("want error confirming unknown providerRef")
	}
}
