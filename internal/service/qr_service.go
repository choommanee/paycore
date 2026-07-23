package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/external"
	"github.com/yourco/payment-gateway/internal/pkg/money"
	"github.com/yourco/payment-gateway/internal/pkg/promptpay"
	"github.com/yourco/payment-gateway/internal/repository"
)

// QRService handles the async QR payment lifecycle.
type QRService interface {
	Create(ctx context.Context, req domain.CreateQRRequest) (*domain.QRPayment, error)
	// Get is scoped to the authenticated merchant so one merchant cannot poll
	// another merchant's QR payment by id.
	Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.QRPayment, error)
	ConfirmFromWebhook(ctx context.Context, signature string, body []byte) (*domain.QRPayment, error)
}

type qrService struct {
	repo     repository.Querier
	txm      TxManager // begins pgx transactions (nil in scaffold)
	provider external.QRProvider
	// merchantTarget resolves a merchant's PromptPay proxy (mobile / national id
	// / e-wallet). In production this comes from the merchant profile.
	promptpayTarget promptpay.Target
	log             zerolog.Logger
}

func NewQRService(repo repository.Querier, txm TxManager, provider external.QRProvider, target promptpay.Target, log zerolog.Logger) QRService {
	return &qrService{
		repo:            repo,
		txm:             txm,
		provider:        provider,
		promptpayTarget: target,
		log:             log.With().Str("service", "qr").Logger(),
	}
}

// Create mints a QR payment.
//
//  1. Validate amount/method.
//  2. For PromptPay: build the EMVCo payload locally (dynamic embeds amount).
//     For card-scheme / cross-border: ask the provider to mint the payload.
//  3. Register the reference upstream for reconciliation.
//  4. Persist qr_payment with status=awaiting_payment + expiry.
//  5. Return payload + expires_at to the merchant (render as a QR image).
func (s *qrService) Create(ctx context.Context, req domain.CreateQRRequest) (*domain.QRPayment, error) {
	if err := validateAmount(req.Amount, req.Currency); err != nil {
		return nil, err
	}
	if err := money.Validate(req.Amount, req.Currency); err != nil {
		return nil, domain.ErrInvalidRequest
	}
	amountMinor, err := money.ToMinor(req.Amount, req.Currency)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	expiry := req.ExpiresIn
	if expiry == 0 {
		expiry = 900 // 15 min default
	}

	qp := &domain.QRPayment{
		ID:         uuid.New(),
		MerchantID: req.MerchantID,
		Method:     req.Method,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Status:     domain.QRAwaitingPayment,
		Reference:  req.Reference,
		ExpiresAt:  time.Now().Add(time.Duration(expiry) * time.Second),
		CreatedAt:  time.Now(),
	}

	// correlationRef is the stable key an inbound webhook is matched against.
	// It is minted here so every method — including locally-built PromptPay QR,
	// which has no upstream provider_ref — can be correlated on confirmation.
	var correlationRef string

	switch req.Method {
	case domain.QRPromptPayDynamic, domain.QRPromptPayStatic:
		dynamic := req.Method == domain.QRPromptPayDynamic
		// PromptPay is built locally with no upstream mint, so we generate a
		// stable correlation reference and embed it in the EMVCo tag-62 payload.
		// The bank echoes this value back in the webhook's `reference` field,
		// letting ConfirmFromWebhook match the callback to this row.
		correlationRef = newPromptPayRef(req.Reference, qp.ID)
		payload, err := promptpay.Build(s.promptpayTarget, req.Amount, dynamic, correlationRef)
		if err != nil {
			return nil, err
		}
		qp.QRPayload = payload
	case domain.QRCardScheme, domain.QRCrossBorder:
		res, err := s.provider.Generate(ctx, external.QRGenRequest{
			Method: string(req.Method), Amount: req.Amount, Currency: req.Currency,
			Reference: req.Reference, ExpirySec: expiry,
		})
		if err != nil {
			return nil, err
		}
		qp.QRPayload = res.Payload
		qp.QRImageURL = res.ImageURL
		qp.ProviderRef = res.ProviderRef
		// The PSP provider_ref doubles as the correlation key for card rails.
		correlationRef = res.ProviderRef
	default:
		return nil, domain.ErrInvalidRequest
	}

	// Persist the minted QR (status=awaiting_payment). The provider Generate call
	// above doubles as registering the reference upstream for reconciliation.
	if s.repo != nil {
		row, err := s.repo.CreateQRPayment(ctx, repository.CreateQRPaymentParams{
			ID:             toPgUUID(qp.ID),
			MerchantID:     toPgUUID(qp.MerchantID),
			Method:         string(qp.Method),
			AmountMinor:    amountMinor,
			Currency:       qp.Currency,
			Status:         string(qp.Status),
			Reference:      qp.Reference,
			QrPayload:      qp.QRPayload,
			QrImageUrl:     strPtr(qp.QRImageURL),
			ProviderRef:    strPtr(qp.ProviderRef),
			CorrelationRef: strPtr(correlationRef),
			ExpiresAt:      pgTimestamptz(qp.ExpiresAt),
		})
		if err != nil {
			// A reused (merchant_id, reference) or correlation_ref hits a UNIQUE
			// index. That is a client duplicate, not a server fault — surface a
			// clean 409 instead of a 500.
			if isUniqueViolation(err) {
				return nil, domain.ErrDuplicateRequest
			}
			return nil, err
		}
		return toDomainQR(row), nil
	}
	qp.CorrelationRef = correlationRef
	return qp, nil
}

