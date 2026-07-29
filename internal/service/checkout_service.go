package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/external"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/pkg/money"
	"github.com/yourco/payment-gateway/internal/repository"
)

// checkoutSessionTTL is the short lifetime of a hosted-checkout session.
const checkoutSessionTTL = 30 * time.Minute

// Charger is the slice of PaymentService the checkout needs (card charge).
// *paymentService satisfies it.
type Charger interface {
	Create(ctx context.Context, idemKey string, req domain.CreatePaymentRequest) (*domain.Payment, error)
}

// QRIssuer is the slice of QRService the checkout needs (PromptPay mint + poll).
// *qrService satisfies it.
type QRIssuer interface {
	Create(ctx context.Context, req domain.CreateQRRequest) (*domain.QRPayment, error)
	Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.QRPayment, error)
}

// Tokenizer turns a raw PAN into a vault token. *crypto.Vault satisfies it. Used
// ONLY on the sandbox card path; the token is handed to the payment service,
// which detokenizes it inside PCI scope.
type Tokenizer interface {
	Tokenize(ctx context.Context, pan string) (token string, last4 string, err error)
}

// CheckoutService drives the public hosted-checkout page. It turns a payment link
// into a short-lived session authenticated by an opaque token (stored hashed),
// initiates card / PromptPay payments via the existing services, and reports
// status for polling. The session token is the ONLY credential; every operation
// is scoped to the session it resolves — merchant context comes from the row,
// never the caller.
type CheckoutService interface {
	CreateFromLink(ctx context.Context, publicID string) (*domain.CheckoutSessionView, error)
	Get(ctx context.Context, token string) (*domain.CheckoutSessionView, error)
	Pay(ctx context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error)
	// ConfirmMock simulates a wallet approve (true) / decline (false) for a session
	// awaiting action. SANDBOX ONLY — the HTTP route is absent in production.
	ConfirmMock(ctx context.Context, token string, approve bool) (*domain.CheckoutSessionView, error)
}

type checkoutService struct {
	repo    repository.Querier
	charger Charger
	qr      QRIssuer
	vault   Tokenizer
	rail    external.PaymentRail
	sandbox bool
	log     zerolog.Logger
}

// NewCheckoutService wires the checkout service. sandbox gates the raw-PAN card
// path: when false, card pay is refused (real hosted-fields tokenization is out
// of scope). vault/charger/qr may be nil in scaffolding but are required for the
// pay paths. rail is the crypto/stablecoin settlement rail (nil disables the
// "thaichain" method); it is used sandbox-only in this phase.
func NewCheckoutService(repo repository.Querier, charger Charger, qr QRIssuer, vault Tokenizer, rail external.PaymentRail, sandbox bool, log zerolog.Logger) CheckoutService {
	return &checkoutService{
		repo:    repo,
		charger: charger,
		qr:      qr,
		vault:   vault,
		rail:    rail,
		sandbox: sandbox,
		log:     log.With().Str("service", "checkout").Logger(),
	}
}

// CreateFromLink opens a session for an active, unexpired payment link. The raw
// token is returned once (on the view); only its hash is persisted.
func (s *checkoutService) CreateFromLink(ctx context.Context, publicID string) (*domain.CheckoutSessionView, error) {
	if s.repo == nil {
		return nil, domain.ErrCheckoutSessionNotFound
	}
	link, err := s.repo.GetPaymentLinkByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentLinkNotFound
		}
		return nil, err
	}
	// A link must be active and unexpired to open a checkout. We return the same
	// 404 for missing / disabled / expired so the endpoint is not an existence or
	// state oracle.
	if link.Status != "active" {
		return nil, domain.ErrPaymentLinkNotFound
	}
	if link.ExpiresAt.Valid && time.Now().After(link.ExpiresAt.Time) {
		return nil, domain.ErrPaymentLinkNotFound
	}

	token, err := generateSessionToken()
	if err != nil {
		return nil, err
	}
	row, err := s.repo.CreateCheckoutSession(ctx, repository.CreateCheckoutSessionParams{
		ID:               toPgUUID(uuid.New()),
		MerchantID:       link.MerchantID,
		PaymentLinkID:    link.ID,
		SessionTokenHash: middleware.HashAPIKey(token),
		AmountMinor:      link.AmountMinor,
		Currency:         link.Currency,
		Status:           string(domain.CheckoutOpen),
		SelectedMethod:   "",
		CustomerEmail:    "",
		ReturnUrl:        "",
		ExpiresAt:        pgTimestamptz(time.Now().Add(checkoutSessionTTL)),
		// PaymentID / QrPaymentID left as zero pgtype.UUID (NULL) until a method runs.
	})
	if err != nil {
		return nil, err
	}
	view := s.buildView(ctx, row, &link)
	view.Token = token
	return view, nil
}

