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

type PaymentLinkHandler struct {
	svc      service.PaymentLinkService
	validate *validator.Validate
	log      zerolog.Logger
}

func NewPaymentLinkHandler(svc service.PaymentLinkService, log zerolog.Logger) *PaymentLinkHandler {
	return &PaymentLinkHandler{svc: svc, validate: validator.New(), log: log}
}

// Create godoc
// @Router /v1/payment-links [post]
func (h *PaymentLinkHandler) Create(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	var req domain.CreatePaymentLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return validationErrorResponse(c, err)
	}
	var createdBy *uuid.UUID
	if uid, ok := middleware.UserIDFromCtx(c); ok {
		createdBy = &uid
	}
	pl, err := h.svc.Create(c.Context(), mid, createdBy, req)
	if err != nil {
		return err
	}
	return domain.Created(c, pl)
}

// List godoc
// @Router /v1/payment-links [get]
func (h *PaymentLinkHandler) List(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	limit, offset := paginate(c)
	items, err := h.svc.List(c.Context(), mid, limit, offset)
	if err != nil {
		return err
	}
	return domain.Success(c, items)
}

// Get godoc
// @Router /v1/payment-links/{id} [get]
func (h *PaymentLinkHandler) Get(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid link id")
	}
	pl, err := h.svc.Get(c.Context(), mid, id)
	if err != nil {
		return err
	}
	return domain.Success(c, pl)
}

// Disable godoc
// @Router /v1/payment-links/{id} [patch]
func (h *PaymentLinkHandler) Disable(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid link id")
	}
	pl, err := h.svc.Disable(c.Context(), mid, id)
	if err != nil {
		return err
	}
	return domain.Success(c, pl)
}
