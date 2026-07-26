package handler

import (
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/service"
)

// Handlers aggregates all HTTP handlers for dependency injection.
type Handlers struct {
	Payment  *PaymentHandler
	QR       *QRHandler
	Merchant *MerchantHandler
	Dispute  *DisputeHandler
	Admin    *AdminHandler
	Health   *HealthHandler
	// Sandbox is the public payer-simulator handler. It is nil unless
	// SANDBOX_MODE is enabled (set via WithSandbox in main); the router mounts the
	// sandbox routes only when it is non-nil.
	Sandbox *SandboxHandler
	// Auth serves dashboard human login (Google OIDC + session cookie). Attached
	// via WithAuth in main; the router mounts the auth routes only when non-nil.
	Auth *AuthHandler
	// PaymentLink serves shareable payment-link CRUD. Attached via
	// WithPaymentLinks in main; the router mounts the /payment-links routes only
	// when non-nil.
	PaymentLink *PaymentLinkHandler
}

// WithSandbox attaches the sandbox payer-simulator handler. Called from main
// ONLY when SANDBOX_MODE=true so the routes stay absent (404) otherwise.
func (h *Handlers) WithSandbox(svc service.SandboxService, log zerolog.Logger) *Handlers {
	h.Sandbox = NewSandboxHandler(svc, log)
	return h
}

// WithAuth attaches the dashboard auth handler (Google OIDC + session).
func (h *Handlers) WithAuth(a *AuthHandler) *Handlers {
	h.Auth = a
	return h
}

// WithPaymentLinks attaches the payment-link handler.
func (h *Handlers) WithPaymentLinks(p *PaymentLinkHandler) *Handlers {
	h.PaymentLink = p
	return h
}

func New(
	paySvc service.PaymentService,
	qrSvc service.QRService,
	merchantSvc service.MerchantService,
	disputeSvc service.DisputeService,
	adminSvc service.AdminService,
	db Pinger,
	log zerolog.Logger,
) *Handlers {
	return &Handlers{
		Payment:  NewPaymentHandler(paySvc, log),
		QR:       NewQRHandler(qrSvc, log),
		Merchant: NewMerchantHandler(merchantSvc, log),
		Dispute:  NewDisputeHandler(disputeSvc, log),
		Admin:    NewAdminHandler(adminSvc, log),
		Health:   NewHealthHandler(db),
	}
}