// Get returns the current session view. It first syncs live state: an expired
// session is swept to expired; a PromptPay session reflects its QR payment's
// confirmed/expired/failed state (and closes a single_use link on paid).
func (s *checkoutService) Get(ctx context.Context, token string) (*domain.CheckoutSessionView, error) {
	row, err := s.loadByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	row, qr, err := s.syncStatus(ctx, row)
	if err != nil {
		return nil, err
	}
	view := s.buildView(ctx, row, nil)
	// Re-surface the PromptPay payload so a reloaded page can re-render the QR.
	// qr is whatever syncStatus already fetched (nil if it fetched nothing, or if
	// the session moved to a terminal state) — reused here so Get never polls the
	// QR provider a second time for the same request.
	if row.SelectedMethod == "promptpay" && row.QrPaymentID.Valid &&
		row.Status == string(domain.CheckoutRequiresAction) && qr != nil {
		view.QRPayload = qr.QRPayload
	}
	// Re-surface the on-chain instructions so a reloaded rail page can re-render
	// the deposit address + memo. railInstructions is idempotent (keyed by session
	// id) and never resets a settled charge.
	if domain.IsRailMethod(row.SelectedMethod) &&
		row.Status == string(domain.CheckoutRequiresAction) && s.rail != nil {
		if instr, ierr := s.railInstructions(ctx, row); ierr == nil {
			view.RailInstructions = instr
		}
	}
	return view, nil
}

// syncStatus advances a non-terminal session based on live state. Terminal
// states (paid/failed/expired) are returned unchanged. It also returns the
// *domain.QRPayment fetched from the provider, if any, so callers (Get) can
// reuse it instead of polling the provider again; it is nil when no QR fetch
// occurred (or when the session moved to a terminal state).
func (s *checkoutService) syncStatus(ctx context.Context, row repository.CheckoutSession) (repository.CheckoutSession, *domain.QRPayment, error) {
	status := domain.CheckoutStatus(row.Status)
	if status == domain.CheckoutPaid || status == domain.CheckoutFailed || status == domain.CheckoutExpired {
		return row, nil, nil
	}
	if row.ExpiresAt.Valid && time.Now().After(row.ExpiresAt.Time) {
		updated, err := s.transition(ctx, row, domain.CheckoutExpired)
		return updated, nil, err
	}
	if row.SelectedMethod == "promptpay" && row.QrPaymentID.Valid &&
		status == domain.CheckoutRequiresAction && s.qr != nil {
		qr, err := s.qr.Get(ctx, pgUUIDToUUID(row.MerchantID), pgUUIDToUUID(row.QrPaymentID))
		if err != nil {
			return row, nil, nil // transient read error: report no change, poll again
		}
		switch qr.Status {
		case domain.QRPaid:
			updated, terr := s.transition(ctx, row, domain.CheckoutPaid)
			if terr != nil {
				return row, nil, terr
			}
			s.markLinkPaid(ctx, updated)
			return updated, nil, nil
		case domain.QRExpired:
			updated, err := s.transition(ctx, row, domain.CheckoutExpired)
			return updated, nil, err
		case domain.QRFailed:
			updated, err := s.transition(ctx, row, domain.CheckoutFailed)
			return updated, nil, err
		}
		// Still awaiting payment: hand back the payload we already fetched so
		// Get can re-surface it without a second poll.
		return row, qr, nil
	}
	// Crypto/stablecoin rail: poll the rail for on-chain finality.
	if domain.IsRailMethod(row.SelectedMethod) && status == domain.CheckoutRequiresAction && s.rail != nil {
		conf, err := s.rail.Confirm(ctx, pgUUIDToUUID(row.ID).String())
		if err != nil {
			return row, nil, nil // transient / charge not in memory: report no change, poll again
		}
		switch conf.Status {
		case external.RailConfirmed:
			updated, terr := s.transition(ctx, row, domain.CheckoutPaid)
			if terr != nil {
				return row, nil, terr
			}
			s.markLinkPaid(ctx, updated)
			return updated, nil, nil
		case external.RailFailed:
			updated, err := s.transition(ctx, row, domain.CheckoutFailed)
			return updated, nil, err
		case external.RailExpired:
			updated, err := s.transition(ctx, row, domain.CheckoutExpired)
			return updated, nil, err
		}
		// pending / underpaid: stay in requires_action and poll again.
		return row, nil, nil
	}
	return row, nil, nil
}

