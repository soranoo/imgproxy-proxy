package logging

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

// captureLoggerOutput provides a buffer and creates a logger with that buffer as output
func captureLoggerOutput(level int, f func(*Logger)) string {
	var buf bytes.Buffer
	logger := NewLoggerWithWriters(level, &buf, &buf, &buf, &buf, &buf, 0) // Use flag 0 to avoid timestamps in tests
	f(logger)
	return buf.String()
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name      string
		level     int
		logFunc   func(*Logger)
		contains  string
		notExists string
	}{
		{
			name:      "Debug level shows debug logs",
			level:     LevelDebug,
			logFunc:   func(l *Logger) { l.Debug("debug message") },
			contains:  "DEBUG: debug message",
			notExists: "",
		},
		{
			name:      "Info level hides debug logs",
			level:     LevelInfo,
			logFunc:   func(l *Logger) { l.Debug("hidden debug message") },
			contains:  "",
			notExists: "DEBUG: hidden debug message",
		},
		{
			name:      "Info level shows info logs",
			level:     LevelInfo,
			logFunc:   func(l *Logger) { l.Info("info message") },
			contains:  "INFO: info message",
			notExists: "",
		},
		{
			name:      "Warn level shows warn logs",
			level:     LevelWarn,
			logFunc:   func(l *Logger) { l.Warn("warning message") },
			contains:  "WARN: warning message",
			notExists: "",
		},
		{
			name:      "Error level shows error logs",
			level:     LevelError,
			logFunc:   func(l *Logger) { l.Error("error message") },
			contains:  "ERROR: error message",
			notExists: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureLoggerOutput(tt.level, tt.logFunc)

			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("expected output to contain %q, got %q", tt.contains, output)
			}

			if tt.notExists != "" && strings.Contains(output, tt.notExists) {
				t.Errorf("expected output to not contain %q, but it did", tt.notExists)
			}
		})
	}
}

func TestRequestLogger(t *testing.T) {
	output := captureLoggerOutput(LevelInfo, func(logger *Logger) {
		logger.RequestLogger("GET", "/test", "200", 150*time.Millisecond)
	})

	expectedParts := []string{"INFO:", "GET", "/test", "[200]"}
	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("expected output to contain %q, got %q", part, output)
		}
	}
}

func TestFormatter(t *testing.T) {
	formatter := NewFormatter()

	t.Run("FormatError", func(t *testing.T) {
		errMsg := formatter.FormatError(nil)
		if errMsg != "" {
			t.Errorf("expected empty string for nil error, got %q", errMsg)
		}

		errMsg = formatter.FormatError(os.ErrNotExist)
		if !strings.Contains(errMsg, "Error:") || !strings.Contains(errMsg, os.ErrNotExist.Error()) {
			t.Errorf("expected error message to contain 'Error:' and error text, got %q", errMsg)
		}
	})

	t.Run("FormatConfig", func(t *testing.T) {
		msg := formatter.FormatConfig("KEY", "value")
		if msg != "KEY: value" {
			t.Errorf("expected 'KEY: value', got %q", msg)
		}
	})

	t.Run("FormatServerStart", func(t *testing.T) {
		msg := formatter.FormatServerStart(":8080", "http://backend")
		if !strings.Contains(msg, ":8080") || !strings.Contains(msg, "http://backend") {
			t.Errorf("expected message to contain port and backend URL, got %q", msg)
		}
	})
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name  string
		level int
		msg   string
		want  string
	}{
		{
			name:  "Debug level logger",
			level: LevelDebug,
			msg:   "test debug message",
			want:  "DEBUG: ",
		},
		{
			name:  "Info level logger ignores debug",
			level: LevelInfo,
			msg:   "test debug message",
			want:  "",
		},
		{
			name:  "Error level logger",
			level: LevelError,
			msg:   "test error message",
			want:  "ERROR: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Temporarily redirect stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			logger := NewLogger(tt.level)

			// Test the appropriate log level
			switch tt.level {
			case LevelDebug:
				logger.Debug(tt.msg)
			case LevelError:
				logger.Error(tt.msg)
			default:
				logger.Debug(tt.msg) // Will be hidden for higher levels
			}

			w.Close()
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Restore original stdout and stderr
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			// Check for prefix and message content instead of exact match
			if tt.want != "" {
				if !strings.HasPrefix(output, tt.want) || !strings.Contains(output, tt.msg) {
					t.Errorf("NewLogger(%d) output = %q, should start with %q and contain %q",
						tt.level, output, tt.want, tt.msg)
				}
			} else if len(output) > 0 {
				t.Errorf("NewLogger(%d) output = %q, want empty string", tt.level, output)
			}
		})
	}
}

func TestFatalLogger(t *testing.T) {
	// Create a logger with a custom writer to capture output
	var buf bytes.Buffer
	logger := NewLoggerWithWriters(LevelFatal, &buf, &buf, &buf, &buf, &buf, 0)

	// We need to mock os.Exit to prevent the test from actually exiting
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	exitCode := 0
	osExit = func(code int) {
		exitCode = code
		panic("os.Exit called") // We'll recover from this panic
	}

	defer func() {
		if r := recover(); r != nil {
			if r != "os.Exit called" {
				t.Errorf("unexpected panic: %v", r)
			}
		} else {
			t.Error("expected panic from os.Exit, but none occurred")
		}
	}()

	logger.Fatal("fatal error message")

	output := buf.String()
	if !strings.Contains(output, "FATAL: fatal error message") {
		t.Errorf("expected output to contain %q, got %q", "FATAL: fatal error message", output)
	}

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}
