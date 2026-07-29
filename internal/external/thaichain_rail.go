package external

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ThaiChain v2 network constants (Tempo protocol, EVM-compatible). ThaiChain is
// payment-native: gas is paid in a TIP-20 stablecoin rather than a volatile
// token, transfers carry a 32-byte memo, and a feePayer can sponsor gas so the
// payer never holds a gas balance.
const (
	ThaiChainID            int64  = 7
	ThaiChainRPC           string = "https://rpc.thaichain.org"
	ThaiChainExplorer      string = "https://exp.thaichain.org"
	ThaiChainName          string = "thaichain"
	ThaiChainTCHAsset      string = "TCH"
	ThaiChainTCHAddress    string = "0x20C0000000000000000000000000000000000000" // default TIP-20 fee token
	ThaiChainDecimals      int32  = 6                                            // TIP-21: 6-decimal precision
	thaiChainConfirmations int    = 2                                            // block depth required for finality
	thaiChainDefaultExpiry        = 15 * time.Minute
)

// ThaiChainRail is a PaymentRail over ThaiChain v2 (chain id 7). A charge asks
// the payer to send a TIP-20 stablecoin to the deposit address with the
// PaymentID as the transfer memo; we reconcile the on-chain transfer to the
// order by that memo, so no address-per-order scheme is needed. Gas is sponsored
// (feePayer), so the payer holds no gas token — the UX matches a fiat payment.
//
// This is the SANDBOX implementation, gated by SANDBOX_MODE exactly like
// MockAcquirer / MockQRProvider. CreateCharge registers an expected payment in
// memory; SimulateDeposit marks it settled (the sandbox payer-simulator path,
// analogous to the checkout confirm-mock endpoint). The PRODUCTION rail keeps
// this identical interface and instead confirms against the chain — see the
// "Real confirmation" note on Confirm.
type ThaiChainRail struct {
	// DepositAddress is the receiving address payers send to. In sandbox this is
	// a fixed stand-in; a real deploy derives/rotates a per-merchant address.
	DepositAddress string
	// Asset / AssetAddress identify the TIP-20 stablecoin used for settlement.
	Asset        string
	AssetAddress string

	sandbox  bool
	mu       sync.Mutex
	expected map[string]*thaiChainCharge // providerRef -> registered charge
}

// thaiChainCharge is the in-memory record of an expected sandbox payment.
type thaiChainCharge struct {
	req       RailChargeRequest
	memo      string
	expiresAt time.Time

	// Sandbox settlement state, set by SimulateDeposit.
	settled    bool
	txHash     string
	paidAmount decimal.Decimal
	settledAt  time.Time
}

// NewThaiChainRailSandbox returns a sandbox ThaiChain rail. An empty
// depositAddress / asset fall back to the TCH defaults. Gate construction on
// SANDBOX_MODE at the wiring layer, mirroring the other mock adapters.
func NewThaiChainRailSandbox(depositAddress, asset string) *ThaiChainRail {
	if strings.TrimSpace(depositAddress) == "" {
		// Deterministic sandbox stand-in (a zero-ish dev address).
		depositAddress = "0x7HA1C0DE0000000000000000000000000000C0DE"
	}
	assetAddr := ThaiChainTCHAddress
	if strings.TrimSpace(asset) == "" {
		asset = ThaiChainTCHAsset
	}
	return &ThaiChainRail{
		DepositAddress: depositAddress,
		Asset:          asset,
		AssetAddress:   assetAddr,
		sandbox:        true,
		expected:       make(map[string]*thaiChainCharge),
	}
}

// compile-time assertion.
var _ PaymentRail = (*ThaiChainRail)(nil)

// Name reports the rail identifier used as a payment method.
func (r *ThaiChainRail) Name() string { return ThaiChainName }