// Pay initiates payment for an open session with the selected method. It is
// idempotent: a session that is no longer open returns its current view (a
// double-submit is a no-op), and an expired session is swept to expired.
func (s *checkoutService) Pay(ctx context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	row, err := s.loadByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if row.ExpiresAt.Valid && time.Now().After(row.ExpiresAt.Time) {
		_, _ = s.transition(ctx, row, domain.CheckoutExpired)
		return nil, domain.ErrCheckoutSessionExpired
	}
	// Only an open session can be paid; anything else returns its current view
	// (idempotent double-submit).
	if domain.CheckoutStatus(row.Status) != domain.CheckoutOpen {
		return s.Get(ctx, token)
	}
	// The method must be one this link + this phase actually supports.
	allowed := domain.DisplayMethods(s.linkAllowedMethods(ctx, row))
	if !containsMethod(allowed, req.Method) {
		return nil, domain.ErrCheckoutMethodUnavailable
	}
	switch req.Method {
	case "promptpay":
		return s.payPromptPay(ctx, row, req)
	case "card":
		return s.payCard(ctx, row, req)
	default:
		// The six Phase 4 wallet / redirect methods share one mock adapter.
		if domain.IsWalletMethod(req.Method) {
			return s.payWallet(ctx, row, req)
		}
		// Crypto/stablecoin rails (e.g. ThaiChain) register an on-chain charge.
		if domain.IsRailMethod(req.Method) {
			return s.payThaiChain(ctx, row, req)
		}
		return nil, domain.ErrCheckoutMethodUnavailable
	}
}

// payThaiChain registers a crypto/stablecoin charge on the PaymentRail and moves
// the session to requires_action carrying the on-chain payment instructions
// (deposit address + memo == session id + asset). It is SANDBOX-ONLY in this
// phase: a real deploy confirms deposits via a chain-watcher, so with sandbox off
// it refuses rather than surface an unconfirmable address. Like payWallet it
// charges nothing here and reserves a single_use link so two sessions cannot both
// complete the same link; the sandbox /confirm-mock endpoint simulates the
// deposit and flips the session to paid (or failed, releasing the reservation).
func (s *checkoutService) payThaiChain(ctx context.Context, row repository.CheckoutSession, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	if !s.sandbox || s.rail == nil {
		return nil, domain.ErrCheckoutMethodUnavailable
	}

	// Reserve a single_use link BEFORE going to requires_action — identical to the
	// card / wallet paths.
	var reservedLink *repository.PaymentLink
	if row.PaymentLinkID.Valid {
		link, err := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
			ID: row.PaymentLinkID, MerchantID: row.MerchantID,
		})
		if err != nil {
			return nil, err
		}
		if link.LinkType == "single_use" {
			consumed, cerr := s.repo.ConsumePaymentLinkIfActive(ctx, repository.ConsumePaymentLinkIfActiveParams{
				ID: row.PaymentLinkID, MerchantID: row.MerchantID,
			})
			if cerr != nil {
				if errors.Is(cerr, pgx.ErrNoRows) {
					_, _ = s.transition(ctx, row, domain.CheckoutFailed)
					return nil, domain.ErrCheckoutMethodUnavailable
				}
				return nil, cerr
			}
			reservedLink = &consumed
		}
	}

	instr, err := s.railInstructions(ctx, row)
	if err != nil {
		s.releaseLinkReservation(ctx, reservedLink, row)
		return nil, err
	}

	updated, err := s.repo.UpdateCheckoutSession(ctx, repository.UpdateCheckoutSessionParams{
		ID:             row.ID,
		Status:         string(domain.CheckoutRequiresAction),
		SelectedMethod: req.Method,
		PaymentID:      row.PaymentID,   // stays NULL
		QrPaymentID:    row.QrPaymentID, // stays NULL
		CustomerEmail:  strFallback(req.CustomerEmail, row.CustomerEmail),
	})
	if err != nil {
		s.releaseLinkReservation(ctx, reservedLink, row)
		return nil, err
	}
	view := s.buildView(ctx, updated, nil)
	view.RailInstructions = instr
	return view, nil
}

