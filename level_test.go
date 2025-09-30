package logr

import (
	"testing"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelInfo, "INFO"},
		{LevelError, "ERROR"},
		{LevelDebug, "DEBUG"},
		{LevelWarn, "WARN"},
		{LevelTest, "TEST"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if tt.level.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.level.String())
		}
	}
}
