package proxy

import (
	"net/url"
	"strings"
	"testing"
)

func TestParseQueryToOptions(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected ImageOptimizationOptions
	}{
		{
			name:     "Empty query",
			query:    "",
			expected: ImageOptimizationOptions{},
		},
		{
			name:  "Width only",
			query: "w=300",
			expected: ImageOptimizationOptions{
				Width: 300,
			},
		},
		{
			name:  "Multiple parameters",
			query: "w=300&h=200&q=90",
			expected: ImageOptimizationOptions{
				Width:   300,
				Height:  200,
				Quality: 90,
			},
		},
		{
			name:  "Invalid number",
			query: "w=invalid&h=200",
			expected: ImageOptimizationOptions{
				Height: 200,
			},
		},
		{
			name:  "Extra parameters",
			query: "w=300&h=200&extra=value",
			expected: ImageOptimizationOptions{
				Width:  300,
				Height: 200,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			got := ParseQueryToOptions(values)

			if got.Width != tt.expected.Width {
				t.Errorf("ParseQueryToOptions().Width = %v, want %v", got.Width, tt.expected.Width)
			}
			if got.Height != tt.expected.Height {
				t.Errorf("ParseQueryToOptions().Height = %v, want %v", got.Height, tt.expected.Height)
			}
			if got.Quality != tt.expected.Quality {
				t.Errorf("ParseQueryToOptions().Quality = %v, want %v", got.Quality, tt.expected.Quality)
			}
		})
	}
}

func TestParseQueryToOptionsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected ImageOptimizationOptions
	}{
		{
			name:  "Negative values",
			query: "w=-100&h=-200&q=-50",
			expected: ImageOptimizationOptions{
				Width:   -100,
				Height:  -200,
				Quality: -50,
			},
		},
		{
			name:  "Zero values",
			query: "w=0&h=0&q=0",
			expected: ImageOptimizationOptions{
				Width:   0,
				Height:  0,
				Quality: 0,
			},
		},
		{
			name:  "Very large values",
			query: "w=999999&h=999999&q=999999",
			expected: ImageOptimizationOptions{
				Width:   999999,
				Height:  999999,
				Quality: 999999,
			},
		},
		{
			name:  "Mixed valid and empty values",
			query: "w=100&h=&q=",
			expected: ImageOptimizationOptions{
				Width: 100,
			},
		},
		{
			name:     "Decimal values",
			query:    "w=100.5&h=200.7&q=90.9",
			expected: ImageOptimizationOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			got := ParseQueryToOptions(values)

			if got.Width != tt.expected.Width {
				t.Errorf("ParseQueryToOptions().Width = %v, want %v", got.Width, tt.expected.Width)
			}
			if got.Height != tt.expected.Height {
				t.Errorf("ParseQueryToOptions().Height = %v, want %v", got.Height, tt.expected.Height)
			}
			if got.Quality != tt.expected.Quality {
				t.Errorf("ParseQueryToOptions().Quality = %v, want %v", got.Quality, tt.expected.Quality)
			}
		})
	}
}

