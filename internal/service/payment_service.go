package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/external"
	"github.com/yourco/payment-gateway/internal/pkg/money"
	"github.com/yourco/payment-gateway/internal/repository"
)

// PaymentService is the core business logic boundary. Every by-id operation is
// scoped to the authenticated merchant so one merchant can never Get/Capture/
// Refund/Void or resume 3DS on another merchant's payment (BOLA/IDOR).
type PaymentService interface {
	Create(ctx context.Context, idemKey string, req domain.CreatePaymentRequest) (*domain.Payment, error)
	Capture(ctx context.Context, merchantID, id uuid.UUID, req domain.CaptureRequest) (*domain.Payment, error)
	Void(ctx context.Context, merchantID, id uuid.UUID) (*domain.Payment, error)
	Refund(ctx context.Context, merchantID, id uuid.UUID, req domain.RefundRequest) (*domain.Payment, error)
	Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.Payment, error)
	List(ctx context.Context, merchantID uuid.UUID, limit, offset int32) ([]*domain.Payment, error)
	HandleThreeDSResult(ctx context.Context, merchantID, id uuid.UUID, ok bool) (*domain.Payment, error)
	// VerifyThreeDSResult validates the ACS challenge result (CRes / signed
	// authentication value) for a payment. It returns true only when the CRes is
	// present and its signature verifies AND the network transStatus is "Y", so a
	// client cannot force authorization by POSTing transStatus=Y alone.
	VerifyThreeDSResult(paymentID uuid.UUID, cres, transStatus string) bool
}

// TxManager begins a database transaction. *pgxpool.Pool satisfies this, so the
// service can wrap multi-write operations (payment row + ledger entry + audit
// log) atomically. When nil (e.g. dev scaffolding without a DB) the service
// falls back to running queries directly against the injected Querier.
type TxManager interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Detokenizer resolves a single-use vault token back to a PAN inside PCI scope.
// crypto.Vault satisfies this. The plaintext PAN is used only for the acquirer
// call and is never persisted or logged.
type Detokenizer interface {
	Detokenize(ctx context.Context, token string) (string, error)
}

// StatusObserver records a payment reaching a given status. *metrics.Metrics
// satisfies this via ObservePaymentStatus. It may be nil (metrics disabled).
type StatusObserver interface {
	ObservePaymentStatus(status string)
}

// AuditSink accepts audit_log rows for asynchronous, off-hot-path persistence.
// *worker.AuditWorker satisfies this. When set, the authorize path enqueues its
// audit row here instead of INSERTing it inside the payment transaction, cutting
// one row-write per authorize from the synchronous tx. A nil sink keeps the
// in-transaction write (the safe default; used in tests and when disabled).
type AuditSink interface {
	Enqueue(p repository.WriteAuditLogParams) bool
}

type paymentService struct {
	repo     repository.Querier // sqlc-generated data access (nil in scaffold)
	txm      TxManager          // begins pgx transactions (nil in scaffold)
	vault    Detokenizer        // card-data vault (nil in scaffold)
	acquirer external.Acquirer  // upstream card-network switch / acquirer
	threeds  external.ThreeDS   // 3-D Secure authentication provider
	metrics  StatusObserver     // payment status counter (nil = disabled)
	audit    AuditSink          // async audit outbox (nil = write in-tx)
	// threeDSSecret signs/verifies the ACS challenge-result (CRes) value on the
	// browser 3DS return. Set via WithThreeDSSecret; empty disables CRes-based
	// finalization (the return then always fails-closed, never authorizing on a
	// bare transStatus=Y).
	threeDSSecret []byte
	log           zerolog.Logger
}

// NewPaymentService wires the payment service. txm and vault may be nil in local
// scaffolding; production passes a *pgxpool.Pool and a *crypto.Vault.
func NewPaymentService(repo repository.Querier, txm TxManager, vault Detokenizer, acq external.Acquirer, tds external.ThreeDS, log zerolog.Logger) PaymentService {
	return &paymentService{
		repo:     repo,
		txm:      txm,
		vault:    vault,
		acquirer: acq,
		threeds:  tds,
		log:      log.With().Str("service", "payment").Logger(),
	}
}

// WithMetrics attaches a status observer so payment lifecycle transitions are
// counted. Returns the service for chaining. Safe to skip (metrics disabled).
func (s *paymentService) WithMetrics(obs StatusObserver) PaymentService {
	s.metrics = obs
	return s
}

