package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/yourco/payment-gateway/internal/pkg/metrics"
)

// Metrics returns a Fiber middleware that records handler latency into the
// Prometheus histogram. The route label uses the matched route template (not the
// raw path) so path parameters like :id do not explode label cardinality. The
// /metrics endpoint itself is skipped to avoid self-observation noise.
func Metrics(m *metrics.Metrics) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if m == nil || c.Path() == "/metrics" {
			return c.Next()
		}
		start := time.Now()
		err := c.Next()

		route := c.Route().Path
		if route == "" {
			route = c.Path()
		}
		m.HTTPRequestDuration.
			WithLabelValues(c.Method(), route, statusClass(c.Response().StatusCode())).
			Observe(time.Since(start).Seconds())
		return err
	}
}

// statusClass buckets an HTTP status code into its class (2xx, 4xx, ...) to keep
// the metric label low-cardinality.
func statusClass(code int) string {
	if code < 100 || code >= 600 {
		return "unknown"
	}
	return strconv.Itoa(code/100) + "xx"
}

// MetricsHandler serves the Prometheus exposition format for the gateway's
// private registry at GET /metrics. It adapts the standard promhttp handler to
// fasthttp/Fiber.
func MetricsHandler(m *metrics.Metrics) fiber.Handler {
	return adaptor.HTTPHandler(promhttp.HandlerFor(m.Registry(), promhttp.HandlerOpts{}))
}
