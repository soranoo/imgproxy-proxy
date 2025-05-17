// Command server starts the imgproxy proxy service.
package main

import (
	"encoding/json"
	"net/http"
	"time"

	"imgproxy-proxy/internal/logging"
	"imgproxy-proxy/internal/proxy"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Health represents the health check response structure
type Health struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// healthHandler returns a handler function for health check requests
func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := Health{
			Status:    "ok",
			Timestamp: time.Now(),
			Version:   "1.0.0", // TODO: This could be extracted from build info in a more complex setup
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(health)
	}
}

// loadEnvFile loads environment variables from .env file
// if it exists, otherwise it uses the environment variables set in the system.
func loadEnvFile(logger *logging.Logger) {
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found, using environment variables")
	}
}

func main() {
	// Initialize logger with default log level
	logger := logging.NewLogger(logging.LevelInfo)
	formatter := logging.NewFormatter()

	// Load environment variables from .env file
	loadEnvFile(logger)

	// Load configuration from environment variables
	config := proxy.MustLoadConfig()

	// Update logger with configured log level
	logger = logging.NewLogger(config.LogLevel)

	// Create the handler with the loaded configuration
	handler := proxy.CreateHandler(config)

	// Register the handler for all paths except metrics path
	http.HandleFunc("/", handler)

	// Setup Prometheus metrics endpoint if enabled
	if config.MetricsEnabled {
		http.Handle(config.MetricsEndpoint, promhttp.Handler())
		logger.Info("Prometheus metrics enabled at %s", config.MetricsEndpoint)
	}

	// Register health check endpoint
	http.HandleFunc("/health", healthHandler())

	// Start the server
	logger.Info(formatter.FormatServerStart(config.ServerPort, config.BaseURL))
	if err := http.ListenAndServe(config.ServerPort, nil); err != nil {
		logger.Fatal("Server error: %v", err)
	}
}
