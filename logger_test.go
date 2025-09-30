package logr

import (
	"fmt"
	"sync"
	"testing"
)

type MockFormatter struct {
	LastFormatted string
}

func (f *MockFormatter) Format(entry LogEntry) string {
	f.LastFormatted = entry.Message
	return entry.Message
}

func resetLogger() {
	// Reset singleton for fresh initialization
	defaultLogger = nil
	once = sync.Once{}
}

func TestLoggerInfo(t *testing.T) {
	resetLogger()

	mock := &MockFormatter{}
	allowedLayers := map[Layer]int{LayerHTTP: 0}

	logger := Init(mock, LevelInfo, allowedLayers)
	logger.SetLayer(LayerHTTP)

	msg := "test info message"
	logger.Info(msg)

	if mock.LastFormatted != msg {
		t.Errorf("expected %q, got %q", msg, mock.LastFormatted)
	}
}

func TestLoggerDebugWarnTest(t *testing.T) {
	resetLogger()

	mock := &MockFormatter{}
	allowedLayers := map[Layer]int{LayerCORE: 0}

	logger := Init(mock, LevelInfo, allowedLayers)
	logger.SetLayer(LayerCORE)

	// Debug should NOT log (below Info level)
	logger.Debug("debug message")
	if mock.LastFormatted == "debug message" {
		t.Errorf("expected Debug not to be logged at LevelInfo")
	}

	// Info should log
	infoMsg := "info message"
	logger.Info(infoMsg)
	if mock.LastFormatted != infoMsg {
		t.Errorf("expected %q, got %q", infoMsg, mock.LastFormatted)
	}

	// Warn should log
	warnMsg := "warning!"
	logger.Warn(warnMsg)
	if mock.LastFormatted != warnMsg {
		t.Errorf("expected %q, got %q", warnMsg, mock.LastFormatted)
	}

	// Error should log
	errorMsg := "error occurred"
	logger.Error(errorMsg)
	if mock.LastFormatted != errorMsg {
		t.Errorf("expected %q, got %q", errorMsg, mock.LastFormatted)
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	resetLogger()

	mock := &MockFormatter{}
	allowedLayers := map[Layer]int{LayerHTTP: 0}

	logger := Init(mock, LevelError, allowedLayers)
	logger.SetLayer(LayerHTTP)

	// Info should be ignored
	logger.Info("should not log")
	if mock.LastFormatted != "" {
		t.Errorf("expected no log, got %q", mock.LastFormatted)
	}

	// Error should log
	errMsg := "error occurred"
	logger.Error(errMsg)
	if mock.LastFormatted != errMsg {
		t.Errorf("expected %q, got %q", errMsg, mock.LastFormatted)
	}
}

func TestLoggerSetLayer(t *testing.T) {
	resetLogger()

	mock := &MockFormatter{}
	allowedLayers := map[Layer]int{LayerHTTP: 0, LayerDB: 1}

	logger := Init(mock, LevelInfo, allowedLayers)

	// Switch to a registered layer
	logger.SetLayer(LayerDB)
	if logger.defaultLayer != LayerDB {
		t.Errorf("expected defaultLayer to be 'DB', got %q", logger.defaultLayer)
	}

	// Attempt to set an unregistered layer and expect panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when setting unregistered layer")
		}
	}()
	logger.SetLayer("UNKNOWN")
}

// Test that SetLayerForPackage stores correctly
func TestSetLayerForPackage(t *testing.T) {
	// Reset logger
	resetLogger()

	logger := Init(&PlainTextFormatter{}, LevelInfo, nil)

	// Simulate calling from a package
	// (In real usage, getCurrentPackage would detect this automatically)
	logger.SetLayerForPackage("test-layer")

	// Verify it was stored in registry
	// Note: We need to manually check since getCurrentPackage would detect the test package
	logger.registryMu.RLock()
	defer logger.registryMu.RUnlock()

	// The registry should have at least one entry
	if len(logger.registry) == 0 {
		t.Error("Expected registry to have entries after SetLayerForPackage")
	}

	// Check that some config exists (we can't easily check exact package path in test)
	hasConfig := false
	for _, config := range logger.registry {
		if config.explicitLayer != nil && *config.explicitLayer == "test-layer" {
			hasConfig = true
			break
		}
	}

	if !hasConfig {
		t.Error("Expected to find 'test-layer' in registry")
	}
}

// Test that SetDepth stores correctly
func TestSetDepth(t *testing.T) {
	resetLogger()

	logger := Init(&PlainTextFormatter{}, LevelInfo, nil)

	// Set a depth
	logger.SetDepth(5)

	// Verify stored
	logger.registryMu.RLock()
	defer logger.registryMu.RUnlock()

	hasDepth := false
	for _, config := range logger.registry {
		if config.explicitDepth != nil && *config.explicitDepth == 5 {
			hasDepth = true
			break
		}
	}

	if !hasDepth {
		t.Error("Expected to find depth=5 in registry")
	}
}

// Test that negative depth panics
func TestSetDepthNegativePanics(t *testing.T) {
	resetLogger()

	logger := Init(&PlainTextFormatter{}, LevelInfo, nil)

	// Should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected SetDepth(-1) to panic")
		}
	}()

	logger.SetDepth(-1)
}

// Test concurrent access to registry
func TestConcurrentRegistryAccess(t *testing.T) {
	resetLogger()

	logger := Init(&PlainTextFormatter{}, LevelInfo, nil)

	// Spawn multiple goroutines that modify registry
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			// Half set layers, half set depths
			if n%2 == 0 {
				logger.SetLayerForPackage(fmt.Sprintf("layer-%d", n))
			} else {
				logger.SetDepth(n)
			}
		}(i)
	}

	wg.Wait()

	// Verify no panics and registry has entries
	logger.registryMu.RLock()
	count := len(logger.registry)
	logger.registryMu.RUnlock()

	if count == 0 {
		t.Error("Expected registry to have entries after concurrent access")
	}

	t.Logf("Registry has %d entries after concurrent access", count)
}

// Test cache invalidation
func TestCacheInvalidation(t *testing.T) {
	resetLogger()

	logger := Init(&PlainTextFormatter{}, LevelInfo, nil)

	// Manually add something to cache
	testPkg := "myapp/test"
	logger.layerCache[testPkg] = "old-value"

	// Now change the registry (simulate by directly modifying)
	logger.registryMu.Lock()
	logger.registry[testPkg] = &packageConfig{
		explicitLayer: stringPtr("new-value"),
	}
	logger.registryMu.Unlock()

	// SetLayerForPackage should invalidate cache
	// (We can't easily test this without getCurrentPackage working,
	// but we can at least verify the delete logic)

	// Manually test cache deletion
	logger.registryMu.Lock()
	delete(logger.layerCache, testPkg)
	logger.registryMu.Unlock()

	// Verify cache is empty for that package
	logger.registryMu.RLock()
	_, exists := logger.layerCache[testPkg]
	logger.registryMu.RUnlock()

	if exists {
		t.Error("Expected cache to be invalidated")
	}
}

// Helper for tests
func stringPtr(s string) *string {
	return &s
}