// railInstructions registers (idempotently) the rail charge for a session and
// returns its public on-chain payment instructions. The charge is keyed by the
// session id, which is also the transfer memo — so this is safe to call again on
// a later Get to re-surface the instructions for a reloaded page (it never resets
// an already-settled charge). Amount is the session amount; in this sandbox phase
// the stablecoin amount tracks the fiat amount 1:1, a real deploy applies FX.
func (s *checkoutService) railInstructions(ctx context.Context, row repository.CheckoutSession) (*domain.CheckoutRailInstructions, error) {
	amount, err := money.FromMinor(row.AmountMinor, row.Currency)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	res, err := s.rail.CreateCharge(ctx, external.RailChargeRequest{
		PaymentID: pgUUIDToUUID(row.ID).String(),
		Amount:    amount,
	})
	if err != nil {
		return nil, err
	}
	in := res.Instructions
	return &domain.CheckoutRailInstructions{
		Asset:        in.Asset,
		Address:      in.Address,
		Memo:         in.Memo,
		ChainID:      in.ChainID,
		FeeSponsored: in.FeeSponsored,
		URI:          in.URI,
		ExplorerURL:  external.ThaiChainExplorer,
	}, nil
}

func (s *checkoutService) payPromptPay(ctx context.Context, row repository.CheckoutSession, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	if s.qr == nil {
		return nil, domain.ErrCheckoutMethodUnavailable
	}
	amount, err := money.FromMinor(row.AmountMinor, row.Currency)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	qr, err := s.qr.Create(ctx, domain.CreateQRRequest{
		MerchantID: pgUUIDToUUID(row.MerchantID),
		Method:     domain.QRPromptPayDynamic,
		Amount:     amount,
		Currency:   row.Currency,
	})
	if err != nil {
		return nil, err
	}
	updated, err := s.repo.UpdateCheckoutSession(ctx, repository.UpdateCheckoutSessionParams{
		ID:             row.ID,
		Status:         string(domain.CheckoutRequiresAction),
		SelectedMethod: "promptpay",
		PaymentID:      row.PaymentID,
		QrPaymentID:    toPgUUID(qr.ID),
		CustomerEmail:  strFallback(req.CustomerEmail, row.CustomerEmail),
	})
	if err != nil {
		return nil, err
	}
	view := s.buildView(ctx, updated, nil)
	view.QRPayload = qr.QRPayload
	return view, nil
}

