package logr

import (
	"strings"
	"testing"
)

func TestPlainTextFormatterWithMetadata(t *testing.T) {
	// Create a LogEntry
	entry := NewEntry(LevelInfo, "HTTP", "Request started")
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
