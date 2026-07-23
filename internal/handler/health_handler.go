package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Pinger is the readiness dependency (a *pgxpool.Pool satisfies this).
type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	// db is pinged by the readiness probe. May be nil in scaffold/test setups,
	// in which case readiness reports the DB as unconfigured.
	db Pinger
}

// NewHealthHandler wires the readiness DB dependency.
func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Live is a liveness probe (process is up). It must not touch dependencies.
func (h *HealthHandler) Live(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// Ready is a readiness probe: verifies the DB is reachable. Returns 503 on
// failure so load balancers/orchestrators stop routing traffic here.
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	if h.db == nil {
		return c.Status(fiber.StatusServiceUnavailable).
			JSON(fiber.Map{"status": "unavailable", "reason": "database not configured"})
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).
			JSON(fiber.Map{"status": "unavailable", "reason": "database unreachable"})
	}

	return c.JSON(fiber.Map{"status": "ready"})
}