// WithThreeDSSecret sets the HMAC secret used to verify the ACS challenge-result
// (CRes) on the browser 3DS return. Returns the service for chaining.
func (s *paymentService) WithThreeDSSecret(secret string) PaymentService {
	s.threeDSSecret = []byte(secret)
	return s
}

// WithAuditSink attaches an async audit outbox so the authorize hot path writes
// its audit row off the synchronous transaction. Returns the service for
// chaining. A nil sink keeps the in-transaction audit write.
func (s *paymentService) WithAuditSink(sink AuditSink) PaymentService {
	s.audit = sink
	return s
}

// VerifyThreeDSResult validates the CRes / signed authentication value returned
// by the ACS. The expected value is HMAC-SHA256(secret, paymentID) — a stand-in
// for the issuer-signed CRes signature the ACS would return. Verification is a
// constant-time compare and requires transStatus == "Y". When no secret is
// configured it fails closed (returns false) so the endpoint can never authorize
// on an unverified, client-supplied transStatus alone.
func (s *paymentService) VerifyThreeDSResult(paymentID uuid.UUID, cres, transStatus string) bool {
	if transStatus != "Y" || len(s.threeDSSecret) == 0 || cres == "" {
		return false
	}
	mac := hmac.New(sha256.New, s.threeDSSecret)
	mac.Write(paymentID[:])
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(cres), []byte(expected))
}

// observeStatus records a payment reaching status when a metrics observer is
// wired. A nil observer is a no-op.
func (s *paymentService) observeStatus(status domain.PaymentStatus) {
	if s.metrics != nil {
		s.metrics.ObservePaymentStatus(string(status))
	}
}

// riskDecision is the output of the fraud/risk hook.
type riskDecision struct {
	Approve bool
	Reason  string
}

// scoreRisk is the extension point for fraud / risk scoring. Production plugs in
// a rules engine or ML model here (velocity checks, BIN risk, device
// fingerprint, etc.). The stub approves everything so the happy path is
// exercised end to end; it must never receive or log the PAN.
func (s *paymentService) scoreRisk(_ context.Context, _ domain.CreatePaymentRequest) riskDecision {
	return riskDecision{Approve: true}
}

