// Package proxy implements the core functionality of the imgproxy proxy service.
// It provides HTTP handlers, configuration, and utilities for interacting with
// backend imgproxy services.
package proxy

import (
	"fmt"

	"imgproxy-proxy/internal/logging"

	"github.com/kelseyhightower/envconfig"
)

// Config holds configuration options for generating imgproxy URLs.
type Config struct {
	Encode        bool   `envconfig:"IMGPROXY_ENCODE" default:"true"`       // Encode indicates whether the source URI should be Base64 encoded.
	Salt          string `envconfig:"IMGPROXY_SALT"`                        // Salt is the hex-encoded salt used for signing secure URLs.
	Key           string `envconfig:"IMGPROXY_KEY"`                         // Key is the hex-encoded key used for signing secure URLs.
	SignatureSize int    `envconfig:"IMGPROXY_SIGNATURE_SIZE" default:"32"` // SignatureSize specifies the desired length of the generated signature in bytes (max 32).
	BaseURL       string `envconfig:"IMGPROXY_BASE_URL"`                    // BaseURL is the base URL of the imgproxy service.

	// Metrics and logging configuration
	MetricsEnabled   bool   `envconfig:"METRICS_ENABLED" default:"true"`             // Whether to enable Prometheus metrics
	MetricsEndpoint  string `envconfig:"METRICS_ENDPOINT" default:"/metrics"`        // Endpoint for Prometheus metrics
	MetricsNamespace string `envconfig:"METRICS_NAMESPACE" default:"imgproxy_proxy"` // Namespace for Prometheus metrics
	LogLevel         int    `envconfig:"LOG_LEVEL" default:"1"`                      // Log level (0=DEBUG, 1=INFO, 2=WARN, 3=ERROR, 4=FATAL)
	ServerPort       string `envconfig:"SERVER_PORT" default:":8080"`                // Port on which the server listens
}

// LoadConfig loads configuration from environment variables.
// It returns a Config struct and an error if the configuration is invalid.
func LoadConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("", &config); err != nil {
		return config, fmt.Errorf("error loading configuration: %w", err)
	}

	// Validate required fields
	if config.Key == "" {
		return config, fmt.Errorf("IMGPROXY_KEY environment variable is required")
	}
	if config.Salt == "" {
		return config, fmt.Errorf("IMGPROXY_SALT environment variable is required")
	}
	if config.BaseURL == "" {
		return config, fmt.Errorf("IMGPROXY_BASE_URL environment variable is required")
	}

	return config, nil
}

// MustLoadConfig loads configuration from environment variables and
// exits the program if the configuration is invalid.
func MustLoadConfig() Config {
	logger := logging.NewLogger(logging.LevelInfo)
	config, err := LoadConfig()
	if err != nil {
		logger.Fatal("Configuration error: %v", err)
	}

	// Update logger with configured log level
	logger = logging.NewLogger(config.LogLevel)
	return config
}