func (s *checkoutService) payCard(ctx context.Context, row repository.CheckoutSession, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	// Receiving a raw PAN is sandbox-only. In production, real hosted-fields
	// tokenization is out of scope, so refuse rather than accept card data.
	if !s.sandbox {
		return nil, domain.ErrCheckoutMethodUnavailable
	}
	if req.Card == nil {
		return nil, domain.ErrInvalidRequest
	}
	if s.vault == nil || s.charger == nil {
		return nil, domain.ErrCheckoutMethodUnavailable
	}

	// single_use links must be reserved BEFORE we tokenize/charge: this is the
	// only thing that prevents two concurrent sessions from a single_use link
	// both charging the card. ConsumePaymentLinkIfActive atomically flips the
	// link active -> paid; if it reports no row, another session already
	// consumed it (or it's not active for another reason) and we must refuse
	// without charging. Reusable links skip this — they may be paid repeatedly.
	var reservedLink *repository.PaymentLink
	if row.PaymentLinkID.Valid {
		link, err := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
			ID: row.PaymentLinkID, MerchantID: row.MerchantID,
		})
		// Fail closed on ANY load error (including ErrNoRows): a session that
		// references a link must be able to load it to know whether it is
		// single_use and needs reservation. If we cannot verify it, refuse to
		// charge rather than risk a double payment against an unverifiable link.
		if err != nil {
			return nil, err
		}
		if link.LinkType == "single_use" {
			consumed, cerr := s.repo.ConsumePaymentLinkIfActive(ctx, repository.ConsumePaymentLinkIfActiveParams{
				ID: row.PaymentLinkID, MerchantID: row.MerchantID,
			})
			if cerr != nil {
				if errors.Is(cerr, pgx.ErrNoRows) {
					// Link already consumed by another session (or otherwise
					// unavailable): do not charge. Fail the session and report
					// the method as unavailable rather than double-charging.
					_, _ = s.transition(ctx, row, domain.CheckoutFailed)
					return nil, domain.ErrCheckoutMethodUnavailable
				}
				return nil, cerr
			}
			reservedLink = &consumed
		}
	}

	// Tokenize the PAN so the PaymentService receives a vault token (never a PAN).
	tok, _, err := s.vault.Tokenize(ctx, req.Card.Number)
	if err != nil {
		s.releaseLinkReservation(ctx, reservedLink, row)
		return nil, domain.ErrInvalidRequest
	}
	amount, err := money.FromMinor(row.AmountMinor, row.Currency)
	if err != nil {
		s.releaseLinkReservation(ctx, reservedLink, row)
		return nil, domain.ErrInvalidRequest
	}
	sessionID := pgUUIDToUUID(row.ID)
	p, err := s.charger.Create(ctx, "cs_"+sessionID.String(), domain.CreatePaymentRequest{
		MerchantID:   pgUUIDToUUID(row.MerchantID),
		Amount:       amount,
		Currency:     row.Currency,
		PaymentToken: tok,
		Capture:      true,
		ReturnURL:    row.ReturnUrl,
	})
	if err != nil {
		// A decline / insufficient funds marks the session failed and surfaces the
		// error (mapped to 402 by the error handler) so the page can show it. Since
		// no money moved, release the reservation so the single_use link can be
		// retried.
		_, _ = s.transition(ctx, row, domain.CheckoutFailed)
		s.releaseLinkReservation(ctx, reservedLink, row)
		return nil, err
	}
	next := domain.CheckoutPaid
	if p.Status == domain.StatusRequiresAction {
		next = domain.CheckoutRequiresAction
	}
	updated, err := s.repo.UpdateCheckoutSession(ctx, repository.UpdateCheckoutSessionParams{
		ID:             row.ID,
		Status:         string(next),
		SelectedMethod: "card",
		PaymentID:      toPgUUID(p.ID),
		QrPaymentID:    row.QrPaymentID,
		CustomerEmail:  strFallback(req.CustomerEmail, row.CustomerEmail),
	})
	if err != nil {
		return nil, err
	}
	// NOTE: the single_use link (if any) was already flipped to 'paid' by the
	// reservation above; markLinkPaid is a no-op for it here (idempotent) and
	// still handles the reusable-link case (which is a no-op by link_type).
	if next == domain.CheckoutPaid {
		s.markLinkPaid(ctx, updated)
	}
	view := s.buildView(ctx, updated, nil)
	view.NextActionURL = p.NextActionURL
	return view, nil
}

