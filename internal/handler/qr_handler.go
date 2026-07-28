package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/service"
)

type QRHandler struct {
	svc      service.QRService
	validate *validator.Validate
	log      zerolog.Logger
}

func NewQRHandler(svc service.QRService, log zerolog.Logger) *QRHandler {
	return &QRHandler{svc: svc, validate: validator.New(), log: log}
}

// Create godoc
// @Summary Create a QR payment (PromptPay / card-scheme / cross-border)
// @Tags    QR
// @Accept  json
// @Produce json
// @Param   request body domain.CreateQRRequest true "QR request"
// @Success 201 {object} domain.APIResponse{data=domain.QRPayment}
// @Router  /v1/qr-payments [post]
func (h *QRHandler) Create(c *fiber.Ctx) error {
	var req domain.CreateQRRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	// Scope to the authenticated merchant; never trust merchant_id from the body.
	if mid, ok := middleware.MerchantIDFromCtx(c); ok {
		req.MerchantID = mid
	}
	if err := h.validate.Struct(req); err != nil {
		return validationErrorResponse(c, err)
	}
	qp, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return err
	}
	return domain.Created(c, qp)
}

// Get godoc
// @Summary Poll a QR payment status
// @Router  /v1/qr-payments/{id} [get]
func (h *QRHandler) Get(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid id")
	}
	qp, err := h.svc.Get(c.Context(), mid, id)
	if err != nil {
		return err
	}
	return domain.Success(c, qp)
}

// Webhook godoc
// @Summary Bank/PSP confirmation callback for a paid QR
// @Description Signature is verified before the body is trusted. Reconciled by reference.
// @Router  /v1/webhooks/qr [post]
func (h *QRHandler) Webhook(c *fiber.Ctx) error {
	// Prefer the timestamped, replay-resistant signature (X-PayCore-Signature:
	// t=<unix>,v1=<hex>); fall back to the legacy body-only X-Signature header so
	// existing/older bank senders keep working.
	signature := c.Get("X-PayCore-Signature")
	if signature == "" {
		signature = c.Get("X-Signature")
	}
	qp, err := h.svc.ConfirmFromWebhook(c.Context(), signature, c.Body())
	if err != nil {
		return err
	}
	return domain.Success(c, qp)
}
