package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/service"
)

// CheckoutHandler serves the PUBLIC hosted-checkout endpoints. There is NO auth
// middleware on these routes: the opaque session token in the URL path is the
// credential, and the service scopes everything to the resolved session.
type CheckoutHandler struct {
	svc      service.CheckoutService
	validate *validator.Validate
	log      zerolog.Logger
}

func NewCheckoutHandler(svc service.CheckoutService, log zerolog.Logger) *CheckoutHandler {
	return &CheckoutHandler{svc: svc, validate: validator.New(), log: log}
}

// Create opens a checkout session from a payment link's public id.
// @Router /v1/checkout/sessions [post]
func (h *CheckoutHandler) Create(c *fiber.Ctx) error {
	var req domain.CheckoutSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return validationErrorResponse(c, err)
	}
	view, err := h.svc.CreateFromLink(c.Context(), req.Link)
	if err != nil {
		return err
	}
	return domain.Created(c, view)
}

// Get returns the current session state (page load / polling).
// @Router /v1/checkout/sessions/{token} [get]
func (h *CheckoutHandler) Get(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_TOKEN", "missing session token")
	}
	view, err := h.svc.Get(c.Context(), token)
	if err != nil {
		return err
	}
	return domain.Success(c, view)
}

// Pay initiates payment with the selected method.
// @Router /v1/checkout/sessions/{token}/pay [post]
func (h *CheckoutHandler) Pay(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_TOKEN", "missing session token")
	}
	var req domain.CheckoutPayRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return validationErrorResponse(c, err)
	}
	view, err := h.svc.Pay(c.Context(), token, req)
	if err != nil {
		return err
	}
	return domain.Success(c, view)
}

// ConfirmMock simulates a wallet approve/decline for a session awaiting action.
// PUBLIC + SANDBOX ONLY — the router mounts this route only when sandbox is on.
// @Router /v1/checkout/sessions/{token}/confirm-mock [post]
func (h *CheckoutHandler) ConfirmMock(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_TOKEN", "missing session token")
	}
	var req domain.CheckoutConfirmMockRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	view, err := h.svc.ConfirmMock(c.Context(), token, req.Approve)
	if err != nil {
		return err
	}
	return domain.Success(c, view)
}
