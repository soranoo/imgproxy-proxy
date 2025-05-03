// Package metrics provides Prometheus metrics collection functionality for the imgproxy proxy service.
package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all the prometheus metrics used in the application
type Metrics struct {
	RequestsTotal      *prometheus.CounterVec
	RequestDuration    *prometheus.HistogramVec
	RequestsInProgress *prometheus.GaugeVec
	BackendErrors      *prometheus.CounterVec
	SignatureErrors    *prometheus.CounterVec
}

// Add a package-level variable to hold the singleton instance
var metricsInstance *Metrics

// Add a mutex to ensure thread-safe initialization
var metricsOnce sync.Once

// NewMetrics creates and registers all prometheus metrics (singleton pattern)
func NewMetrics(namespace string) *Metrics {
	metricsOnce.Do(func() {
		metricsInstance = &Metrics{
			RequestsTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: namespace,
					Name:      "requests_total",
					Help:      "Total number of image proxy requests processed",
				},
				[]string{"status", "path"},
			),
			RequestDuration: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: namespace,
					Name:      "request_duration_seconds",
					Help:      "Duration of image proxy requests in seconds",
					Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
				},
				[]string{"status", "path"},
			),
			RequestsInProgress: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "requests_in_progress",
					Help:      "Current number of image proxy requests being processed",
				},
				[]string{"path"},
			),
			BackendErrors: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: namespace,
					Name:      "backend_errors_total",
					Help:      "Total number of backend errors during image proxying",
				},
				[]string{"type"},
			),
			SignatureErrors: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: namespace,
					Name:      "signature_errors_total",
					Help:      "Total number of signature validation errors",
				},
				[]string{"type"},
			),
		}
	})
	return metricsInstance
}

// ObserveRequestDuration records the duration of a request
func (m *Metrics) ObserveRequestDuration(start time.Time, status string, path string) {
	duration := time.Since(start).Seconds()
	m.RequestDuration.WithLabelValues(status, path).Observe(duration)
}

// IncrementRequestsTotal increments the total requests counter
func (m *Metrics) IncrementRequestsTotal(status string, path string) {
	m.RequestsTotal.WithLabelValues(status, path).Inc()
}

// AddRequestInProgress increments the in-progress requests gauge
func (m *Metrics) AddRequestInProgress(path string) {
	m.RequestsInProgress.WithLabelValues(path).Inc()
}

// RemoveRequestInProgress decrements the in-progress requests gauge
func (m *Metrics) RemoveRequestInProgress(path string) {
	m.RequestsInProgress.WithLabelValues(path).Dec()
}

// IncrementBackendError increments the backend error counter
func (m *Metrics) IncrementBackendError(errorType string) {
	m.BackendErrors.WithLabelValues(errorType).Inc()
}

// IncrementSignatureError increments the signature error counter
func (m *Metrics) IncrementSignatureError(errorType string) {
	m.SignatureErrors.WithLabelValues(errorType).Inc()
}
