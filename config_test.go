// config_test.go
package logr

import (
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.DefaultDepth != 3 {
		t.Errorf("Expected DefaultDepth=3, got %d", config.DefaultDepth)
	}

	if config.StrictMode {
		t.Error("Expected StrictMode=false by default")
	}

	if len(config.SkipSegments) == 0 {
		t.Error("Expected default skip segments")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				DefaultDepth:  3,
				SkipSegments:  []string{"internal"},
				StrictMode:    false,
				AllowedLayers: nil,
			},
			wantError: false,
		},
		{
			name: "negative depth",
			config: Config{
				DefaultDepth: -1,
			},
			wantError: true,
			errorMsg:  "must be >= 0",
		},
		{
			name: "strict mode without allowed layers",
			config: Config{
				DefaultDepth:  3,
				StrictMode:    true,
				AllowedLayers: nil,
			},
			wantError: true,
			errorMsg:  "requires at least one AllowedLayer",
		},
		{
			name: "strict mode with allowed layers",
			config: Config{
				DefaultDepth:  3,
				StrictMode:    true,
				AllowedLayers: []Layer{LayerHTTP, LayerDB},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.wantError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestShouldSkipSegment(t *testing.T) {
	config := Config{
		SkipSegments: []string{"internal", "pkg", "adapters"},
	}

	tests := []struct {
		segment string
		want    bool
	}{
		{"internal", true},
		{"pkg", true},
		{"adapters", true},
		{"api", false},
		{"db", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.segment, func(t *testing.T) {
			got := config.ShouldSkipSegment(tt.segment)
			if got != tt.want {
				t.Errorf("ShouldSkipSegment(%q) = %v, want %v", tt.segment, got, tt.want)
			}
		})
	}
}

func TestShouldSkipSegmentNilList(t *testing.T) {
	config := Config{
		SkipSegments: nil,
	}

	if config.ShouldSkipSegment("internal") {
		t.Error("Expected false when SkipSegments is nil")
	}
}

func TestIsLayerAllowed(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		layer       Layer
		wantAllowed bool
	}{
		{
			name: "strict mode - allowed layer",
			config: Config{
				StrictMode:    true,
				AllowedLayers: []Layer{LayerHTTP, LayerDB},
			},
			layer:       LayerHTTP,
			wantAllowed: true,
		},
		{
			name: "strict mode - disallowed layer",
			config: Config{
				StrictMode:    true,
				AllowedLayers: []Layer{LayerHTTP, LayerDB},
			},
			layer:       LayerCORE,
			wantAllowed: false,
		},
		{
			name: "non-strict mode - any layer allowed",
			config: Config{
				StrictMode:    false,
				AllowedLayers: []Layer{LayerHTTP},
			},
			layer:       LayerDB,
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsLayerAllowed(tt.layer)
			if got != tt.wantAllowed {
				t.Errorf("IsLayerAllowed() = %v, want %v", got, tt.wantAllowed)
			}
		})
	}
}
