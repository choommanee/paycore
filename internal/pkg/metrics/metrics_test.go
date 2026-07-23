package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewRegistersInstruments(t *testing.T) {
	m := New()
	if m.Registry() == nil {
		t.Fatal("registry is nil")
	}
	if m.PaymentsTotal == nil || m.HTTPRequestDuration == nil {
		t.Fatal("instruments not initialized")
	}
}

func TestObservePaymentStatus(t *testing.T) {
	m := New()
	m.ObservePaymentStatus("authorized")
	m.ObservePaymentStatus("authorized")
	m.ObservePaymentStatus("captured")

	if got := testutil.ToFloat64(m.PaymentsTotal.WithLabelValues("authorized")); got != 2 {
		t.Fatalf("authorized counter=%v want 2", got)
	}
	if got := testutil.ToFloat64(m.PaymentsTotal.WithLabelValues("captured")); got != 1 {
		t.Fatalf("captured counter=%v want 1", got)
	}
}

// TestObservePaymentStatusNilSafe verifies the nil-receiver no-op contract.
func TestObservePaymentStatusNilSafe(t *testing.T) {
	var m *Metrics
	m.ObservePaymentStatus("failed") // must not panic
}

func TestHistogramObserves(t *testing.T) {
	m := New()
	m.HTTPRequestDuration.WithLabelValues("POST", "/v1/payments", "2xx").Observe(0.05)
	count := testutil.CollectAndCount(m.HTTPRequestDuration)
	if count == 0 {
		t.Fatal("expected histogram to expose at least one series")
	}
}

// TestGatherExposesMetrics confirms the registry gathers the gateway metrics in
// the Prometheus exposition format.
func TestGatherExposesMetrics(t *testing.T) {
	m := New()
	m.ObservePaymentStatus("authorized")
	m.HTTPRequestDuration.WithLabelValues("POST", "/v1/payments", "2xx").Observe(0.01)
	mfs, err := m.Registry().Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	var names []string
	for _, mf := range mfs {
		names = append(names, mf.GetName())
	}
	joined := strings.Join(names, ",")
	if !strings.Contains(joined, "payment_gateway_payments_total") {
		t.Fatalf("payments counter missing from gathered metrics: %s", joined)
	}
	if !strings.Contains(joined, "payment_gateway_http_request_duration_seconds") {
		t.Fatalf("http histogram missing from gathered metrics: %s", joined)
	}
}
