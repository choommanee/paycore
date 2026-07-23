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

type DisputeHandler struct {
	svc      service.DisputeService
	validate *validator.Validate
	log      zerolog.Logger
}

func NewDisputeHandler(svc service.DisputeService, log zerolog.Logger) *DisputeHandler {
	return &DisputeHandler{svc: svc, validate: validator.New(), log: log}
}

// Open godoc
// @Summary  Open a dispute/chargeback against a payment
// @Tags     Disputes
// @Accept   json
// @Produce  json
// @Param    id path string true "Payment ID"
// @Param    request body domain.OpenDisputeRequest true "Dispute"
// @Success  201 {object} domain.APIResponse{data=domain.Dispute}
// @Router   /v1/payments/{id}/disputes [post]
func (h *DisputeHandler) Open(c *fiber.Ctx) error {
	paymentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid payment id")
	}
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	var req domain.OpenDisputeRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return validationErrorResponse(c, err)
	}
	d, err := h.svc.Open(c.Context(), mid, paymentID, req)
	if err != nil {
		return err
	}
	return domain.Created(c, d)
}

// ListByPayment godoc
// @Summary  List disputes for a payment
// @Tags     Disputes
// @Produce  json
// @Param    id path string true "Payment ID"
// @Success  200 {object} domain.APIResponse{data=[]domain.Dispute}
// @Router   /v1/payments/{id}/disputes [get]
func (h *DisputeHandler) ListByPayment(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	paymentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid payment id")
	}
	items, err := h.svc.ListByPayment(c.Context(), mid, paymentID)
	if err != nil {
		return err
	}
	return domain.Success(c, items)
}

// ListByMerchant godoc
// @Summary  List all disputes for the authenticated merchant
// @Tags     Disputes
// @Produce  json
// @Success  200 {object} domain.APIResponse{data=[]domain.Dispute}
// @Router   /v1/disputes [get]
func (h *DisputeHandler) ListByMerchant(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	items, err := h.svc.ListByMerchant(c.Context(), mid, c.QueryInt("limit", 100))
	if err != nil {
		return err
	}
	return domain.Success(c, items)
}