// payWallet is the MOCK adapter for the Phase 4 e-wallet / redirect methods
// (mobile_banking, truemoney, shopeepay, alipay, wechat, card_installment). It is
// SANDBOX-ONLY: in production these are stubs for a real PSP redirect, which is
// out of scope, so it refuses. Unlike card/promptpay it charges NOTHING — it
// reserves a single_use link (exactly like the card path, so two sessions cannot
// both complete the same link) and moves the session to requires_action. The
// sandbox-only ConfirmMock endpoint then flips it to paid (closing the link) or
// failed (releasing the reservation). No payments/qr_payments row is created and
// no money is converted or moved.
func (s *checkoutService) payWallet(ctx context.Context, row repository.CheckoutSession, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	if !s.sandbox {
		// Real PSP redirect integration is out of scope: refuse rather than
		// silently mark anything paid.
		return nil, domain.ErrCheckoutMethodUnavailable
	}

	// Reserve a single_use link BEFORE going to requires_action — identical to the
	// card path. ConsumePaymentLinkIfActive atomically flips active -> paid; no row
	// means another session already consumed it, so refuse without proceeding.
	var reservedLink *repository.PaymentLink
	if row.PaymentLinkID.Valid {
		link, err := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
			ID: row.PaymentLinkID, MerchantID: row.MerchantID,
		})
		// Fail closed on any load error: a session referencing a link must load it
		// to know whether it is single_use and needs reservation.
		if err != nil {
			return nil, err
		}
		if link.LinkType == "single_use" {
			consumed, cerr := s.repo.ConsumePaymentLinkIfActive(ctx, repository.ConsumePaymentLinkIfActiveParams{
				ID: row.PaymentLinkID, MerchantID: row.MerchantID,
			})
			if cerr != nil {
				if errors.Is(cerr, pgx.ErrNoRows) {
					_, _ = s.transition(ctx, row, domain.CheckoutFailed)
					return nil, domain.ErrCheckoutMethodUnavailable
				}
				return nil, cerr
			}
			reservedLink = &consumed
		}
	}

	updated, err := s.repo.UpdateCheckoutSession(ctx, repository.UpdateCheckoutSessionParams{
		ID:             row.ID,
		Status:         string(domain.CheckoutRequiresAction),
		SelectedMethod: req.Method,
		PaymentID:      row.PaymentID,   // stays NULL
		QrPaymentID:    row.QrPaymentID, // stays NULL
		CustomerEmail:  strFallback(req.CustomerEmail, row.CustomerEmail),
	})
	if err != nil {
		// Persisting requires_action failed: release the reservation so the
		// single_use link can be retried.
		s.releaseLinkReservation(ctx, reservedLink, row)
		return nil, err
	}
	// No QRPayload / NextActionURL: the mock approve/decline is driven inline by
	// the frontend against ConfirmMock; there is no external redirect.
	return s.buildView(ctx, updated, nil), nil
}

