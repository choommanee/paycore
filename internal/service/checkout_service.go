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
}

type checkoutService struct {
	repo    repository.Querier
	charger Charger
	qr      QRIssuer
	vault   Tokenizer
	sandbox bool
	log     zerolog.Logger
}

// NewCheckoutService wires the checkout service. sandbox gates the raw-PAN card
// path: when false, card pay is refused (real hosted-fields tokenization is out
// of scope). vault/charger/qr may be nil in scaffolding but are required for the
// pay paths.
func NewCheckoutService(repo repository.Querier, charger Charger, qr QRIssuer, vault Tokenizer, sandbox bool, log zerolog.Logger) CheckoutService {
	return &checkoutService{
		repo:    repo,
		charger: charger,
		qr:      qr,
		vault:   vault,
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
		return nil, domain.ErrCheckoutMethodUnavailable
	}
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
	// Tokenize the PAN so the PaymentService receives a vault token (never a PAN).
	tok, _, err := s.vault.Tokenize(ctx, req.Card.Number)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	amount, err := money.FromMinor(row.AmountMinor, row.Currency)
	if err != nil {
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
		// error (mapped to 402 by the error handler) so the page can show it.
		_, _ = s.transition(ctx, row, domain.CheckoutFailed)
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
	if next == domain.CheckoutPaid {
		s.markLinkPaid(ctx, updated)
	}
	view := s.buildView(ctx, updated, nil)
	view.NextActionURL = p.NextActionURL
	return view, nil
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

// markLinkPaid closes a single_use link once its session is paid. Best-effort:
// a failure is logged, not fatal to the payment.
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
	if _, err := s.repo.UpdatePaymentLinkStatus(ctx, repository.UpdatePaymentLinkStatusParams{
		ID: row.PaymentLinkID, MerchantID: row.MerchantID, Status: "paid",
	}); err != nil {
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
