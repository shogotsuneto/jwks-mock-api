package logger

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestLogLevel(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	tests := []struct {
		name        string
		level       string
		logFunc     func()
		shouldLog   bool
		expectedTag string
	}{
		{
			name:        "Debug level shows debug messages",
			level:       "debug",
			logFunc:     func() { Debugf("test debug") },
			shouldLog:   true,
			expectedTag: "[DEBUG]",
		},
		{
			name:        "Info level shows info messages",
			level:       "info",
			logFunc:     func() { Infof("test info") },
			shouldLog:   true,
			expectedTag: "[INFO]",
		},
		{
			name:        "Info level hides debug messages",
			level:       "info",
			logFunc:     func() { Debugf("test debug") },
			shouldLog:   false,
			expectedTag: "[DEBUG]",
		},
		{
			name:        "Error level shows error messages",
			level:       "error",
			logFunc:     func() { Errorf("test error") },
			shouldLog:   true,
			expectedTag: "[ERROR]",
		},
		{
			name:        "Error level hides info messages",
			level:       "error",
			logFunc:     func() { Infof("test info") },
			shouldLog:   false,
			expectedTag: "[INFO]",
		},
		{
			name:        "Printf works as info alias",
			level:       "info",
			logFunc:     func() { Printf("test printf") },
			shouldLog:   true,
			expectedTag: "[INFO]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset buffer
			buf.Reset()

			// Initialize logger with test level
			Init(tt.level)

			// Execute log function
			tt.logFunc()

			// Check output
			output := buf.String()
			if tt.shouldLog {
				if !strings.Contains(output, tt.expectedTag) {
					t.Errorf("Expected log output to contain %s, got: %s", tt.expectedTag, output)
				}
			} else {
				if output != "" {
					t.Errorf("Expected no log output, got: %s", output)
				}
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", DEBUG},
		{"DEBUG", DEBUG},
		{"info", INFO},
		{"INFO", INFO},
		{"warn", WARN},
		{"WARN", WARN},
		{"error", ERROR},
		{"ERROR", ERROR},
		{"invalid", INFO}, // Default to INFO for invalid levels
		{"", INFO},        // Default to INFO for empty string
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