// ConfirmMock completes a wallet session in requires_action: approve -> paid
// (closing a single_use link), decline -> failed (releasing the reservation). It
// is SANDBOX ONLY (the route is not mounted in production) and applies ONLY to a
// wallet session awaiting action; anything else (already terminal, promptpay
// QR-awaiting, expired) is a no-op that returns the current view. It scopes
// everything to the session the token resolves; merchant context comes from the
// row.
func (s *checkoutService) ConfirmMock(ctx context.Context, token string, approve bool) (*domain.CheckoutSessionView, error) {
	if !s.sandbox {
		return nil, domain.ErrCheckoutMethodUnavailable
	}
	row, err := s.loadByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if row.ExpiresAt.Valid && time.Now().After(row.ExpiresAt.Time) {
		_, _ = s.transition(ctx, row, domain.CheckoutExpired)
		return nil, domain.ErrCheckoutSessionExpired
	}
	// Only a wallet or crypto-rail session actually awaiting action can be
	// confirmed. Anything else is an idempotent no-op returning the current view
	// (double-submit, an already-terminal session, or a promptpay requires_action
	// which confirms via the QR webhook instead).
	isWallet := domain.IsWalletMethod(row.SelectedMethod)
	isRail := domain.IsRailMethod(row.SelectedMethod)
	if domain.CheckoutStatus(row.Status) != domain.CheckoutRequiresAction || (!isWallet && !isRail) {
		return s.Get(ctx, token)
	}

	if approve {
		// For a crypto rail, approving simulates the on-chain deposit (sandbox
		// payer-simulator) so the charge reaches finality before we mark paid. The
		// deposited amount tracks the session amount, so it is never underpaid.
		if isRail {
			if sr, ok := s.rail.(external.SandboxRail); ok {
				if amount, aerr := money.FromMinor(row.AmountMinor, row.Currency); aerr == nil {
					if derr := sr.SimulateDeposit(pgUUIDToUUID(row.ID).String(), amount); derr != nil {
						s.log.Warn().Err(derr).Msg("thaichain sandbox deposit failed")
					}
				}
			}
		}
		updated, err := s.transition(ctx, row, domain.CheckoutPaid)
		if err != nil {
			return nil, err
		}
		// single_use link was already reserved (flipped to paid) in payWallet /
		// payThaiChain; this is an idempotent no-op for it and handles reusable too.
		s.markLinkPaid(ctx, updated)
		return s.buildView(ctx, updated, nil), nil
	}

	updated, err := s.transition(ctx, row, domain.CheckoutFailed)
	if err != nil {
		return nil, err
	}
	// Release the single_use reservation so the link can be retried. Load the link
	// to pass releaseLinkReservation a non-nil pointer ONLY when it is single_use
	// (reusable links were never reserved, so pass nil -> no-op).
	var reserved *repository.PaymentLink
	if row.PaymentLinkID.Valid {
		if l, lerr := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
			ID: row.PaymentLinkID, MerchantID: row.MerchantID,
		}); lerr == nil && l.LinkType == "single_use" {
			reserved = &l
		}
	}
	s.releaseLinkReservation(ctx, reserved, row)
	return s.buildView(ctx, updated, nil), nil
}

// releaseLinkReservation undoes a single_use link reservation made by
// ConsumePaymentLinkIfActive when the charge did not go through (no money
// moved), so the link can be retried. No-op if nothing was reserved.
func (s *checkoutService) releaseLinkReservation(ctx context.Context, reserved *repository.PaymentLink, row repository.CheckoutSession) {
	if reserved == nil {
		return
	}
	// Conditional release: only revert to 'active' if the link is still 'paid'
	// (the state we reserved it in). pgx.ErrNoRows means the status changed under
	// us (e.g. a concurrent Disable) — leave it as-is rather than reviving it.
	if _, err := s.repo.ReleasePaymentLinkReservation(ctx, repository.ReleasePaymentLinkReservationParams{
		ID: row.PaymentLinkID, MerchantID: row.MerchantID,
	}); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		s.log.Warn().Err(err).Msg("release single_use link reservation failed")
	}
}

// ---- shared helpers (also used by Get in Task 5) ---------------------------

func (s *checkoutService) loadByToken(ctx context.Context, token string) (repository.CheckoutSession, error) {
	if s.repo == nil {
		return repository.CheckoutSession{}, domain.ErrCheckoutSessionNotFound
	}
	row, err := s.repo.GetCheckoutSessionByTokenHash(ctx, middleware.HashAPIKey(token))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.CheckoutSession{}, domain.ErrCheckoutSessionNotFound
		}
		return repository.CheckoutSession{}, err
	}
	return row, nil
}

// transition updates only the status, preserving the row's other fields.
func (s *checkoutService) transition(ctx context.Context, row repository.CheckoutSession, status domain.CheckoutStatus) (repository.CheckoutSession, error) {
	return s.repo.UpdateCheckoutSession(ctx, repository.UpdateCheckoutSessionParams{
		ID:             row.ID,
		Status:         string(status),
		SelectedMethod: row.SelectedMethod,
		PaymentID:      row.PaymentID,
		QrPaymentID:    row.QrPaymentID,
		CustomerEmail:  row.CustomerEmail,
	})
}

