package logr

import (
	"strings"
	"testing"
	"time"
)

func TestPlainTextFormatterWithMetadata(t *testing.T) {
	// Create a LogEntry
	entry := NewEntry(LevelInfo, LayerHTTP, "Request started")
	entry.AddMetadata("userID", 123)
	entry.AddMetadata("session", "abc")

	// Initialize formatter
	formatter := &PlainTextFormatter{}

	// Format the entry
	output := formatter.Format(*entry)

	// Check that metadata is included
	if !strings.Contains(output, "userID=123") || !strings.Contains(output, "session=abc") {
		t.Errorf("expected metadata in output, got %q", output)
	}

	// Optional: check other parts of the log
	if !strings.Contains(output, "INFO") || !strings.Contains(output, "HTTP") || !strings.Contains(output, "Request started") {
		t.Errorf("expected log level, layer, message in output, got %q", output)
	}
}

func TestJSONFormatter(t *testing.T) {
	formatter := JSONFormatter{}

	entry := LogEntry{
		Level:     LevelInfo,
		Layer:     LayerHTTP,
		Message:   "test message",
		Timestamp: time.Date(2025, 9, 29, 12, 0, 0, 0, time.UTC),
	}

	// Metadata test
	meta := NewMetadata()
	meta.Add("requestID", "abc123")
	entry.Metadata = meta

	jsonStr := formatter.Format(entry)

	if !strings.Contains(jsonStr, `"requestID":"abc123"`) {
		t.Errorf("expected metadata in JSON output, got: %s", jsonStr)
	}

	// Also check level as string
	if !strings.Contains(jsonStr, `"level":"INFO"`) {
		t.Errorf("expected level INFO in JSON output, got: %s", jsonStr)
	}
}
