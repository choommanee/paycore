// Package metrics exposes Prometheus instruments for the payment gateway:
// a counter of payments by terminal/interim status and a histogram of HTTP
// handler latency. Instruments are registered against a private registry that
// is served at GET /metrics.
//
// Metrics never carry cardholder data — labels are limited to low-cardinality
// dimensions (status, method, route, HTTP status class).
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Metrics bundles the gateway's Prometheus instruments and their registry.
type Metrics struct {
	reg *prometheus.Registry

	// PaymentsTotal counts payment lifecycle outcomes, labelled by the resulting
	// status (authorized, captured, refunded, voided, failed, requires_action).
	PaymentsTotal *prometheus.CounterVec

	// HTTPRequestDuration observes handler latency in seconds, labelled by
	// method, route (path template) and status class (2xx/4xx/5xx).
	HTTPRequestDuration *prometheus.HistogramVec
}

// New builds the instrument set and registers it on a fresh registry (plus the
// standard Go/process collectors). Using a private registry keeps the gateway's
// /metrics output isolated from any global default registry.
func New() *Metrics {
	reg := prometheus.NewRegistry()

	m := &Metrics{
		reg: reg,
		PaymentsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "payment_gateway",
			Subsystem: "payments",
			Name:      "total",
			Help:      "Count of payment state changes by resulting status.",
		}, []string{"status"}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "payment_gateway",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP handler latency in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "route", "status"}),
	}

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		m.PaymentsTotal,
		m.HTTPRequestDuration,
	)
	return m
}

// Registry returns the underlying registry so a /metrics handler can gather it.
func (m *Metrics) Registry() *prometheus.Registry { return m.reg }

// ObservePaymentStatus increments the payments counter for the given status.
// A nil receiver is a no-op so callers need not guard when metrics are disabled.
func (m *Metrics) ObservePaymentStatus(status string) {
	if m == nil {
		return
	}
	m.PaymentsTotal.WithLabelValues(status).Inc()
}