// Create authorizes (and optionally captures) a payment. See the interface doc
// for the seven-step production flow.
func (s *paymentService) Create(ctx context.Context, idemKey string, req domain.CreatePaymentRequest) (*domain.Payment, error) {
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

	// (1) Idempotency: a replay of the same (merchant, key) returns the stored
	// result instead of charging the card twice.
	//
	// This must guard against a *concurrent* duplicate, not just a sequential
	// replay: a plain read-then-write would let two simultaneous requests both
	// miss the read and both call the acquirer (a real double charge) before the
	// unique constraint trips at insert time. To close that window we reserve the
	// payment row (via a unique (merchant_id, idempotency_key) insert) BEFORE the
	// acquirer is ever contacted. If the reservation fails on the unique
	// constraint, another request already owns this key, so we return that stored
	// payment and never touch the card.
	if idemKey != "" {
		existing, err := s.repo.GetByIdempotencyKey(ctx, repository.GetByIdempotencyKeyParams{
			MerchantID:     toPgUUID(req.MerchantID),
			IdempotencyKey: idemKey,
		})
		if err == nil {
			// A replay must carry the SAME request payload. If the fingerprint
			// differs (e.g. same key, different amount) reject with 409 instead of
			// silently returning the stored payment and dropping the new charge.
			if !sameRequestFingerprint(existing, req, amountMinor) {
				return nil, domain.ErrIdempotencyKeyReuse
			}
			return toDomainPayment(existing), nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}

	// (1b) Reserve the idempotency key by inserting a pending payment row up
	// front. A concurrent duplicate loses the race on the unique constraint and
	// short-circuits here — before the acquirer is called — eliminating the
	// double-charge window. The reserved row is later transitioned in place
	// (authorized/failed/requires_action) rather than a second row being
	// inserted.
	reservedID := uuid.New()
	if idemKey != "" {
		reserved, err := s.reserve(ctx, reservedID, idemKey, req, amountMinor)
		if err != nil {
			if isUniqueViolation(err) {
				// Another request already reserved this (merchant, key). Return its
				// stored payment; the card is charged (at most) once.
				existing, gerr := s.repo.GetByIdempotencyKey(ctx, repository.GetByIdempotencyKeyParams{
					MerchantID:     toPgUUID(req.MerchantID),
					IdempotencyKey: idemKey,
				})
				if gerr != nil {
					return nil, domain.ErrDuplicateRequest
				}
				if !sameRequestFingerprint(existing, req, amountMinor) {
					return nil, domain.ErrIdempotencyKeyReuse
				}
				return toDomainPayment(existing), nil
			}
			return nil, err
		}
		_ = reserved
	}

	// (2) Detokenize the vault token -> PAN. This is the only place the PAN
	// exists in memory; it is handed straight to the acquirer and never stored
	// or logged.
	pan, err := s.detokenize(ctx, req.PaymentToken)
	if err != nil {
		s.log.Error().Err(err).Msg("detokenize failed")
		return nil, domain.ErrInvalidRequest
	}

	// (3) Risk / fraud hook (extension point).
	if decision := s.scoreRisk(ctx, req); !decision.Approve {
		s.log.Warn().Str("reason", decision.Reason).Msg("risk declined")
		return nil, domain.ErrCardDeclined
	}

	// (4) 3-D Secure. When the issuer requires a challenge we persist the
	// payment as requires_action with a next_action_url and return; the shopper
	// completes the challenge and HandleThreeDSResult resumes authorization.
	var cryptogram string
	if s.threeds != nil {
		challengeURL, crypto, err := s.threeds.Authenticate(ctx, pan, req.Amount, req.Currency, req.ReturnURL)
		if err != nil {
			return nil, domain.ErrCardDeclined
		}
		if challengeURL != "" {
			return s.persistNew(ctx, reservedID, idemKey, req, amountMinor, persistOpts{
				status:        domain.StatusRequiresAction,
				brand:         brandFor(pan),
				last4:         panLast4(pan),
				nextActionURL: challengeURL,
				action:        "payment.requires_action",
			})
		}
		cryptogram = crypto
	}

	// (5) Authorization request to the acquirer.
	auth, err := s.acquirer.Authorize(ctx, external.AuthRequest{
		PAN:               pan,
		Amount:            req.Amount,
		Currency:          req.Currency,
		Merchant:          req.MerchantID.String(),
		ThreeDSCryptogram: cryptogram,
	})
	if err != nil {
		return nil, err
	}
	if !auth.Approved {
		if auth.DeclineCode == external.Decline3DSRequired {
			return nil, domain.ErrThreeDSRequired
		}
		// (6) Persist the failed attempt so it is auditable.
		p, perr := s.persistNew(ctx, reservedID, idemKey, req, amountMinor, persistOpts{
			status: domain.StatusFailed,
			brand:  auth.CardBrand,
			last4:  auth.Last4,
			action: "payment.failed",
		})
		if perr != nil {
			return nil, perr
		}
		_ = p
		if auth.DeclineCode == external.DeclineInsufficientFunds {
			return nil, domain.ErrInsufficientFund
		}
		return nil, domain.ErrCardDeclined
	}

	// (6) Persist the authorized payment + authorize ledger entry + audit row in
	// one transaction.
	p, err := s.persistNew(ctx, reservedID, idemKey, req, amountMinor, persistOpts{
		status:      domain.StatusAuthorized,
		brand:       auth.CardBrand,
		last4:       auth.Last4,
		acquirerRef: auth.AcquirerRef,
		authCode:    auth.AuthCode,
		ledgerType:  "authorize",
		ledgerMinor: amountMinor,
		action:      "payment.authorized",
	})
	if err != nil {
		return nil, err
	}

	// (7) Immediate capture when requested (charge flow). The merchant scope is
	// the request's authenticated merchant.
	if req.Capture {
		return s.Capture(ctx, req.MerchantID, p.ID, domain.CaptureRequest{Amount: req.Amount})
	}
	return p, nil
}

// persistOpts describes a single persisted payment write (row + optional ledger
// entry + audit row).
type persistOpts struct {
	status        domain.PaymentStatus
	brand         string
	last4         string
	acquirerRef   string
	authCode      string
	nextActionURL string
	ledgerType    string // empty = no ledger entry
	ledgerMinor   int64
	action        string
}

// reserve inserts a pending payment row that claims the (merchant_id,
// idempotency_key) unique constraint before the acquirer is contacted. A
// concurrent duplicate loses this insert with a unique-violation, which the
// caller uses to short-circuit without charging the card twice. The reserved
// row is later transitioned in place by persistNew.
func (s *paymentService) reserve(ctx context.Context, id uuid.UUID, idemKey string, req domain.CreatePaymentRequest, amountMinor int64) (repository.Payment, error) {
	return s.repo.CreatePayment(ctx, repository.CreatePaymentParams{
		ID:             toPgUUID(id),
		MerchantID:     toPgUUID(req.MerchantID),
		AmountMinor:    amountMinor,
		Currency:       req.Currency,
		Status:         string(domain.StatusRequiresAction), // pending until resolved
		Reference:      strPtr(req.Reference),
		IdempotencyKey: idemKey,
	})
}

// persistNew resolves a payment aggregate to its final create-time state,
// together with its ledger entry and audit log row, atomically. When id refers
// to a pre-inserted reservation row (the idempotent path) it transitions that
// row in place; otherwise it inserts a fresh row. Either way, exactly one
// payment row exists for a given (merchant, idempotency key).
func (s *paymentService) persistNew(ctx context.Context, id uuid.UUID, idemKey string, req domain.CreatePaymentRequest, amountMinor int64, opts persistOpts) (*domain.Payment, error) {
	var result *domain.Payment
	err := s.withTx(ctx, func(q repository.Querier) error {
		var row repository.Payment
		var err error
		if idemKey != "" {
			// A reservation row already exists (inserted before the acquirer call);
			// transition it in place rather than inserting a duplicate.
			row, err = q.UpdatePaymentStatus(ctx, repository.UpdatePaymentStatusParams{
				ID:                  toPgUUID(id),
				Status:              string(opts.status),
				CapturedAmountMinor: 0,
				RefundedAmountMinor: 0,
				AcquirerRef:         strPtr(opts.acquirerRef),
				AuthCode:            strPtr(opts.authCode),
			})
			if err != nil {
				return err
			}
		} else {
			row, err = q.CreatePayment(ctx, repository.CreatePaymentParams{
				ID:             toPgUUID(id),
				MerchantID:     toPgUUID(req.MerchantID),
				AmountMinor:    amountMinor,
				Currency:       req.Currency,
				Status:         string(opts.status),
				CardBrand:      strPtr(opts.brand),
				CardLast4:      strPtr(opts.last4),
				Reference:      strPtr(req.Reference),
				IdempotencyKey: idemKey,
			})
			if err != nil {
				return err
			}
			// Some fields (acquirer ref, auth code) are set via an update so the row
			// reflects the acquirer response; captured/refunded stay at zero here.
			if opts.acquirerRef != "" || opts.authCode != "" {
				row, err = q.UpdatePaymentStatus(ctx, repository.UpdatePaymentStatusParams{
					ID:                  row.ID,
					Status:              string(opts.status),
					CapturedAmountMinor: row.CapturedAmountMinor,
					RefundedAmountMinor: row.RefundedAmountMinor,
					AcquirerRef:         strPtr(opts.acquirerRef),
					AuthCode:            strPtr(opts.authCode),
				})
				if err != nil {
					return err
				}
			}
		}
		if opts.ledgerType != "" {
			if err := q.InsertLedgerEntry(ctx, repository.InsertLedgerEntryParams{
				ID:          toPgUUID(uuid.New()),
				PaymentID:   row.ID,
				EntryType:   opts.ledgerType,
				AmountMinor: opts.ledgerMinor,
				Currency:    req.Currency,
			}); err != nil {
				return err
			}
		}
		auditMeta := map[string]any{
			"status":       string(opts.status),
			"amount_minor": amountMinor,
			"currency":     req.Currency,
		}
		if s.audit != nil {
			// Enqueue the audit row for async persistence instead of writing it
			// inside this tx — removes one row-write from the authorize hot path.
			ap, aerr := auditParams(req.MerchantID.String(), opts.action, id, auditMeta)
			if aerr != nil {
				return aerr
			}
			s.audit.Enqueue(ap)
		} else if err := writeAudit(ctx, q, req.MerchantID.String(), opts.action, id, auditMeta); err != nil {
			return err
		}
		p := toDomainPayment(row)
		p.NextActionURL = opts.nextActionURL
		result = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.observeStatus(opts.status)
	return result, nil
}

// Capture settles a previously authorized payment (full or partial). Scoped to
// the authenticated merchant.
func (s *paymentService) Capture(ctx context.Context, merchantID, id uuid.UUID, req domain.CaptureRequest) (*domain.Payment, error) {
	row, err := s.getRow(ctx, merchantID, id)
	if err != nil {
		return nil, err
	}
	if err := assertTransition(domain.PaymentStatus(row.Status), domain.StatusCaptured); err != nil {
		return nil, err
	}
	captureMinor, err := money.ToMinor(req.Amount, row.Currency)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	if captureMinor <= 0 || captureMinor > row.AmountMinor {
		return nil, domain.ErrInvalidRequest
	}

	res, err := s.acquirer.Capture(ctx, ptrStr(row.AcquirerRef), req.Amount)
	if err != nil {
		return nil, err
	}
	if !res.Approved {
		return nil, domain.ErrCardDeclined
	}

	var result *domain.Payment
	err = s.withTx(ctx, func(q repository.Querier) error {
		updated, err := q.UpdatePaymentStatus(ctx, repository.UpdatePaymentStatusParams{
			ID:                  row.ID,
			Status:              string(domain.StatusCaptured),
			CapturedAmountMinor: captureMinor,
			RefundedAmountMinor: row.RefundedAmountMinor,
			AcquirerRef:         row.AcquirerRef,
			AuthCode:            row.AuthCode,
		})
		if err != nil {
			return err
		}
		if err := q.InsertLedgerEntry(ctx, repository.InsertLedgerEntryParams{
			ID:          toPgUUID(uuid.New()),
			PaymentID:   row.ID,
			EntryType:   "capture",
			AmountMinor: captureMinor,
			Currency:    row.Currency,
		}); err != nil {
			return err
		}
		if err := writeAudit(ctx, q, ptrStr(nil), "payment.captured", id, map[string]any{
			"captured_amount_minor": captureMinor,
			"currency":              row.Currency,
		}); err != nil {
			return err
		}
		result = toDomainPayment(updated)
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.observeStatus(domain.StatusCaptured)
	return result, nil
}

// Void reverses an authorization before it is captured. Scoped to the
// authenticated merchant.
func (s *paymentService) Void(ctx context.Context, merchantID, id uuid.UUID) (*domain.Payment, error) {
	row, err := s.getRow(ctx, merchantID, id)
	if err != nil {
		return nil, err
	}
	if err := assertTransition(domain.PaymentStatus(row.Status), domain.StatusVoided); err != nil {
		return nil, err
	}

	res, err := s.acquirer.Void(ctx, ptrStr(row.AcquirerRef))
	if err != nil {
		return nil, err
	}
	if !res.Approved {
		return nil, domain.ErrCardDeclined
	}

	var result *domain.Payment
	err = s.withTx(ctx, func(q repository.Querier) error {
		updated, err := q.UpdatePaymentStatus(ctx, repository.UpdatePaymentStatusParams{
			ID:                  row.ID,
			Status:              string(domain.StatusVoided),
			CapturedAmountMinor: row.CapturedAmountMinor,
			RefundedAmountMinor: row.RefundedAmountMinor,
			AcquirerRef:         row.AcquirerRef,
			AuthCode:            row.AuthCode,
		})
		if err != nil {
			return err
		}
		if err := q.InsertLedgerEntry(ctx, repository.InsertLedgerEntryParams{
			ID:          toPgUUID(uuid.New()),
			PaymentID:   row.ID,
			EntryType:   "void",
			AmountMinor: row.AmountMinor,
			Currency:    row.Currency,
		}); err != nil {
			return err
		}
		if err := writeAudit(ctx, q, ptrStr(nil), "payment.voided", id, map[string]any{
			"amount_minor": row.AmountMinor,
			"currency":     row.Currency,
		}); err != nil {
			return err
		}
		result = toDomainPayment(updated)
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.observeStatus(domain.StatusVoided)
	return result, nil
}

// Refund returns funds against a captured payment (partial or full). The running
// refunded total is guarded to never exceed the captured amount.
func (s *paymentService) Refund(ctx context.Context, merchantID, id uuid.UUID, req domain.RefundRequest) (*domain.Payment, error) {
	row, err := s.getRow(ctx, merchantID, id)
	if err != nil {
		return nil, err
	}
	status := domain.PaymentStatus(row.Status)
	if status != domain.StatusCaptured && status != domain.StatusPartialRefunded {
		return nil, domain.ErrInvalidState
	}
	refundMinor, err := money.ToMinor(req.Amount, row.Currency)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	if refundMinor <= 0 {
		return nil, domain.ErrInvalidRequest
	}

	newRefunded := row.RefundedAmountMinor + refundMinor
	if newRefunded > row.CapturedAmountMinor {
		return nil, domain.ErrRefundExceeds
	}

	next := domain.StatusPartialRefunded
	if newRefunded == row.CapturedAmountMinor {
		next = domain.StatusRefunded
	}
	if err := assertTransition(status, next); err != nil {
		return nil, err
	}

	res, err := s.acquirer.Refund(ctx, ptrStr(row.AcquirerRef), req.Amount)
	if err != nil {
		return nil, err
	}
	if !res.Approved {
		return nil, domain.ErrCardDeclined
	}

	var result *domain.Payment
	err = s.withTx(ctx, func(q repository.Querier) error {
		updated, err := q.UpdatePaymentStatus(ctx, repository.UpdatePaymentStatusParams{
			ID:                  row.ID,
			Status:              string(next),
			CapturedAmountMinor: row.CapturedAmountMinor,
			RefundedAmountMinor: newRefunded,
			AcquirerRef:         row.AcquirerRef,
			AuthCode:            row.AuthCode,
		})
		if err != nil {
			return err
		}
		if err := q.InsertLedgerEntry(ctx, repository.InsertLedgerEntryParams{
			ID:          toPgUUID(uuid.New()),
			PaymentID:   row.ID,
			EntryType:   "refund",
			AmountMinor: refundMinor,
			Currency:    row.Currency,
		}); err != nil {
			return err
		}
		if err := writeAudit(ctx, q, ptrStr(nil), "payment.refunded", id, map[string]any{
			"refund_amount_minor":  refundMinor,
			"refunded_total_minor": newRefunded,
			"currency":             row.Currency,
			"reason":               req.Reason,
		}); err != nil {
			return err
		}
		result = toDomainPayment(updated)
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.observeStatus(next)
	return result, nil
}

// Get returns a payment by id, scoped to the authenticated merchant.
func (s *paymentService) Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.Payment, error) {
	row, err := s.getRow(ctx, merchantID, id)
	if err != nil {
		return nil, err
	}
	return toDomainPayment(row), nil
}

// List returns a merchant's payments, most recent first, paginated. The result
// is always scoped to the given merchant id (the authenticated merchant).
func (s *paymentService) List(ctx context.Context, merchantID uuid.UUID, limit, offset int32) ([]*domain.Payment, error) {
	rows, err := s.repo.ListPaymentsByMerchant(ctx, repository.ListPaymentsByMerchantParams{
		MerchantID: toPgUUID(merchantID),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Payment, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDomainPayment(r))
	}
	return out, nil
}

// HandleThreeDSResult resumes a requires_action payment once the shopper
// completes (or fails) the 3-D Secure challenge. On success it authorizes with
// the acquirer; on failure the payment transitions to failed.
func (s *paymentService) HandleThreeDSResult(ctx context.Context, merchantID, id uuid.UUID, ok bool) (*domain.Payment, error) {
	row, err := s.getRow(ctx, merchantID, id)
	if err != nil {
		return nil, err
	}
	current := domain.PaymentStatus(row.Status)
	if current != domain.StatusRequiresAction {
		return nil, domain.ErrInvalidState
	}

	if !ok {
		return s.finalizeThreeDS(ctx, row, domain.StatusFailed, "", "", "", "payment.failed")
	}

	// Challenge succeeded: re-authorize with the acquirer. The cryptogram is
	// re-derived from the vault token so the PAN is never persisted.
	pan, err := s.detokenizeFor3DS(ctx, row)
	if err != nil {
		return nil, domain.ErrCardDeclined
	}
	var cryptogram string
	if s.threeds != nil {
		amount, _ := money.FromMinor(row.AmountMinor, row.Currency)
		_, cryptogram, err = s.threeds.Authenticate(ctx, pan, amount, row.Currency, "")
		if err != nil {
			return nil, domain.ErrCardDeclined
		}
	}
	amount, _ := money.FromMinor(row.AmountMinor, row.Currency)
	auth, err := s.acquirer.Authorize(ctx, external.AuthRequest{
		PAN:               pan,
		Amount:            amount,
		Currency:          row.Currency,
		Merchant:          pgUUIDString(row.MerchantID),
		ThreeDSCryptogram: cryptogram,
	})
	if err != nil {
		return nil, err
	}
	if !auth.Approved {
		return s.finalizeThreeDS(ctx, row, domain.StatusFailed, "", "", "", "payment.failed")
	}
	return s.finalizeThreeDS(ctx, row, domain.StatusAuthorized, auth.AcquirerRef, auth.AuthCode, "authorize", "payment.authorized")
}

// finalizeThreeDS transitions a requires_action payment to its resolved state,
// writing the ledger entry (for authorize) and audit row atomically.
func (s *paymentService) finalizeThreeDS(ctx context.Context, row repository.Payment, status domain.PaymentStatus, acquirerRef, authCode, ledgerType, action string) (*domain.Payment, error) {
	if err := assertTransition(domain.StatusRequiresAction, status); err != nil {
		return nil, err
	}
	var result *domain.Payment
	err := s.withTx(ctx, func(q repository.Querier) error {
		acqRef := row.AcquirerRef
		ac := row.AuthCode
		if acquirerRef != "" {
			acqRef = strPtr(acquirerRef)
		}
		if authCode != "" {
			ac = strPtr(authCode)
		}
		updated, err := q.UpdatePaymentStatus(ctx, repository.UpdatePaymentStatusParams{
			ID:                  row.ID,
			Status:              string(status),
			CapturedAmountMinor: row.CapturedAmountMinor,
			RefundedAmountMinor: row.RefundedAmountMinor,
			AcquirerRef:         acqRef,
			AuthCode:            ac,
		})
		if err != nil {
			return err
		}
		if ledgerType != "" {
			if err := q.InsertLedgerEntry(ctx, repository.InsertLedgerEntryParams{
				ID:          toPgUUID(uuid.New()),
				PaymentID:   row.ID,
				EntryType:   ledgerType,
				AmountMinor: row.AmountMinor,
				Currency:    row.Currency,
			}); err != nil {
				return err
			}
		}
		if err := writeAudit(ctx, q, pgUUIDString(row.MerchantID), action, pgUUIDToUUID(row.ID), map[string]any{
			"status": string(status),
		}); err != nil {
			return err
		}
		result = toDomainPayment(updated)
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.observeStatus(status)
	return result, nil
}

// detokenizeFor3DS resolves the PAN needed to re-authorize after a 3DS
// challenge. The token is not persisted on the payment row (PCI scope), so a
// production implementation would retrieve it from the segmented vault keyed by
// payment id. The stub returns an error when no vault is wired.
func (s *paymentService) detokenizeFor3DS(ctx context.Context, _ repository.Payment) (string, error) {
	// The single-use token is not stored on the operational payment row. Real
	// deployments look it up in the PCI-scoped vault by payment id; here there is
	// nothing to detokenize, so signal a decline rather than fabricate card data.
	_ = ctx
	return "", errors.New("payment token unavailable for re-authorization")
}

// ---- helpers ---------------------------------------------------------------

// withTx runs fn inside a database transaction when a TxManager is available,
// committing on success and rolling back on error. When no TxManager is wired
// (dev scaffolding) it runs fn directly against the base Querier.
func (s *paymentService) withTx(ctx context.Context, fn func(q repository.Querier) error) error {
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

// getRow fetches a payment scoped to the authenticated merchant. A row that
// belongs to another merchant returns ErrPaymentNotFound (404) rather than a
// 403, so the API is not a cross-tenant existence oracle.
func (s *paymentService) getRow(ctx context.Context, merchantID, id uuid.UUID) (repository.Payment, error) {
	row, err := s.repo.GetPaymentForMerchant(ctx, repository.GetPaymentForMerchantParams{
		ID:         toPgUUID(id),
		MerchantID: toPgUUID(merchantID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Payment{}, domain.ErrPaymentNotFound
		}
		return repository.Payment{}, err
	}
	return row, nil
}

func (s *paymentService) detokenize(ctx context.Context, token string) (string, error) {
	if s.vault == nil {
		return "", errors.New("vault not configured")
	}
	return s.vault.Detokenize(ctx, token)
}

// auditParams builds an immutable audit_log row. metadata is JSON-encoded; it
// must never contain card data.
func auditParams(actor, action string, entityID uuid.UUID, metadata map[string]any) (repository.WriteAuditLogParams, error) {
	if actor == "" {
		actor = "system"
	}
	var raw []byte
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err != nil {
			return repository.WriteAuditLogParams{}, err
		}
		raw = b
	}
	return repository.WriteAuditLogParams{
		ID:         toPgUUID(uuid.New()),
		Actor:      actor,
		Action:     action,
		EntityType: "payment",
		EntityID:   entityID.String(),
		Metadata:   raw,
	}, nil
}

// writeAudit inserts an immutable audit_log row synchronously via q.
func writeAudit(ctx context.Context, q repository.Querier, actor, action string, entityID uuid.UUID, metadata map[string]any) error {
	p, err := auditParams(actor, action, entityID, metadata)
	if err != nil {
		return err
	}
	return q.WriteAuditLog(ctx, p)
}

// assertTransition enforces the payment state machine. Invalid transitions are
// rejected with ErrInvalidState.
//
//	requires_action -> authorized | failed
//	authorized      -> captured | voided
//	captured        -> partial_refunded | refunded
//	partial_refunded-> partial_refunded | refunded
func assertTransition(from, to domain.PaymentStatus) error {
	allowed := map[domain.PaymentStatus]map[domain.PaymentStatus]bool{
		domain.StatusRequiresAction: {
			domain.StatusAuthorized: true,
			domain.StatusFailed:     true,
		},
		domain.StatusAuthorized: {
			domain.StatusCaptured: true,
			domain.StatusVoided:   true,
		},
		domain.StatusCaptured: {
			domain.StatusPartialRefunded: true,
			domain.StatusRefunded:        true,
		},
		domain.StatusPartialRefunded: {
			domain.StatusPartialRefunded: true,
			domain.StatusRefunded:        true,
		},
	}
	if allowed[from][to] {
		return nil
	}
	return domain.ErrInvalidState
}

// toDomainPayment maps a repository row to the domain aggregate.
func toDomainPayment(r repository.Payment) *domain.Payment {
	amount, _ := money.FromMinor(r.AmountMinor, r.Currency)
	captured, _ := money.FromMinor(r.CapturedAmountMinor, r.Currency)
	refunded, _ := money.FromMinor(r.RefundedAmountMinor, r.Currency)
	return &domain.Payment{
		ID:             pgUUIDToUUID(r.ID),
		MerchantID:     pgUUIDToUUID(r.MerchantID),
		Amount:         amount,
		CapturedAmount: captured,
		RefundedAmount: refunded,
		Currency:       r.Currency,
		Status:         domain.PaymentStatus(r.Status),
		CardBrand:      ptrStr(r.CardBrand),
		CardLast4:      ptrStr(r.CardLast4),
		AcquirerRef:    ptrStr(r.AcquirerRef),
		AuthCode:       ptrStr(r.AuthCode),
		Reference:      ptrStr(r.Reference),
		IdempotencyKey: r.IdempotencyKey,
		CreatedAt:      r.CreatedAt.Time,
		UpdatedAt:      r.UpdatedAt.Time,
	}
}

// panLast4 returns the last four digits of a PAN (never the full PAN).
func panLast4(pan string) string {
	if len(pan) <= 4 {
		return pan
	}
	return pan[len(pan)-4:]
}

// brandFor infers a card brand from the PAN prefix (coarse; only used until the
// acquirer echoes its own brand).
func brandFor(pan string) string {
	switch {
	case len(pan) > 0 && pan[0] == '4':
		return "visa"
	case len(pan) > 0 && (pan[0] == '5' || pan[0] == '2'):
		return "mastercard"
	case len(pan) >= 2 && pan[0] == '3' && (pan[1] == '4' || pan[1] == '7'):
		return "amex"
	default:
		return "unknown"
	}
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgUUIDToUUID(p pgtype.UUID) uuid.UUID {
	return uuid.UUID(p.Bytes)
}

func pgUUIDString(p pgtype.UUID) string {
	return pgUUIDToUUID(p).String()
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// isUniqueViolation reports whether err is a Postgres unique-constraint
// violation (SQLSTATE 23505). Used to detect a lost race on the idempotency-key
// reservation insert.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// sameRequestFingerprint reports whether a replayed Create with the given
// idempotency key carries the same money-moving payload as the already-stored
// payment. It compares the persisted, non-card fields (merchant, amount,
// currency, reference); the single-use payment token is never stored, so it is
// intentionally out of scope. A mismatch means the client reused the key with a
// different request, which must be rejected rather than silently deduped.
func sameRequestFingerprint(existing repository.Payment, req domain.CreatePaymentRequest, amountMinor int64) bool {
	if existing.AmountMinor != amountMinor {
		return false
	}
	if existing.Currency != req.Currency {
		return false
	}
	if pgUUIDToUUID(existing.MerchantID) != req.MerchantID {
		return false
	}
	if ptrStr(existing.Reference) != req.Reference {
		return false
	}
	return true
}

func validateAmount(amount decimal.Decimal, currency string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return domain.ErrInvalidRequest
	}
	_ = currency
	return nil
}
