// Command server starts the imgproxy proxy service.
package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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
// When running in Docker, it will skip loading the .env file and use environment variables directly
func loadEnvFile(logger *logging.Logger) {
	// Check if we're running in Docker environment
	_, inDocker := os.LookupEnv("DOCKER_ENV")
	if inDocker {
		logger.Info("Running in Docker environment, using environment variables directly")
		return
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		logger.Warn("Unable to identify current directory, .env loading may fail")
	}

	// Go up two directories from cmd/server to the project root
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	envPath := filepath.Join(rootDir, ".env")

	if err := godotenv.Load(envPath); err != nil {
		// Just warn rather than fail - env vars might be set some other way
		logger.Warn("Could not load .env file from %s: %v", envPath, err)
	} else {
		logger.Info("Loaded environment variables from %s", envPath)
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
