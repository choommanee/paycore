package handler

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/service"
)

type MerchantHandler struct {
	svc      service.MerchantService
	validate *validator.Validate
	log      zerolog.Logger
}

func NewMerchantHandler(svc service.MerchantService, log zerolog.Logger) *MerchantHandler {
	return &MerchantHandler{svc: svc, validate: validator.New(), log: log}
}

// Onboard godoc
// @Summary  Onboard a merchant (returns the API key once)
// @Tags     Merchants
// @Accept   json
// @Produce  json
// @Param    request body domain.CreateMerchantRequest true "Merchant"
// @Success  201 {object} domain.APIResponse{data=domain.MerchantCredential}
// @Router   /v1/merchants [post]
func (h *MerchantHandler) Onboard(c *fiber.Ctx) error {
	var req domain.CreateMerchantRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return validationErrorResponse(c, err)
	}
	cred, err := h.svc.Onboard(c.Context(), req)
	if err != nil {
		return err
	}
	return domain.Created(c, cred)
}

// Get godoc
// @Router /v1/merchants/{id} [get]
func (h *MerchantHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid merchant id")
	}
	// Ownership check (OWASP API1:2023 BOLA). GET /v1/merchants/:id is behind
	// auth but the :id path is otherwise unscoped, so an authenticated merchant
	// could read any other merchant's profile by guessing UUIDs. Only allow a
	// merchant to read its own profile, and return 404 (not 403) so the endpoint
	// does not become an existence oracle for merchant ids.
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok || mid != id {
		return domain.Error(c, fiber.StatusNotFound, "MERCHANT_NOT_FOUND", "not found")
	}
	m, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return err
	}
	return domain.Success(c, m)
}

// Me godoc
// @Summary  The authenticated merchant (self-view)
// @Router   /v1/me [get]
func (h *MerchantHandler) Me(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	p, err := h.svc.Profile(c.Context(), mid)
	if err != nil {
		return err
	}
	return domain.Success(c, p)
}

// Stats godoc
// @Summary  Merchant KPIs over an optional [from, to] window
// @Router   /v1/stats [get]
func (h *MerchantHandler) Stats(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	// Default window: the last 30 days through now. from/to are RFC3339 (or a
	// date). A malformed value is a client error rather than a silent default so
	// the dashboard surfaces a bad range instead of showing wrong numbers.
	now := time.Now().UTC()
	from := now.AddDate(0, 0, -30)
	to := now
	if v := c.Query("from"); v != "" {
		t, err := parseTimeParam(v)
		if err != nil {
			return domain.Error(c, fiber.StatusBadRequest, "INVALID_RANGE", "invalid from")
		}
		from = t
	}
	if v := c.Query("to"); v != "" {
		t, err := parseTimeParam(v)
		if err != nil {
			return domain.Error(c, fiber.StatusBadRequest, "INVALID_RANGE", "invalid to")
		}
		to = t
	}
	if to.Before(from) {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_RANGE", "to is before from")
	}
	stats, err := h.svc.Stats(c.Context(), mid, from, to)
	if err != nil {
		return err
	}
	return domain.Success(c, stats)
}

// StatsSeries godoc
// @Summary  Daily payment series (volume + count) for the dashboard sparkline
// @Router   /v1/stats/series [get]
func (h *MerchantHandler) StatsSeries(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	// days defaults to 30 and is clamped to [1, 90]; unlike from/to on /v1/stats,
	// an out-of-range value is clamped rather than rejected since it only bounds
	// how much history is drawn, not which payments are counted.
	days := c.QueryInt("days", 30)
	if days < 1 {
		days = 1
	}
	if days > 90 {
		days = 90
	}
	series, err := h.svc.StatsSeries(c.Context(), mid, days)
	if err != nil {
		return err
	}
	return domain.Success(c, series)
}

// Settlements godoc
// @Summary  The merchant's settlement / payout rows
// @Router   /v1/settlements [get]
func (h *MerchantHandler) Settlements(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	rows, err := h.svc.ListSettlements(c.Context(), mid, c.QueryInt("limit", 50))
	if err != nil {
		return err
	}
	return domain.Success(c, rows)
}

// ListTransactions godoc
// @Summary  Unified transaction feed (card + PromptPay + wallet), newest first
// @Router   /v1/transactions [get]
func (h *MerchantHandler) ListTransactions(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	limit, offset := paginate(c)
	items, err := h.svc.ListTransactions(c.Context(), mid, limit, offset)
	if err != nil {
		return err
	}
	return domain.Success(c, items)
}

// RotateKey godoc
// @Summary  Issue a new API key (returned once), invalidating the old key
// @Router   /v1/me/rotate-key [post]
func (h *MerchantHandler) RotateKey(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	cred, err := h.svc.RotateAPIKey(c.Context(), mid)
	if err != nil {
		return err
	}
	return domain.Success(c, cred)
}

// SetWebhook godoc
// @Summary  Set the merchant webhook URL and rotate its signing secret (once)
// @Router   /v1/me/webhook [put]
func (h *MerchantHandler) SetWebhook(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	var req domain.SetWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return validationErrorResponse(c, err)
	}
	cfg, err := h.svc.SetWebhook(c.Context(), mid, req.URL)
	if err != nil {
		return err
	}
	return domain.Success(c, cfg)
}

// parseTimeParam accepts an RFC3339 timestamp or a bare YYYY-MM-DD date.
func parseTimeParam(v string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t.UTC(), nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}
