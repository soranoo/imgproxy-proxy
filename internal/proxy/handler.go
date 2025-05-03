package proxy

import (
	"io"
	"net/http"
	"strings"
	"time"

	"imgproxy-proxy/internal/logging"
	"imgproxy-proxy/internal/metrics"
	"imgproxy-proxy/pkg/signing"
)

// ProxyHandler encapsulates the dependencies needed for handling image proxy requests
type ProxyHandler struct {
	config  Config
	logger  *logging.Logger
	metrics *metrics.Metrics
}

// NewProxyHandler creates a new instance of ProxyHandler with the provided dependencies
func NewProxyHandler(config Config, logger *logging.Logger, metrics *metrics.Metrics) *ProxyHandler {
	return &ProxyHandler{
		config:  config,
		logger:  logger,
		metrics: metrics,
	}
}

// HandleImageProxy processes incoming image proxy requests by verifying signatures,
// handling image optimization options, and forwarding requests to the underlying imgproxy service.
//
// The function expects URLs in the format: /{signature}/{options}/{encoded-uri}
// where:
//   - signature: A URL-safe Base64 encoded HMAC-SHA256 signature
//   - options: Optional image processing parameters (e.g., "w:100/h:50/q:80")
//   - encoded-uri: Base64 encoded or plain source image URI
func (h *ProxyHandler) HandleImageProxy(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	path := r.URL.Path

	// Track request metrics
	h.metrics.AddRequestInProgress(path)
	defer h.metrics.RemoveRequestInProgress(path)

	// Log request start
	h.logger.Debug("Received request: %s %s", r.Method, path)

	// Parse URL and extract parts
	urlPath := r.URL.Path
	parts := strings.Split(urlPath, "/")
	if len(parts) < 3 {
		status := http.StatusBadRequest
		h.metrics.IncrementRequestsTotal(http.StatusText(status), path)
		h.metrics.ObserveRequestDuration(startTime, http.StatusText(status), path)
		h.logger.Warn("Invalid URL format: %s", path)
		http.Error(w, "Invalid URL format", status)
		return
	}

	// Extract signature and verify
	signature := parts[1]
	signablePath := strings.Join(parts[2:], "/")
	expectedSignature, err := signing.Sign(h.config.Key, h.config.Salt, "/"+signablePath, h.config.SignatureSize)
	if err != nil {
		status := http.StatusInternalServerError
		h.metrics.IncrementRequestsTotal(http.StatusText(status), path)
		h.metrics.IncrementSignatureError("invalid_key_salt")
		h.metrics.ObserveRequestDuration(startTime, http.StatusText(status), path)
		h.logger.Error("Error verifying signature: %v", err)
		http.Error(w, "Error verifying signature", status)
		return
	}

	if signature != expectedSignature {
		status := http.StatusForbidden
		h.metrics.IncrementRequestsTotal(http.StatusText(status), path)
		h.metrics.IncrementSignatureError("invalid_signature")
		h.metrics.ObserveRequestDuration(startTime, http.StatusText(status), path)
		h.logger.Warn("Invalid signature for path: %s", path)
		http.Error(w, "Invalid signature", status)
		return
	}

	// Parse existing options and query parameters
	existingOpts := ParsePathOptions(parts[2:])
	queryOpts := ParseQueryToOptions(r.URL.Query())

	// Merge options
	finalOpts := MergeOptions(existingOpts, queryOpts)

	// Determine best image format based on Accept header
	finalOpts = addFormatFromAcceptHeader(finalOpts, r.Header.Get("Accept"))

	// Generate new signed URL with updated options
	var b64TargetUri string // Base64 encoded target URI
	if len(parts) > 2 {
		b64TargetUri = parts[len(parts)-1]
	}

	// Decode the target URI if it was Base64 encoded
	decodedTargetUrl, err := signing.UrlSafeDecode(b64TargetUri)
	if err != nil {
		status := http.StatusBadRequest
		h.metrics.IncrementRequestsTotal(http.StatusText(status), path)
		h.metrics.ObserveRequestDuration(startTime, http.StatusText(status), path)
		h.logger.Error("Error decoding URL: %v", err)
		http.Error(w, "Error decoding URL", status)
		return
	}

	newUrl, err := GenerateURL(string(decodedTargetUrl), finalOpts, h.config)
	if err != nil {
		status := http.StatusInternalServerError
		h.metrics.IncrementRequestsTotal(http.StatusText(status), path)
		h.metrics.ObserveRequestDuration(startTime, http.StatusText(status), path)
		h.logger.Error("Error generating URL: %v", err)
		http.Error(w, "Error generating URL", status)
		return
	}

	// Forward the request
	h.logger.Debug("Forwarding request to backend: %s", newUrl)
	resp, err := http.Get(newUrl)
	if err != nil {
		status := http.StatusInternalServerError
		h.metrics.IncrementRequestsTotal(http.StatusText(status), path)
		h.metrics.IncrementBackendError("connection_error")
		h.metrics.ObserveRequestDuration(startTime, http.StatusText(status), path)
		h.logger.Error("Error fetching image from backend: %v", err)
		http.Error(w, "Error fetching image", status)
		return
	}
	defer resp.Body.Close()

	// Copy headers and content
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		h.logger.Error("Error copying response body: %v", err)
		h.metrics.IncrementBackendError("response_copy_error")
	}

	// Record final metrics and log
	h.metrics.IncrementRequestsTotal(http.StatusText(resp.StatusCode), path)
	h.metrics.ObserveRequestDuration(startTime, http.StatusText(resp.StatusCode), path)
	h.logger.RequestLogger(r.Method, path, http.StatusText(resp.StatusCode), time.Since(startTime))
}

// addFormatFromAcceptHeader adds format option based on Accept header.
func addFormatFromAcceptHeader(options string, acceptHeader string) string {
	var format string
	if strings.Contains(acceptHeader, "image/avif") {
		format = "f:avif"
	} else if strings.Contains(acceptHeader, "image/webp") {
		format = "f:webp"
	} else if strings.Contains(acceptHeader, "image/jpeg") {
		format = "f:jpg"
	} else if strings.Contains(acceptHeader, "image/png") {
		format = "f:png"
	}

	// Add format to options if specified
	if format != "" {
		if options != "" {
			options += "/"
		}
		options += format
	}

	return options
}

// CreateHandler returns an HTTP handler function that uses the provided configuration.
func CreateHandler(config Config) http.HandlerFunc {
	logger := logging.NewLogger(config.LogLevel)
	pMetrics := metrics.NewMetrics(config.MetricsNamespace)
	handler := NewProxyHandler(config, logger, pMetrics)

	return handler.HandleImageProxy
}
