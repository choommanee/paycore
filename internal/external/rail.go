package external

import (
	"context"

	"github.com/shopspring/decimal"
)

// RailChargeRequest asks a PaymentRail to register an expected inbound payment.
// PaymentID is our internal order/payment identifier: it is both the
// reconciliation key and, for crypto rails, the on-chain transfer memo — so an
// observed settlement maps back to exactly one order without minting a fresh
// address per payment.
type RailChargeRequest struct {
	PaymentID string          // internal order/payment id; reconciliation key + memo
	Amount    decimal.Decimal // expected amount in the asset's major units
	Asset     string          // settlement asset symbol, e.g. "USDC", "TCH", "THB"
	ExpirySec int             // how long the charge stays payable (0 => rail default)
}

// RailInstructions is everything a payer needs to complete a charge. Fields are
// rail-specific and optional: a crypto rail fills Address/Memo/Asset/ChainID; a
// QR/bank rail would fill Payload; a hosted/redirect method fills RedirectURL.
type RailInstructions struct {
	Address      string // crypto: deposit address the payer sends to
	Memo         string // crypto: TIP-20 transfer memo (== PaymentID) for reconciliation
	Asset        string // settlement asset symbol
	AssetAddress string // crypto: TIP-20 token contract address
	ChainID      int64  // crypto: EVM chain id (ThaiChain = 7)
	FeeSponsored bool   // crypto: gas paid by the merchant (feePayer) — payer holds no gas token
	URI          string // optional deep-link / payment URI for wallets
	Payload      string // non-crypto payload (e.g. an EMVCo QR string)
	RedirectURL  string // hosted/redirect methods
}

// RailChargeResult is returned when a charge is registered. ProviderRef is the
// rail-side handle a caller passes back to Confirm.
type RailChargeResult struct {
	ProviderRef   string
	Instructions  RailInstructions
	ExpiresAtUnix int64
}

// RailSettlementStatus is the normalized settlement lifecycle shared across every
// rail, so callers branch on one enum regardless of whether the money moved over
// a card network, a bank QR, or a blockchain.
type RailSettlementStatus string

const (
	RailPending   RailSettlementStatus = "pending"   // registered; not yet settled with finality
	RailConfirmed RailSettlementStatus = "confirmed" // settled and final (enough confirmations / signed callback)
	RailUnderpaid RailSettlementStatus = "underpaid" // observed a payment for less than the expected amount
	RailExpired   RailSettlementStatus = "expired"   // not settled before expiry
	RailFailed    RailSettlementStatus = "failed"    // settlement failed / reverted
)

// RailConfirmation is the current settlement state of a previously created
// charge. Confirmations/RequiredConfirmations model rail-native finality (block
// depth for crypto; 0 for rails that settle atomically via a signed callback).
type RailConfirmation struct {
	Status                RailSettlementStatus
	TxRef                 string          // on-chain tx hash / bank reference (empty until observed)
	Confirmations         int             // finality depth observed so far
	RequiredConfirmations int             // depth required before Status becomes RailConfirmed
	PaidAmount            decimal.Decimal // actual amount observed (may differ from requested)
	SettledAtUnix         int64           // when settlement reached finality (0 until confirmed)
}

// PaymentRail abstracts a settlement rail PayCore can accept an inbound payment
// over — a card acquirer, a QR/bank rail, or a crypto/stablecoin chain. It
// generalizes "register an expected payment, then confirm it settled with
// rail-native finality," so adding a new rail (USDC-on-Base, ThaiChain, a local
// bank switch) is a new implementation rather than a change to callers.
//
// It follows the same mock-first pattern as Acquirer and QRProvider: a
// deterministic sandbox implementation is gated by SANDBOX_MODE for local/test
// use, and a production provider slots into the identical interface.
type PaymentRail interface {
	// Name is the stable rail identifier surfaced as a payment method
	// (e.g. "thaichain").
	Name() string

	// CreateCharge registers an expected inbound payment and returns the
	// instructions a payer needs to complete it. Safe to treat as idempotent by
	// PaymentID at the caller layer.
	CreateCharge(ctx context.Context, req RailChargeRequest) (*RailChargeResult, error)

	// Confirm reports the current settlement state of a charge by its ProviderRef.
	// It is safe to call repeatedly (polling): it checks rail-native finality and
	// only reports RailConfirmed once the settlement is final.
	Confirm(ctx context.Context, providerRef string) (*RailConfirmation, error)
}
