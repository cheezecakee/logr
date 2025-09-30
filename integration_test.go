package logr

import (
	"strings"
	"sync"
	"testing"
)

// ============================================================================
// Integration Tests - End-to-End Scenarios
// ============================================================================

// TestEndToEndDefaultBehavior tests the complete flow with default config
func TestEndToEndDefaultBehavior(t *testing.T) {
	resetLogger()

	// Initialize with default config
	config := DefaultConfig()
	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, config)

	// Simulate logging from different "packages"
	// In real usage, getCurrentPackage would detect this automatically
	// For testing, we'll manually set up the scenario

	// Test 1: Log without any explicit configuration
	// Should use default depth extraction
	testPackagePath := "github.com/myapp/internal/api/handlers"

	layer := resolveLayer(logger, testPackagePath)

	// With DefaultDepth=3 and SkipSegments containing "internal"
	// Expected: "API/HANDLERS"
	if !strings.Contains(layer, "API") {
		t.Errorf("Expected layer to contain 'API', got: %q", layer)
	}

	t.Logf("Resolved layer for %s: %s", testPackagePath, layer)
}

// TestEndToEndWithExplicitLayer tests SetLayerForPackage flow
func TestEndToEndWithExplicitLayer(t *testing.T) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
		SkipSegments: []string{"internal"},
	})

	// Simulate a package setting its own layer
	testPkg := "github.com/myapp/database"
	customLayer := "DataLayer"

	// Manually set (in real usage, the package would call this)
	logger.registryMu.Lock()
	logger.registry[testPkg] = &packageConfig{
		explicitLayer: &customLayer,
	}
	logger.registryMu.Unlock()

	// Resolve - should return custom layer
	layer := resolveLayer(logger, testPkg)

	if layer != customLayer {
		t.Errorf("Expected custom layer %q, got %q", customLayer, layer)
	}

	t.Logf("Custom layer resolved: %s", layer)
}

// TestEndToEndInheritance tests parent-child package inheritance
func TestEndToEndInheritance(t *testing.T) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
		SkipSegments: []string{"internal"},
	})

	// Parent package sets explicit layer
	parentPkg := "github.com/myapp/api"
	parentLayer := "APILayer"

	logger.registryMu.Lock()
	logger.registry[parentPkg] = &packageConfig{
		explicitLayer: &parentLayer,
	}
	logger.registryMu.Unlock()

	// Child package should inherit
	childPkg := "github.com/myapp/api/v1/handlers"
	layer := resolveLayer(logger, childPkg)

	if layer != parentLayer {
		t.Errorf("Expected child to inherit parent layer %q, got %q", parentLayer, layer)
	}

	t.Logf("Child inherited layer: %s", layer)
}

// TestEndToEndWithDepthOverride tests SetDepth flow
func TestEndToEndWithDepthOverride(t *testing.T) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
		SkipSegments: nil,
	})

	// Package sets custom depth
	testPkg := "github.com/myapp/api/v1/handlers"
	customDepth := 1

	logger.registryMu.Lock()
	logger.registry[testPkg] = &packageConfig{
		explicitDepth: &customDepth,
	}
	logger.registryMu.Unlock()

	// Should only take last 1 segment
	layer := resolveLayer(logger, testPkg)

	if layer != "HANDLERS" {
		t.Errorf("Expected 'HANDLERS' with depth=1, got %q", layer)
	}

	t.Logf("Layer with custom depth: %s", layer)
}

// TestEndToEndCaching tests that resolution is cached
func TestEndToEndCaching(t *testing.T) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
	})

	testPkg := "github.com/myapp/api/handlers"

	// First resolution - should compute and cache
	layer1 := resolveLayer(logger, testPkg)

	// Check cache was populated
	logger.registryMu.RLock()
	cached, exists := logger.layerCache[testPkg]
	logger.registryMu.RUnlock()

	if !exists {
		t.Error("Expected layer to be cached after first resolution")
	}

	if cached != layer1 {
		t.Errorf("Cached layer %q doesn't match resolved layer %q", cached, layer1)
	}

	// Second resolution - should hit cache
	layer2 := resolveLayer(logger, testPkg)

	if layer1 != layer2 {
		t.Errorf("Expected same layer from cache, got %q and %q", layer1, layer2)
	}

	t.Logf("Layer cached successfully: %s", layer1)
}

// TestEndToEndMultiplePackages simulates real-world scenario
func TestEndToEndMultiplePackages(t *testing.T) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
		SkipSegments: []string{"internal", "pkg"},
	})

	packages := map[string]string{
		"github.com/myapp/internal/api/handlers": "API/HANDLERS",
		"github.com/myapp/internal/db/postgres":  "DB/POSTGRES",
		"github.com/myapp/pkg/cache/redis":       "CACHE/REDIS",
		"github.com/myapp/services/auth":         "SERVICES/AUTH",
	}

	for pkgPath, expectedLayer := range packages {
		layer := resolveLayer(logger, pkgPath)

		if layer != expectedLayer {
			t.Errorf("Package %s: expected %q, got %q", pkgPath, expectedLayer, layer)
		} else {
			t.Logf("✓ Package %s resolved to: %s", pkgPath, layer)
		}
	}
}

