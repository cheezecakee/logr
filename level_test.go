package logr

import (
	"testing"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"}, // Changed order
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LevelTest, "TEST"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if tt.level.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.level.String())
		}
	}
}

func TestLevelOrdering(t *testing.T) {
	if LevelDebug >= LevelInfo {
		t.Error("Expected Debug < Info")
	}
	if LevelInfo >= LevelWarn {
		t.Error("Expected Info < Warn")
	}
	if LevelWarn >= LevelError {
		t.Error("Expected Warn < Error")
	}

	t.Log("Level ordering: Debug(0) < Info(1) < Warn(2) < Error(3)")
}
