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

// Get and Pay are implemented in Tasks 4–5. Stubbed so the file compiles now.
func (s *checkoutService) Get(ctx context.Context, token string) (*domain.CheckoutSessionView, error) {
	return nil, domain.ErrNotImplemented
}

func (s *checkoutService) Pay(ctx context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	return nil, domain.ErrNotImplemented
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
