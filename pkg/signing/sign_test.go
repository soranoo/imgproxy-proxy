package signing

import (
	"testing"
)

func TestSign(t *testing.T) {
	tests := []struct {
		name        string
		keyHex      string
		saltHex     string
		content     string
		size        int
		expected    string
		expectError bool
	}{
		{
			name:        "Valid signature",
			keyHex:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			saltHex:     "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			content:     "/w:500/aHR0cDovL2V4YW1wbGUuY29tL2ltYWdlLmpwZw==",
			size:        32,
			expected:    "w4EatShMk57MwkP0ox051lpBuMdFkeXKm1qQ1IWp91k",
			expectError: false,
		},
		{
			name:        "Invalid key hex",
			keyHex:      "ZZ",
			saltHex:     "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			content:     "/test",
			size:        32,
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid salt hex",
			keyHex:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			saltHex:     "ZZ",
			content:     "/test",
			size:        32,
			expected:    "",
			expectError: true,
		},
		{
			name:        "Short signature size",
			keyHex:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			saltHex:     "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			content:     "/test",
			size:        8,
			expected:    "Y_kxOo0wSb0",
			expectError: false,
		},
		{
			name:        "Negative signature size",
			keyHex:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			saltHex:     "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			content:     "/test",
			size:        -1,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sign(tt.keyHex, tt.saltHex, tt.content, tt.size)
			if (err != nil) != tt.expectError {
				t.Errorf("Sign() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && got != tt.expected && tt.expected != "" {
				t.Errorf("Sign() got = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestUrlSafeEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "Basic encoding",
			input:    []byte("http://example.com/image.jpg"),
			expected: "aHR0cDovL2V4YW1wbGUuY29tL2ltYWdlLmpwZw",
		},
		{
			name:     "Empty string",
			input:    []byte(""),
			expected: "",
		},
		{
			name:     "Special characters",
			input:    []byte("?&=+%#@!"),
			expected: "PyY9KyUjQCE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UrlSafeEncode(tt.input)
			if got != tt.expected {
				t.Errorf("UrlSafeEncode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestUrlSafeDecode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []byte
		expectError bool
	}{
		{
			name:        "Basic decoding",
			input:       "aHR0cDovL2V4YW1wbGUuY29tL2ltYWdlLmpwZw",
			expected:    []byte("http://example.com/image.jpg"),
			expectError: false,
		},
		{
			name:        "Empty string",
			input:       "",
			expected:    []byte(""),
			expectError: false,
		},
		{
			name:        "Invalid Base64",
			input:       "###",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UrlSafeDecode(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("UrlSafeDecode() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && string(got) != string(tt.expected) {
				t.Errorf("UrlSafeDecode() = %v, want %v", string(got), string(tt.expected))
			}
		})
	}
}
