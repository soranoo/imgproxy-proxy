// Command server starts the imgproxy proxy service.
package main

import (
	"net/http"
	"path/filepath"
	"runtime"

	"imgproxy-proxy/internal/logging"
	"imgproxy-proxy/internal/proxy"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// loadEnvFile loads environment variables from .env file
func loadEnvFile(logger *logging.Logger) {
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

	// Start the server
	logger.Info(formatter.FormatServerStart(config.ServerPort, config.BaseURL))
	if err := http.ListenAndServe(config.ServerPort, nil); err != nil {
		logger.Fatal("Server error: %v", err)
	}
}