// TestEndToEndStrictMode tests layer validation in strict mode
func TestEndToEndStrictMode(t *testing.T) {
	resetLogger()

	config := Config{
		DefaultDepth:  2,
		SkipSegments:  []string{"internal"},
		StrictMode:    true,
		AllowedLayers: []Layer{LayerHTTP, LayerDB, LayerCORE},
	}

	_ = InitWithConfig(&PlainTextFormatter{}, LevelInfo, config)

	// Allowed layer - should work
	if !config.IsLayerAllowed(LayerHTTP) {
		t.Error("Expected LayerHTTP to be allowed in strict mode")
	}

	// Disallowed layer - should fail
	customLayer := Layer("CUSTOM")
	if config.IsLayerAllowed(customLayer) {
		t.Error("Expected custom layer to be disallowed in strict mode")
	}

	t.Logf("Strict mode validation working correctly")
}

// TestEndToEndConcurrentUsage tests thread safety under load
func TestEndToEndConcurrentUsage(t *testing.T) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
		SkipSegments: []string{"internal"},
	})

	var wg sync.WaitGroup
	numGoroutines := 100

	packages := []string{
		"github.com/myapp/api/handlers",
		"github.com/myapp/db/postgres",
		"github.com/myapp/cache/redis",
		"github.com/myapp/services/auth",
	}

	// Simulate concurrent logging from multiple packages
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			pkg := packages[idx%len(packages)]

			// Resolve layer (might hit cache or compute)
			layer := resolveLayer(logger, pkg)

			if layer == "" || layer == "UNKNOWN" {
				t.Errorf("Goroutine %d: got invalid layer for %s", idx, pkg)
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Successfully handled %d concurrent resolutions", numGoroutines)
}

// TestEndToEndRealWorldScenario simulates a complete application setup
func TestEndToEndRealWorldScenario(t *testing.T) {
	resetLogger()

	// App initializes logger once at startup
	config := Config{
		DefaultDepth: 2,
		SkipSegments: []string{"internal", "pkg", "cmd"},
		StrictMode:   false,
	}

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, config)

	// Database package sets custom layer
	dbLayer := "Database"
	logger.registryMu.Lock()
	logger.registry["myapp/internal/db"] = &packageConfig{
		explicitLayer: &dbLayer,
	}
	logger.registryMu.Unlock()

	// API package uses custom depth
	apiDepth := 1
	logger.registryMu.Lock()
	logger.registry["myapp/api/v1/users"] = &packageConfig{
		explicitDepth: &apiDepth,
	}
	logger.registryMu.Unlock()

	// Test scenarios
	scenarios := []struct {
		name     string
		pkg      string
		expected string
	}{
		{
			name:     "DB package with explicit layer",
			pkg:      "myapp/internal/db",
			expected: "Database",
		},
		{
			name:     "DB child inherits parent layer",
			pkg:      "myapp/internal/db/postgres",
			expected: "Database",
		},
		{
			name:     "API package with custom depth",
			pkg:      "myapp/api/v1/users",
			expected: "USERS",
		},
		{
			name:     "Default behavior for cache",
			pkg:      "myapp/internal/cache/redis",
			expected: "CACHE/REDIS",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			layer := resolveLayer(logger, sc.pkg)

			if layer != sc.expected {
				t.Errorf("Expected %q, got %q", sc.expected, layer)
			} else {
				t.Logf("✓ %s: %s", sc.name, layer)
			}
		})
	}
}

// TestEndToEndSkipSegmentsFiltering tests various skip scenarios
func TestEndToEndSkipSegmentsFiltering(t *testing.T) {
	resetLogger()

	config := Config{
		DefaultDepth: 3,
		SkipSegments: []string{"internal", "pkg", "adapters"},
	}

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, config)

	tests := []struct {
		pkg      string
		expected string
	}{
		{
			pkg:      "myapp/internal/api/handlers",
			expected: "API/HANDLERS", // "internal" skipped
		},
		{
			pkg:      "myapp/pkg/adapters/http",
			expected: "HTTP", // "pkg" and "adapters" skipped
		},
		{
			pkg:      "myapp/services/auth/jwt",
			expected: "SERVICES/AUTH/JWT", // nothing skipped
		},
	}

	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			layer := resolveLayer(logger, tt.pkg)

			if layer != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, layer)
			}
		})
	}
}