// markLinkPaid closes a single_use link once its session is paid. It is used
// by both the card path (where the link was already reserved to 'paid' by
// ConsumePaymentLinkIfActive before charging, so this call is a benign no-op)
// and the PromptPay QRPaid path in syncStatus, where nothing reserved the link
// up front and this call performs the actual atomic consume.
//
// It uses ConsumePaymentLinkIfActive (active -> paid, conditionally) rather
// than the old non-conditional UpdatePaymentLinkStatus so that it is
// idempotent and safe to call more than once, and so a link that is no longer
// active (already consumed) is left alone instead of being stomped. Best
// effort: pgx.ErrNoRows (already consumed / not single_use / not found) is a
// benign no-op; other errors are logged, not fatal to the payment.
//
// KNOWN LIMITATION (accepted, not fixed here): PromptPay is async — a
// customer can screenshot/share a single_use link's QR, or a merchant could
// (incorrectly) let two customers scan two independently-minted QRs from the
// same link before either settles at the bank. Both could complete payment at
// the bank before either QRPaid webhook/poll arrives, since there is nothing
// to reserve until a QR payment actually confirms. This method still ensures
// only the FIRST QRPaid transition closes the link (subsequent ones are
// no-ops here), so the link cannot be re-paid via checkout after the first
// confirmation, but it cannot prevent two out-of-band bank transfers that
// both raced ahead of the link being marked paid.
func (s *checkoutService) markLinkPaid(ctx context.Context, row repository.CheckoutSession) {
	if !row.PaymentLinkID.Valid {
		return
	}
	link, err := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
		ID: row.PaymentLinkID, MerchantID: row.MerchantID,
	})
	if err != nil || link.LinkType != "single_use" {
		return
	}
	if _, err := s.repo.ConsumePaymentLinkIfActive(ctx, repository.ConsumePaymentLinkIfActiveParams{
		ID: row.PaymentLinkID, MerchantID: row.MerchantID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return // already consumed (or otherwise not active) — benign no-op
		}
		s.log.Warn().Err(err).Msg("mark link paid failed")
	}
}

// linkAllowedMethods returns the session link's allowed methods (nil if none).
func (s *checkoutService) linkAllowedMethods(ctx context.Context, row repository.CheckoutSession) []string {
	if !row.PaymentLinkID.Valid {
		return nil
	}
	link, err := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
		ID: row.PaymentLinkID, MerchantID: row.MerchantID,
	})
	if err != nil {
		return nil
	}
	return link.AllowedMethods
}

func containsMethod(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func strFallback(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}

// buildView renders the public, secret-free projection. link may be passed in to
// avoid a re-read; when nil and the session references one, it is loaded (scoped
// by merchant id). Merchant name is looked up best-effort.
func (s *checkoutService) buildView(ctx context.Context, row repository.CheckoutSession, link *repository.PaymentLink) *domain.CheckoutSessionView {
	if link == nil && row.PaymentLinkID.Valid {
		if l, err := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
			ID: row.PaymentLinkID, MerchantID: row.MerchantID,
		}); err == nil {
			link = &l
		}
	}
	name := ""
	if m, err := s.repo.GetMerchant(ctx, row.MerchantID); err == nil {
		name = m.Name
	}
	view := &domain.CheckoutSessionView{
		ID:             pgUUIDToUUID(row.ID),
		Status:         row.Status,
		AmountMinor:    row.AmountMinor,
		Currency:       row.Currency,
		MerchantName:   name,
		SelectedMethod: row.SelectedMethod,
		ReturnURL:      row.ReturnUrl,
		Sandbox:        s.sandbox,
	}
	if row.ExpiresAt.Valid {
		view.ExpiresAt = row.ExpiresAt.Time
	}
	var allowed []string
	if link != nil {
		view.Title = link.Title
		view.Description = link.Description
		view.ImageURL = link.ImageUrl
		allowed = link.AllowedMethods
	}
	view.AllowedMethods = domain.DisplayMethods(allowed)
	return view
}

// generateSessionToken returns a cryptographically random, URL-safe opaque token
// with a recognizable prefix. Only its hash is persisted.
func generateSessionToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "cs_" + hex.EncodeToString(buf), nil
}
