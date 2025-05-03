package metrics

import (
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	// Create a metrics instance with a test namespace
	m := NewMetrics("test")

	// Verify that all metrics were created
	if m.RequestsTotal == nil {
		t.Error("RequestsTotal metric was not created")
	}
	if m.RequestDuration == nil {
		t.Error("RequestDuration metric was not created")
	}
	if m.RequestsInProgress == nil {
		t.Error("RequestsInProgress metric was not created")
	}
	if m.BackendErrors == nil {
		t.Error("BackendErrors metric was not created")
	}
	if m.SignatureErrors == nil {
		t.Error("SignatureErrors metric was not created")
	}
}

func TestMetricsIncrementAndObserve(t *testing.T) {
	// Create a metrics instance with a unique test namespace
	m := NewMetrics("test_increment")

	// Test incrementing request counter
	m.IncrementRequestsTotal("200", "/test")

	// Test observing request duration
	start := time.Now().Add(-100 * time.Millisecond) // 100ms ago
	m.ObserveRequestDuration(start, "200", "/test")

	// Test request in progress
	m.AddRequestInProgress("/test")
	m.RemoveRequestInProgress("/test")

	// Test error counters
	m.IncrementBackendError("test_error")
	m.IncrementSignatureError("test_error")

	// We're not testing the actual Prometheus values as that would require
	// more complex setup with registries, but we've verified the methods don't panic
}

func TestMetricsRequestFlow(t *testing.T) {
	// Create a metrics instance with a unique test namespace to avoid conflicts
	m := NewMetrics("test_flow")

	// Simulate a complete request flow
	path := "/test/image"

	// Start request
	m.AddRequestInProgress(path)

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// End request successfully
	m.RemoveRequestInProgress(path)
	status := "200"
	start := time.Now().Add(-50 * time.Millisecond)
	m.ObserveRequestDuration(start, status, path)
	m.IncrementRequestsTotal(status, path)

	// Since we can't easily assert on the prometheus metrics in a regular test
	// without more complex setup, we're just verifying that the methods don't panic
}