// CreateCharge registers an expected payment and returns wallet instructions:
// the deposit address, the memo (== PaymentID) the payer must attach, the TIP-20
// asset + contract, chain id, and the fee-sponsored flag. It also builds a
// convenience wallet URI. The memo is how Confirm later reconciles the on-chain
// transfer to this charge.
func (r *ThaiChainRail) CreateCharge(_ context.Context, req RailChargeRequest) (*RailChargeResult, error) {
	if strings.TrimSpace(req.PaymentID) == "" {
		return nil, fmt.Errorf("thaichain: PaymentID is required (used as the transfer memo)")
	}
	if req.Amount.Sign() <= 0 {
		return nil, fmt.Errorf("thaichain: amount must be positive")
	}

	expiry := thaiChainDefaultExpiry
	if req.ExpirySec > 0 {
		expiry = time.Duration(req.ExpirySec) * time.Second
	}

	// The charge is keyed by (and its ProviderRef is) the PaymentID: it is both
	// the on-chain transfer memo and a stable, caller-derivable handle, so a
	// caller that only persists the PaymentID can Confirm without storing a
	// separate rail reference. CreateCharge is idempotent by PaymentID — a repeat
	// call returns the same charge and never resets an already-settled one.
	memo := req.PaymentID
	r.mu.Lock()
	charge, ok := r.expected[memo]
	if !ok {
		charge = &thaiChainCharge{
			req:       req,
			memo:      memo,
			expiresAt: time.Now().Add(expiry),
		}
		r.expected[memo] = charge
	}
	effReq := charge.req // use the registered request so repeat calls are consistent
	expiresAt := charge.expiresAt
	r.mu.Unlock()

	asset := effReq.Asset
	if strings.TrimSpace(asset) == "" {
		asset = r.Asset
	}
	instr := RailInstructions{
		Address:      r.DepositAddress,
		Memo:         memo,
		Asset:        asset,
		AssetAddress: r.AssetAddress,
		ChainID:      ThaiChainID,
		FeeSponsored: true, // gas sponsored via feePayer — payer holds no gas token
		URI:          thaiChainURI(r.DepositAddress, r.AssetAddress, asset, memo, effReq.Amount),
	}
	return &RailChargeResult{
		ProviderRef:   memo,
		Instructions:  instr,
		ExpiresAtUnix: expiresAt.Unix(),
	}, nil
}

// Confirm reports the settlement state of a charge. See confirmAt for the logic;
// this wraps it with the current time.
//
// Real confirmation (production impl, same signature): query the chain via
// ThaiChainRPC — eth_getLogs for the TIP-20 Transfer event to DepositAddress
// whose memo equals the charge's PaymentID, read the transferred amount, and
// require (head − txBlock) >= thaiChainConfirmations before returning
// RailConfirmed. Underpayment and expiry map to the same statuses returned here,
// so callers are unchanged when the sandbox rail is swapped for the real one.
func (r *ThaiChainRail) Confirm(_ context.Context, providerRef string) (*RailConfirmation, error) {
	return r.confirmAt(providerRef, time.Now())
}

// confirmAt is Confirm with an injected clock so expiry is deterministically
// testable (mirrors MockQRProvider.verifyTimestamped taking now).
func (r *ThaiChainRail) confirmAt(providerRef string, now time.Time) (*RailConfirmation, error) {
	r.mu.Lock()
	charge, ok := r.expected[providerRef]
	r.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("thaichain: unknown providerRef %q", providerRef)
	}

	if charge.settled {
		status := RailConfirmed
		if charge.paidAmount.LessThan(charge.req.Amount) {
			status = RailUnderpaid
		}
		return &RailConfirmation{
			Status:                status,
			TxRef:                 charge.txHash,
			Confirmations:         thaiChainConfirmations,
			RequiredConfirmations: thaiChainConfirmations,
			PaidAmount:            charge.paidAmount,
			SettledAtUnix:         charge.settledAt.Unix(),
		}, nil
	}

	if now.After(charge.expiresAt) {
		return &RailConfirmation{
			Status:                RailExpired,
			RequiredConfirmations: thaiChainConfirmations,
			PaidAmount:            decimal.Zero,
		}, nil
	}

	return &RailConfirmation{
		Status:                RailPending,
		RequiredConfirmations: thaiChainConfirmations,
		PaidAmount:            decimal.Zero,
	}, nil
}

// SimulateDeposit marks a sandbox charge as settled with the given paid amount,
// synthesizing an on-chain-style tx hash. This is the sandbox payer-simulator
// entry point (analogous to the checkout confirm-mock endpoint); it exists only
// on the sandbox rail. A paidAmount below the requested amount surfaces as
// RailUnderpaid on the next Confirm.
func (r *ThaiChainRail) SimulateDeposit(providerRef string, paidAmount decimal.Decimal) error {
	if !r.sandbox {
		return fmt.Errorf("thaichain: SimulateDeposit is sandbox-only")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	charge, ok := r.expected[providerRef]
	if !ok {
		return fmt.Errorf("thaichain: unknown providerRef %q", providerRef)
	}
	if charge.settled {
		return nil // idempotent
	}
	charge.settled = true
	charge.paidAmount = paidAmount
	charge.settledAt = time.Now()
	// A deterministic, explorer-shaped fake tx hash for the sandbox.
	charge.txHash = "0x" + strings.ReplaceAll(uuid.NewString(), "-", "") +
		strings.ReplaceAll(uuid.NewString(), "-", "")[:8]
	return nil
}

// thaiChainURI builds a convenience wallet deep-link for a TIP-20 transfer,
// following the EIP-681 shape extended with a memo. Wallets that understand it
// can prefill the transfer; the canonical fields remain Address/Memo/Asset.
func thaiChainURI(address, assetAddress, asset string, memo string, amount decimal.Decimal) string {
	return fmt.Sprintf(
		"ethereum:%s@%d/transfer?address=%s&uint256=%s&memo=%s&symbol=%s",
		assetAddress, ThaiChainID, address, amount.String(), memo, asset,
	)
}