// Get returns a QR payment by id, reflecting the persisted state. A QR that is
// still awaiting payment past its expiry is transitioned to expired (lazy
// sweep) so the status-polling path always reports an accurate lifecycle state.
func (s *qrService) Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.QRPayment, error) {
	row, err := s.repo.GetQRPaymentForMerchant(ctx, repository.GetQRPaymentForMerchantParams{
		ID:         toPgUUID(id),
		MerchantID: toPgUUID(merchantID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}

	if row.Status == string(domain.QRAwaitingPayment) && row.ExpiresAt.Valid && time.Now().After(row.ExpiresAt.Time) {
		updated, err := s.repo.UpdateQRStatus(ctx, repository.UpdateQRStatusParams{
			ID:          row.ID,
			Status:      string(domain.QRExpired),
			PayerBank:   row.PayerBank,
			ProviderRef: row.ProviderRef,
			PaidAt:      row.PaidAt,
		})
		if err != nil {
			return nil, err
		}
		return toDomainQR(updated), nil
	}
	return toDomainQR(row), nil
}

// ConfirmFromWebhook is called by the bank/PSP callback.
//
//  1. Verify signature/HMAC (reject if invalid).
//  2. Parse body; look up qr_payment by provider_ref.
//  3. Idempotency: if provider_ref already applied (paid) -> return existing.
//  4. Assert amount/currency match; transition awaiting_payment -> paid.
//  5. Enqueue the merchant webhook (payment.paid) for at-least-once delivery.
func (s *qrService) ConfirmFromWebhook(ctx context.Context, signature string, body []byte) (*domain.QRPayment, error) {
	if err := s.provider.VerifyWebhook(ctx, signature, body); err != nil {
		s.log.Warn().Err(err).Msg("qr webhook signature rejected")
		return nil, domain.ErrInvalidRequest
	}

	// (2) Parse the verified body.
	var wh domain.QRWebhook
	if err := json.Unmarshal(body, &wh); err != nil {
		return nil, domain.ErrInvalidRequest
	}
	// A webhook must carry at least one correlation handle: the PSP provider_ref
	// (card/cross-border rails) or the reference we embedded in the PromptPay QR.
	if wh.ProviderRef == "" && wh.Reference == "" {
		return nil, domain.ErrInvalidRequest
	}

	// (2) Correlate the webhook back to the minted QR.
	//   * provider_ref path: card-scheme / cross-border, where the PSP returned a
	//     provider_ref at mint time.
	//   * correlation_ref path: PromptPay, where the bank echoes the reference we
	//     embedded in the EMVCo tag-62 payload.
	// provider_ref is tried first so card rails keep their existing behaviour.
	row, err := s.correlateWebhook(ctx, wh)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	// The idempotency / provider handle we log and echo downstream.
	ref := wh.ProviderRef
	if ref == "" {
		ref = wh.Reference
	}

	// (3) Idempotency: a replayed confirmation for an already-paid QR is a no-op.
	if row.Status == string(domain.QRPaid) {
		return toDomainQR(row), nil
	}
	// Only an awaiting_payment QR can be confirmed.
	if row.Status != string(domain.QRAwaitingPayment) {
		return nil, domain.ErrInvalidState
	}

	// (4) Assert the reported amount/currency match the minted QR.
	whMinor, err := money.ToMinor(wh.Amount, row.Currency)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	if whMinor != row.AmountMinor || wh.Currency != row.Currency {
		s.log.Warn().Str("provider_ref", ref).Msg("qr webhook amount/currency mismatch")
		return nil, domain.ErrInvalidRequest
	}

	// A failed confirmation transitions the QR to failed without a webhook.
	if wh.Status != string(domain.QRPaid) {
		updated, err := s.repo.UpdateQRStatus(ctx, repository.UpdateQRStatusParams{
			ID:          row.ID,
			Status:      string(domain.QRFailed),
			PayerBank:   strPtr(wh.PayerBank),
			ProviderRef: row.ProviderRef,
			PaidAt:      row.PaidAt,
		})
		if err != nil {
			return nil, err
		}
		return toDomainQR(updated), nil
	}

	paidAt := wh.PaidAt
	if paidAt.IsZero() {
		paidAt = time.Now()
	}

	// (4)+(5) Transition -> paid and enqueue the outbound merchant webhook in one
	// transaction so the state change and the notification are atomic.
	var result *domain.QRPayment
	err = s.withTx(ctx, func(q repository.Querier) error {
		updated, err := q.UpdateQRStatus(ctx, repository.UpdateQRStatusParams{
			ID:          row.ID,
			Status:      string(domain.QRPaid),
			PayerBank:   strPtr(wh.PayerBank),
			ProviderRef: row.ProviderRef,
			PaidAt:      pgTimestamptz(paidAt),
		})
		if err != nil {
			return err
		}
		payload, err := json.Marshal(map[string]any{
			"event":         "payment.paid",
			"qr_payment_id": pgUUIDString(updated.ID),
			"reference":     updated.Reference,
			"provider_ref":  ref,
			"amount_minor":  updated.AmountMinor,
			"currency":      updated.Currency,
			"payer_bank":    wh.PayerBank,
			"paid_at":       paidAt,
		})
		if err != nil {
			return err
		}
		if err := q.CreateWebhookEvent(ctx, repository.CreateWebhookEventParams{
			ID:         toPgUUID(uuid.New()),
			MerchantID: updated.MerchantID,
			EventType:  "payment.paid",
			Payload:    payload,
		}); err != nil {
			return err
		}
		result = toDomainQR(updated)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// correlateWebhook resolves the qr_payment row an inbound webhook refers to.
// It prefers the PSP provider_ref (card / cross-border rails) and falls back to
// the correlation_ref echoed back in a PromptPay webhook's `reference` field.
// Returns pgx.ErrNoRows when neither handle matches a known QR.
func (s *qrService) correlateWebhook(ctx context.Context, wh domain.QRWebhook) (repository.QrPayment, error) {
	if wh.ProviderRef != "" {
		pref := wh.ProviderRef
		row, err := s.repo.GetQRByProviderRef(ctx, &pref)
		if err == nil {
			return row, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return repository.QrPayment{}, err
		}
		// fall through: a PromptPay webhook may still carry a provider_ref (the
		// bank txn id) that never matches, but a correlatable reference.
	}
	if wh.Reference != "" {
		cref := wh.Reference
		return s.repo.GetQRByCorrelationRef(ctx, &cref)
	}
	return repository.QrPayment{}, pgx.ErrNoRows
}

// newPromptPayRef mints a stable correlation reference embedded in a PromptPay
// EMVCo tag-62 payload. When the merchant supplied a reference it is used as-is
// (the merchant owns idempotency); otherwise a compact, QR-safe reference is
// derived from the payment id. The result is <=25 chars so it fits the EMVCo
// tag-62 sub-field length limit.
func newPromptPayRef(merchantRef string, id uuid.UUID) string {
	// correlation_ref has a UNIQUE index and is what an inbound webhook is matched
	// against, so it MUST be unique per QR. Derive it from the QR's own id rather
	// than the (possibly reused) merchant reference — otherwise two QR with the
	// same merchant reference collide. The human-facing merchant reference is
	// still stored separately on the row (qp.Reference). 20 hex chars is tag-62
	// safe (EMVCo additional-data length limits).
	_ = merchantRef
	return "PP" + strings.ReplaceAll(id.String(), "-", "")[:20]
}

// ---- helpers ---------------------------------------------------------------

// withTx runs fn inside a database transaction when a TxManager is available,
// committing on success and rolling back on error. When no TxManager is wired
// (dev scaffolding) it runs fn directly against the base Querier.
func (s *qrService) withTx(ctx context.Context, fn func(q repository.Querier) error) error {
	if s.txm == nil {
		return fn(s.repo)
	}
	tx, err := s.txm.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := repository.New(tx)
	if err := fn(q); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// toDomainQR maps a repository row to the domain QR aggregate.
func toDomainQR(r repository.QrPayment) *domain.QRPayment {
	amount, _ := money.FromMinor(r.AmountMinor, r.Currency)
	qp := &domain.QRPayment{
		ID:             pgUUIDToUUID(r.ID),
		MerchantID:     pgUUIDToUUID(r.MerchantID),
		Method:         domain.QRMethod(r.Method),
		Amount:         amount,
		Currency:       r.Currency,
		Status:         domain.QRStatus(r.Status),
		Reference:      r.Reference,
		QRPayload:      r.QrPayload,
		QRImageURL:     ptrStr(r.QrImageUrl),
		PayerBank:      ptrStr(r.PayerBank),
		ProviderRef:    ptrStr(r.ProviderRef),
		CorrelationRef: ptrStr(r.CorrelationRef),
	}
	if r.ExpiresAt.Valid {
		qp.ExpiresAt = r.ExpiresAt.Time
	}
	if r.CreatedAt.Valid {
		qp.CreatedAt = r.CreatedAt.Time
	}
	if r.PaidAt.Valid {
		t := r.PaidAt.Time
		qp.PaidAt = &t
	}
	return qp
}

// pgTimestamptz wraps a time.Time as a valid pgtype.Timestamptz.
func pgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
