package logr

import (
	"fmt"
	"slices"
)

const (
	defaultDepth = 2
	strictMode   = false
)

// Config holds global logger configuration options.
type Config struct {
	// DefaultDepth specifies the starting index in the package path
	// for layer extraction. For example, with path "myapp/internal/api/handlers"
	// and DefaultDepth=3, the layer would be "handlers".
	DefaultDepth int

	// SkipSegments lists package path segments to filter out when
	// displaying layers. Common examples: "internal", "pkg", "adapters".
	SkipSegments []string

	// StrictMode, when enabled, only allows layers specified in AllowedLayers.
	// Attempting to use an unlisted layer will cause a panic.
	StrictMode bool

	// AllowedLayers defines the valid layers when StrictMode is enabled.
	// Ignored when StrictMode is false.
	AllowedLayers []Layer
}

// packageConfig stores per-package layer configuration set via
// SetLayer() or SetDepth() calls.
type packageConfig struct {
	explicitLayer *string // Set via SetLayer()
	explicitDepth *int    // Set via SetDepth()
}

// DefaultConfig returns a Config with sensible defaults for most Go projects.
func DefaultConfig() Config {
	return Config{
		DefaultDepth: defaultDepth,
		SkipSegments: []string{
			"internal",
			"pkg",
			"cmd",
			"adapters",
			"primary",
			"secondary",
		},
		StrictMode:    strictMode,
		AllowedLayers: nil,
	}
}

// Validate checks if the configuration is valid and returns an error if not.
func (c *Config) Validate() error {
	if c.DefaultDepth < 0 {
		return fmt.Errorf("DefaultDepth must be >= 0, got %d", c.DefaultDepth)
	}

	if c.StrictMode && len(c.AllowedLayers) == 0 {
		return fmt.Errorf("StrictMode requires at least one AllowedLayers")
	}

	return nil
}

// ShouldSkipSegment checks if a package path segment should be filtered out.
func (c *Config) ShouldSkipSegment(segment string) bool {
	if c.SkipSegments == nil {
		return false
	}

	return slices.Contains(c.SkipSegments, segment)
}

// IsLayerAllowed checks if a layer is permitted by the current configuration.
// Always returns true when StrictMode is disabled.
func (c *Config) IsLayerAllowed(layer Layer) bool {
	if !c.StrictMode {
		return true
	}

	return slices.Contains(c.AllowedLayers, layer)
}