func TestParsePathOptions(t *testing.T) {
	tests := []struct {
		name         string
		pathSegments []string
		expected     string
	}{
		{
			name:         "Empty path",
			pathSegments: []string{},
			expected:     "",
		},
		{
			name:         "No options",
			pathSegments: []string{"signature", "encoded-url"},
			expected:     "",
		},
		{
			name:         "Single option",
			pathSegments: []string{"signature", "w:300", "encoded-url"},
			expected:     "w:300",
		},
		{
			name:         "Multiple options",
			pathSegments: []string{"signature", "w:300", "h:200", "q:90", "encoded-url"},
			expected:     "w:300/h:200/q:90",
		},
		{
			name:         "Mixed valid and invalid",
			pathSegments: []string{"signature", "w:300", "invalid", "q:90", "encoded-url"},
			expected:     "w:300/q:90",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePathOptions(tt.pathSegments)
			if got != tt.expected {
				t.Errorf("ParsePathOptions() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParsePathOptionsEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		pathSegments []string
		expected     string
	}{
		{
			name:         "Empty segments",
			pathSegments: []string{""},
			expected:     "",
		},
		{
			name:         "Invalid option format",
			pathSegments: []string{"w", "300", "h", "200"},
			expected:     "",
		},
		{
			name:         "Multiple colons",
			pathSegments: []string{"w:300:extra", "h:200"},
			expected:     "h:200",
		},
		{
			name:         "Mixed valid and invalid options",
			pathSegments: []string{"w:300", "invalid:option", "h:200"},
			expected:     "w:300/h:200",
		},
		{
			name:         "Option without value",
			pathSegments: []string{"w:", "h:200"},
			expected:     "h:200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePathOptions(tt.pathSegments)
			if got != tt.expected {
				t.Errorf("ParsePathOptions() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMergeOptions(t *testing.T) {
	tests := []struct {
		name      string
		pathOpts  string
		queryOpts ImageOptimizationOptions
		expected  string
	}{
		{
			name:      "Empty options",
			pathOpts:  "",
			queryOpts: ImageOptimizationOptions{},
			expected:  "",
		},
		{
			name:      "Path only",
			pathOpts:  "w:300/h:200",
			queryOpts: ImageOptimizationOptions{},
			expected:  "w:300/h:200",
		},
		{
			name:     "Query only",
			pathOpts: "",
			queryOpts: ImageOptimizationOptions{
				Width:   300,
				Height:  200,
				Quality: 90,
			},
			expected: "w:300/h:200/q:90",
		},
		{
			name:     "Override path with query",
			pathOpts: "w:300/h:200",
			queryOpts: ImageOptimizationOptions{
				Width:   400,
				Quality: 95,
			},
			expected: "w:400/h:200/q:95",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeOptions(tt.pathOpts, tt.queryOpts)

			// Since map iteration order is non-deterministic, we need to check that all options are present
			gotOptions := make(map[string]string)
			if got != "" {
				for _, opt := range strings.Split(got, "/") {
					parts := strings.Split(opt, ":")
					if len(parts) == 2 {
						gotOptions[parts[0]] = parts[1]
					}
				}
			}

			expectedOptions := make(map[string]string)
			if tt.expected != "" {
				for _, opt := range strings.Split(tt.expected, "/") {
					parts := strings.Split(opt, ":")
					if len(parts) == 2 {
						expectedOptions[parts[0]] = parts[1]
					}
				}
			}

			// Check that all expected options are present with correct values
			for k, v := range expectedOptions {
				if gotOptions[k] != v {
					t.Errorf("MergeOptions() missing or incorrect option %s:%s, got %s:%s", k, v, k, gotOptions[k])
				}
			}

			// Check that no unexpected options are present
			for k := range gotOptions {
				if _, ok := expectedOptions[k]; !ok {
					t.Errorf("MergeOptions() has unexpected option %s:%s", k, gotOptions[k])
				}
			}
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		expected string
	}{
		{
			name:     "No extension",
			urlStr:   "http://example.com/path",
			expected: "",
		},
		{
			name:     "JPG extension",
			urlStr:   "http://example.com/image.jpg",
			expected: "jpg",
		},
		{
			name:     "Uppercase extension",
			urlStr:   "http://example.com/image.JPG",
			expected: "jpg",
		},
		{
			name:     "Multiple dots",
			urlStr:   "http://example.com/path/file.name.png",
			expected: "png",
		},
		{
			name:     "Invalid URL",
			urlStr:   "://invalid",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFileExtension(tt.urlStr)
			if got != tt.expected {
				t.Errorf("GetFileExtension() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetFileExtensionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		expected string
	}{
		{
			name:     "URL with query parameters",
			urlStr:   "http://example.com/image.jpg?width=100",
			expected: "jpg",
		},
		{
			name:     "URL with fragment",
			urlStr:   "http://example.com/image.jpg#fragment",
			expected: "jpg",
		},
		{
			name:     "URL with multiple dots in path",
			urlStr:   "http://example.com/path.to.image.jpg",
			expected: "jpg",
		},
		{
			name:     "URL with dot in hostname",
			urlStr:   "http://sub.example.com/image",
			expected: "",
		},
		{
			name:     "URL with special characters",
			urlStr:   "http://example.com/image%20name.jpg",
			expected: "jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFileExtension(tt.urlStr)
			if got != tt.expected {
				t.Errorf("GetFileExtension() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestGenerateURL tests URL generation with various configurations
func TestGenerateURL(t *testing.T) {
	tests := []struct {
		name          string
		uri           string
		options       string
		config        Config
		expectErr     bool
		expectedParts []string
	}{
		{
			name:    "Basic URL with encoding",
			uri:     "http://example.com/image.jpg",
			options: "w:300",
			config: Config{
				Key:           "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				Salt:          "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				BaseURL:       "http://imgproxy:8080",
				Encode:        true,
				SignatureSize: 32,
			},
			expectErr:     false,
			expectedParts: []string{"http://imgproxy:8080", "w:300"},
		},
		{
			name:    "URL without encoding",
			uri:     "http://example.com/image.jpg",
			options: "w:300",
			config: Config{
				Key:           "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				Salt:          "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				BaseURL:       "http://imgproxy:8080",
				Encode:        false,
				SignatureSize: 32,
			},
			expectErr:     false,
			expectedParts: []string{"http://imgproxy:8080", "plain", "w:300"},
		},
		{
			name:    "Invalid key",
			uri:     "http://example.com/image.jpg",
			options: "w:300",
			config: Config{
				Key:           "invalid",
				Salt:          "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				BaseURL:       "http://imgproxy:8080",
				Encode:        true,
				SignatureSize: 32,
			},
			expectErr: true,
		},
		{
			name:    "Empty options",
			uri:     "http://example.com/image.jpg",
			options: "",
			config: Config{
				Key:           "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				Salt:          "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				BaseURL:       "http://imgproxy:8080",
				Encode:        true,
				SignatureSize: 32,
			},
			expectErr:     false,
			expectedParts: []string{"http://imgproxy:8080"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateURL(tt.uri, tt.options, tt.config)
			if (err != nil) != tt.expectErr {
				t.Errorf("GenerateURL() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr {
				for _, part := range tt.expectedParts {
					if !strings.Contains(got, part) {
						t.Errorf("GenerateURL() = %v, expected to contain %v", got, part)
					}
				}
			}
		})
	}
}
