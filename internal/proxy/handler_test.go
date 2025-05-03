package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"imgproxy-proxy/internal/logging"
	"imgproxy-proxy/internal/metrics"
)

func TestAddFormatFromAcceptHeader(t *testing.T) {
	tests := []struct {
		name         string
		options      string
		acceptHeader string
		expected     string
	}{
		{
			name:         "Empty options, no Accept header",
			options:      "",
			acceptHeader: "",
			expected:     "",
		},
		{
			name:         "Empty options, AVIF Accept header",
			options:      "",
			acceptHeader: "image/avif,image/webp,image/png,image/jpeg",
			expected:     "f:avif",
		},
		{
			name:         "Empty options, WebP Accept header",
			options:      "",
			acceptHeader: "image/webp,image/png,image/jpeg",
			expected:     "f:webp",
		},
		{
			name:         "Empty options, JPEG Accept header",
			options:      "",
			acceptHeader: "image/jpeg",
			expected:     "f:jpg",
		},
		{
			name:         "Empty options, PNG Accept header",
			options:      "",
			acceptHeader: "image/png",
			expected:     "f:png",
		},
		{
			name:         "Existing options, WebP Accept header",
			options:      "w:100/h:200",
			acceptHeader: "image/webp,image/jpeg",
			expected:     "w:100/h:200/f:webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addFormatFromAcceptHeader(tt.options, tt.acceptHeader)
			if got != tt.expected {
				t.Errorf("addFormatFromAcceptHeader() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCreateHandler(t *testing.T) {
	// This is a simple test to verify that CreateHandler returns a http.HandlerFunc
	// More comprehensive tests would mock the HTTP client
	config := Config{
		Key:              "0123456789abcdef0123456789abcdef",
		Salt:             "0123456789abcdef0123456789abcdef",
		BaseURL:          "http://localhost:8081",
		Encode:           true,
		SignatureSize:    32,
		MetricsEnabled:   true,
		MetricsEndpoint:  "/metrics",
		MetricsNamespace: "test",
		LogLevel:         logging.LevelInfo,
	}

	handler := CreateHandler(config)

	// Verify that the returned value is of type http.HandlerFunc
	if handler == nil {
		t.Error("CreateHandler() returned nil")
	}
}

func TestProxyHandler(t *testing.T) {
	// Create dependencies
	config := Config{
		Key:              "0123456789abcdef0123456789abcdef",
		Salt:             "0123456789abcdef0123456789abcdef",
		BaseURL:          "http://localhost:8081",
		Encode:           true,
		SignatureSize:    32,
		MetricsEnabled:   true,
		MetricsEndpoint:  "/metrics",
		MetricsNamespace: "test",
		LogLevel:         logging.LevelInfo,
	}
	logger := logging.NewLogger(logging.LevelDebug)
	metrics := metrics.NewMetrics("test") // Use NewTestMetrics instead of NewMetrics

	// Create handler
	handler := NewProxyHandler(config, logger, metrics)

	if handler == nil {
		t.Fatal("NewProxyHandler() returned nil")
	}

	// Simple validation of handler properties
	if handler.config.Key != config.Key {
		t.Errorf("handler has incorrect Key, got %s, want %s", handler.config.Key, config.Key)
	}

	if handler.logger == nil {
		t.Error("handler.logger is nil")
	}

	if handler.metrics == nil {
		t.Error("handler.metrics is nil")
	}
}

// MockResponseWriter is a mock implementation of http.ResponseWriter for testing
type MockResponseWriter struct {
	Headers    http.Header
	StatusCode int
	Body       []byte
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		Headers: make(http.Header),
	}
}

func (m *MockResponseWriter) Header() http.Header {
	return m.Headers
}

func (m *MockResponseWriter) Write(body []byte) (int, error) {
	m.Body = append(m.Body, body...)
	return len(body), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.StatusCode = statusCode
}

// TestHandleImageProxyBadRequest tests the handler's response to invalid URLs
func TestHandleImageProxyBadRequest(t *testing.T) {
	// Create dependencies
	config := Config{
		Key:              "0123456789abcdef0123456789abcdef",
		Salt:             "0123456789abcdef0123456789abcdef",
		BaseURL:          "http://localhost:8081",
		Encode:           true,
		SignatureSize:    32,
		LogLevel:         logging.LevelDebug,
		MetricsNamespace: "test",
	}
	logger := logging.NewLogger(logging.LevelDebug)
	m := metrics.NewMetrics("test") // Use NewTestMetrics instead of NewMetrics

	// Create handler
	handler := NewProxyHandler(config, logger, m)

	// Create a request with an invalid path (too short)
	req := httptest.NewRequest("GET", "/invalidpath", nil)
	w := NewMockResponseWriter()

	// Call the handler
	handler.HandleImageProxy(w, req)

	// Verify response
	if w.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.StatusCode)
	}
}
