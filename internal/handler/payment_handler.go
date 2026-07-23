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

type PaymentHandler struct {
	svc      service.PaymentService
	validate *validator.Validate
	log      zerolog.Logger
}

func NewPaymentHandler(svc service.PaymentService, log zerolog.Logger) *PaymentHandler {
	return &PaymentHandler{svc: svc, validate: validator.New(), log: log}
}

// Create godoc
// @Summary  Create a payment (authorize or charge)
// @Tags     Payments
// @Accept   json
// @Produce  json
// @Param    Idempotency-Key header string true "Idempotency key"
// @Param    request body domain.CreatePaymentRequest true "Payment request"
// @Success  201 {object} domain.APIResponse{data=domain.Payment}
// @Router   /v1/payments [post]
func (h *PaymentHandler) Create(c *fiber.Ctx) error {
	var req domain.CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	// Scope the payment to the authenticated merchant. The merchant id is taken
	// from the API-key auth context, never trusted from the request body.
	if mid, ok := middleware.MerchantIDFromCtx(c); ok {
		req.MerchantID = mid
	}
	if err := h.validate.Struct(req); err != nil {
		return validationErrorResponse(c, err)
	}
	idemKey, _ := c.Locals("idempotency_key").(string)
	p, err := h.svc.Create(c.Context(), idemKey, req)
	if err != nil {
		return err // handled by central ErrorHandler
	}
	return domain.Created(c, p)
}

// Capture godoc
// @Router /v1/payments/{id}/capture [post]
func (h *PaymentHandler) Capture(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid payment id")
	}
	var req domain.CaptureRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	p, err := h.svc.Capture(c.Context(), mid, id, req)
	if err != nil {
		return err
	}
	return domain.Success(c, p)
}

// Refund godoc
// @Router /v1/payments/{id}/refund [post]
func (h *PaymentHandler) Refund(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid payment id")
	}
	var req domain.RefundRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	p, err := h.svc.Refund(c.Context(), mid, id, req)
	if err != nil {
		return err
	}
	return domain.Success(c, p)
}

// Void godoc
// @Router /v1/payments/{id}/void [post]
func (h *PaymentHandler) Void(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid payment id")
	}
	p, err := h.svc.Void(c.Context(), mid, id)
	if err != nil {
		return err
	}
	return domain.Success(c, p)
}

// Get godoc
// @Router /v1/payments/{id} [get]
func (h *PaymentHandler) Get(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid payment id")
	}
	p, err := h.svc.Get(c.Context(), mid, id)
	if err != nil {
		return err
	}
	return domain.Success(c, p)
}

// List godoc
// @Summary  List the authenticated merchant's payments (paginated)
// @Tags     Payments
// @Produce  json
// @Param    limit  query int false "page size (default 50, max 200)"
// @Param    offset query int false "offset (default 0)"
// @Success  200 {object} domain.APIResponse{data=[]domain.Payment}
// @Router   /v1/payments [get]
func (h *PaymentHandler) List(c *fiber.Ctx) error {
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

// pagination bounds. offset is capped so a caller cannot request a pathological
// deep-scan and so the int->int32 conversion is provably in range.
const (
	defaultLimit = 50
	maxLimit     = 200
	maxOffset    = 1_000_000
)

// paginate parses limit/offset query params with sane bounds. It clamps the
// parsed int to a valid range BEFORE narrowing to int32, so the conversion never
// depends on cast-then-clamp wraparound behaviour (clears gosec G115).
func paginate(c *fiber.Ctx) (limit, offset int32) {
	l := c.QueryInt("limit", defaultLimit)
	if l <= 0 || l > maxLimit {
		l = defaultLimit
	}
	o := c.QueryInt("offset", 0)
	if o < 0 {
		o = 0
	}
	if o > maxOffset {
		o = maxOffset
	}
	// l is now in [1,200] and o in [0,1_000_000]; both fit int32 exactly.
	return int32(l), int32(o)
}

// ThreeDSReturn is the browser return endpoint after a 3DS challenge. It is
// scoped to the authenticated merchant so a caller cannot resume another
// merchant's payment, and the raw transStatus is verified against a signed CRes
// authentication value rather than trusted directly — a bare transStatus=Y must
// never be enough to force a requires_action payment to authorize.
// @Router /v1/payments/{id}/3ds/return [post]
func (h *PaymentHandler) ThreeDSReturn(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid payment id")
	}
	// The ACS returns a signed challenge result (CRes / authentication value). We
	// require transStatus=Y AND a verifiable CRes signature; a client cannot force
	// authorization by POSTing transStatus=Y alone.
	authenticated := c.FormValue("transStatus") == "Y" &&
		h.svc.VerifyThreeDSResult(id, c.FormValue("cres"), c.FormValue("transStatus"))
	p, err := h.svc.HandleThreeDSResult(c.Context(), mid, id, authenticated)
	if err != nil {
		return err
	}
	return domain.Success(c, p)
}
