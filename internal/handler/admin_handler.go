package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/service"
)

// AdminHandler serves the platform operator console (behind X-Admin-Key). The
// AdminAuth middleware authorizes the request before any of these run; the
// handlers themselves are unscoped platform reads.
type AdminHandler struct {
	svc service.AdminService
	log zerolog.Logger
}

func NewAdminHandler(svc service.AdminService, log zerolog.Logger) *AdminHandler {
	return &AdminHandler{svc: svc, log: log}
}

// Merchants godoc
// @Summary  List all merchants (admin)
// @Router   /v1/admin/merchants [get]
func (h *AdminHandler) Merchants(c *fiber.Ctx) error {
	rows, err := h.svc.ListMerchants(c.Context(), c.QueryInt("limit", 100))
	if err != nil {
		return err
	}
	return domain.Success(c, rows)
}

// AuditLog godoc
// @Summary  Recent audit-log rows (admin)
// @Router   /v1/admin/audit-log [get]
func (h *AdminHandler) AuditLog(c *fiber.Ctx) error {
	rows, err := h.svc.ListAuditLog(c.Context(), c.QueryInt("limit", 100))
	if err != nil {
		return err
	}
	return domain.Success(c, rows)
}

// Settlements godoc
// @Summary  All settlement/payout rows across the platform (admin)
// @Router   /v1/admin/settlements [get]
func (h *AdminHandler) Settlements(c *fiber.Ctx) error {
	rows, err := h.svc.ListSettlements(c.Context(), c.QueryInt("limit", 100))
	if err != nil {
		return err
	}
	return domain.Success(c, rows)
}

// Stats godoc
// @Summary  Platform-wide key risk indicators (admin)
// @Router   /v1/admin/stats [get]
func (h *AdminHandler) Stats(c *fiber.Ctx) error {
	stats, err := h.svc.PlatformStats(c.Context())
	if err != nil {
		return err
	}
	return domain.Success(c, stats)
}

// Disputes godoc
// @Summary  All disputes across the platform (admin)
// @Router   /v1/admin/disputes [get]
func (h *AdminHandler) Disputes(c *fiber.Ctx) error {
	rows, err := h.svc.ListDisputes(c.Context(), c.QueryInt("limit", 100))
	if err != nil {
		return err
	}
	return domain.Success(c, rows)
}
