// Package logging provides standardized logging capabilities for the imgproxy proxy service.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// For testing purposes, we can swap this out
var osExit = os.Exit

// Log levels
const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// Logger is a standardized logger for the application
type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	level       int
}

// NewLogger creates a new logger with the specified minimum log level
func NewLogger(level int) *Logger {
	flags := log.Ldate | log.Ltime
	return NewLoggerWithWriters(level, os.Stdout, os.Stdout, os.Stdout, os.Stderr, os.Stderr, flags)
}

// NewLoggerWithWriters creates a new logger with custom writers and flags
// This function is particularly useful for testing to capture log output
func NewLoggerWithWriters(level int, debugWriter, infoWriter, warnWriter, errorWriter, fatalWriter io.Writer, flags int) *Logger {
	return &Logger{
		debugLogger: log.New(debugWriter, "DEBUG: ", flags),
		infoLogger:  log.New(infoWriter, "INFO: ", flags),
		warnLogger:  log.New(warnWriter, "WARN: ", flags),
		errorLogger: log.New(errorWriter, "ERROR: ", flags),
		fatalLogger: log.New(fatalWriter, "FATAL: ", flags),
		level:       level,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= LevelDebug {
		l.debugLogger.Printf(format, v...)
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= LevelInfo {
		l.infoLogger.Printf(format, v...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= LevelWarn {
		l.warnLogger.Printf(format, v...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= LevelError {
		l.errorLogger.Printf(format, v...)
	}
}

// Fatal logs a fatal error message and exits the application
func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.level <= LevelFatal {
		l.fatalLogger.Printf(format, v...)
		osExit(1)
	}
}

// RequestLogger logs HTTP request information with timing
func (l *Logger) RequestLogger(method, path, status string, duration time.Duration) {
	l.Info("%s %s [%s] %s", method, path, status, duration)
}

// Formatter provides consistent message formatting across the application
type Formatter struct{}

// NewFormatter creates a new formatter
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FormatError returns a formatted error message
func (f *Formatter) FormatError(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("Error: %v", err)
}

// FormatConfig formats configuration information for display
func (f *Formatter) FormatConfig(name, value string) string {
	return fmt.Sprintf("%s: %s", name, value)
}

// FormatServerStart formats server startup message
func (f *Formatter) FormatServerStart(addr, target string) string {
	return fmt.Sprintf("Server started on %s, proxying to %s", addr, target)
}
