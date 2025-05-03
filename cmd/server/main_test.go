package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestHealthHandler checks if the health handler returns correct status and JSON format
func TestHealthHandler(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := healthHandler()

	// Call the handler with our test request and response recorder
	handler.ServeHTTP(rr, req)

	// Check the status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the content type is application/json
	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, expectedContentType)
	}

	// Parse the response body into a Health struct
	var health Health
	if err := json.Unmarshal(rr.Body.Bytes(), &health); err != nil {
		t.Errorf("couldn't parse response body: %v", err)
	}

	// Verify the health check fields
	if health.Status != "ok" {
		t.Errorf("expected status to be 'ok', got %v", health.Status)
	}

	// Check timestamp is reasonably recent (within last minute)
	if time.Since(health.Timestamp) > time.Minute {
		t.Errorf("timestamp is too old: %v", health.Timestamp)
	}
}
