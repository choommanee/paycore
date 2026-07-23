package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/service"
)

// SandboxHandler exposes the PUBLIC sandbox payer-simulator endpoints. It is
// mounted ONLY when SANDBOX_MODE=true (see router.Setup); when the flag is off
// these routes are completely absent (404). Every endpoint is unauthenticated
// by design — the simulator is a browser with no merchant key — and the server
// signs the bank confirmation internally so no secret is exposed.
type SandboxHandler struct {
	svc service.SandboxService
	log zerolog.Logger
}

func NewSandboxHandler(svc service.SandboxService, log zerolog.Logger) *SandboxHandler {
	return &SandboxHandler{svc: svc, log: log}
}

// View godoc
// @Summary Sandbox: public display info for a QR payment (simulator)
// @Router  /v1/sandbox/qr-payments/{id} [get]
func (h *SandboxHandler) View(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid id")
	}
	view, err := h.svc.View(c.Context(), id)
	if err != nil {
		return err
	}
	return domain.Success(c, view)
}

// Pay godoc
// @Summary Sandbox: simulate a successful bank payment for a QR
// @Router  /v1/sandbox/qr-payments/{id}/pay [post]
func (h *SandboxHandler) Pay(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid id")
	}
	qp, err := h.svc.Pay(c.Context(), id)
	if err != nil {
		return err
	}
	return domain.Success(c, qp)
}

// Decline godoc
// @Summary Sandbox: simulate a failed/declined bank payment for a QR
// @Router  /v1/sandbox/qr-payments/{id}/decline [post]
func (h *SandboxHandler) Decline(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid id")
	}
	qp, err := h.svc.Decline(c.Context(), id)
	if err != nil {
		return err
	}
	return domain.Success(c, qp)
}

// List godoc
// @Summary Sandbox: recent pending QR payments (incoming requests list)
// @Router  /v1/sandbox/qr-payments [get]
func (h *SandboxHandler) List(c *fiber.Ctx) error {
	status := c.Query("status", string(domain.QRAwaitingPayment))
	// limit is clamped to [1,100] before any conversion, so the int32 cast can
	// never overflow (the service also re-clamps as defence in depth).
	limit := int32(20)
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			if n > 100 {
				n = 100
			}
			limit = int32(n) // #nosec G109,G115 -- n is clamped to [1,100] above so the cast cannot overflow
		}
	}
	items, err := h.svc.ListPending(c.Context(), status, limit)
	if err != nil {
		return err
	}
	return domain.Success(c, items)
}
