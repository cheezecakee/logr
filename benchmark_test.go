package logr

import (
	"fmt"
	"testing"
)

// ============================================================================
// Benchmarks - Performance Testing
// ============================================================================

// BenchmarkGetCurrentPackage measures the cost of detecting calling package
func BenchmarkGetCurrentPackage(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = getCurrentPackage(3)
	}
}

// BenchmarkExtractFromDepth measures path extraction performance
func BenchmarkExtractFromDepth(b *testing.B) {
	packagePath := "github.com/myapp/internal/api/handlers"
	skipSegments := []string{"internal", "pkg"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = extractFromDepth(packagePath, 2, skipSegments)
	}
}

// BenchmarkExtractFromDepthNoSkip measures extraction without filtering
func BenchmarkExtractFromDepthNoSkip(b *testing.B) {
	packagePath := "github.com/myapp/internal/api/handlers"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = extractFromDepth(packagePath, 2, nil)
	}
}

// BenchmarkParentPath measures parent path extraction
func BenchmarkParentPath(b *testing.B) {
	path := "github.com/myapp/internal/api/handlers"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = parentPath(path)
	}
}

// BenchmarkFindInheritedLayer measures inheritance lookup performance
func BenchmarkFindInheritedLayer(b *testing.B) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
	})

	// Set up inheritance chain
	layer := "Database"
	logger.registryMu.Lock()
	logger.registry["github.com/myapp/db"] = &packageConfig{
		explicitLayer: &layer,
	}
	logger.registryMu.Unlock()

	childPkg := "github.com/myapp/db/postgres/connection"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = findInheritedLayer(logger, childPkg)
	}
}

// BenchmarkResolveLayerCached measures resolution with cache hit
func BenchmarkResolveLayerCached(b *testing.B) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
		SkipSegments: []string{"internal"},
	})

	packagePath := "github.com/myapp/api/handlers"

	// Prime the cache
	_ = resolveLayer(logger, packagePath)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = resolveLayer(logger, packagePath)
	}
}

// BenchmarkResolveLayerUncached measures resolution without cache
func BenchmarkResolveLayerUncached(b *testing.B) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
		SkipSegments: []string{"internal"},
	})

	packagePath := "github.com/myapp/api/handlers"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear cache before each iteration to simulate cold start
		logger.registryMu.Lock()
		delete(logger.layerCache, packagePath)
		logger.registryMu.Unlock()

		_ = resolveLayer(logger, packagePath)
	}
}

// BenchmarkResolveLayerWithInheritance measures resolution with parent lookup
func BenchmarkResolveLayerWithInheritance(b *testing.B) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
	})

	// Set up parent
	parentLayer := "Database"
	logger.registryMu.Lock()
	logger.registry["github.com/myapp/db"] = &packageConfig{
		explicitLayer: &parentLayer,
	}
	logger.registryMu.Unlock()

	childPkg := "github.com/myapp/db/postgres"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear cache to force inheritance lookup
		logger.registryMu.Lock()
		delete(logger.layerCache, childPkg)
		logger.registryMu.Unlock()

		_ = resolveLayer(logger, childPkg)
	}
}

// BenchmarkLoggerInfo measures end-to-end logging performance
func BenchmarkLoggerInfo(b *testing.B) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("test message")
	}
}

// BenchmarkLoggerInfoWithMetadata measures logging with metadata
func BenchmarkLoggerInfoWithMetadata(b *testing.B) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		meta := NewMetadata()
		meta.Add("requestID", "abc123")
		meta.Add("userID", 456)

		entry := NewEntry(LevelInfo, LayerHTTP, "request processed", *meta)
		_ = logger.formatter.Format(*entry)
	}
}

// BenchmarkPlainTextFormatter measures formatting performance
func BenchmarkPlainTextFormatter(b *testing.B) {
	formatter := &PlainTextFormatter{}
	entry := NewEntry(LevelInfo, LayerHTTP, "test message")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = formatter.Format(*entry)
	}
}

// BenchmarkJSONFormatter measures JSON formatting performance
func BenchmarkJSONFormatter(b *testing.B) {
	formatter := &JSONFormatter{}
	entry := NewEntry(LevelInfo, LayerHTTP, "test message")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = formatter.Format(*entry)
	}
}

// BenchmarkFormatterWithMetadata compares formatters with metadata
func BenchmarkFormatterWithMetadata(b *testing.B) {
	b.Run("PlainText", func(b *testing.B) {
		formatter := &PlainTextFormatter{}
		meta := NewMetadata()
		meta.Add("requestID", "abc123")
		meta.Add("userID", 456)
		entry := NewEntry(LevelInfo, LayerHTTP, "test message", *meta)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = formatter.Format(*entry)
		}
	})

	b.Run("JSON", func(b *testing.B) {
		formatter := &JSONFormatter{}
		meta := NewMetadata()
		meta.Add("requestID", "abc123")
		meta.Add("userID", 456)
		entry := NewEntry(LevelInfo, LayerHTTP, "test message", *meta)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = formatter.Format(*entry)
		}
	})
}

// BenchmarkCacheOperations measures cache get/set performance
func BenchmarkCacheOperations(b *testing.B) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
	})

	b.Run("Set", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			logger.setCachedLayer(fmt.Sprintf("pkg-%d", i), "LAYER")
		}
	})

	b.Run("Get", func(b *testing.B) {
		// Prime cache
		logger.setCachedLayer("test-pkg", "TEST_LAYER")

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = logger.getCachedLayer("test-pkg")
		}
	})
}

// BenchmarkRegistryOperations measures registry performance
func BenchmarkRegistryOperations(b *testing.B) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
	})

	b.Run("SetLayerForPackage", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			logger.SetLayerForPackage(fmt.Sprintf("layer-%d", i))
		}
	})

	b.Run("SetDepth", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			logger.SetDepth(i % 5) // Vary depth 0-4
		}
	})
}

// BenchmarkConfigValidation measures config validation overhead
func BenchmarkConfigValidation(b *testing.B) {
	config := Config{
		DefaultDepth:  2,
		SkipSegments:  []string{"internal", "pkg"},
		StrictMode:    true,
		AllowedLayers: []Layer{LayerHTTP, LayerDB, LayerCORE},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

// BenchmarkShouldSkipSegment measures segment filtering
func BenchmarkShouldSkipSegment(b *testing.B) {
	config := Config{
		SkipSegments: []string{"internal", "pkg", "adapters", "primary", "secondary"},
	}

	b.Run("Hit", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = config.ShouldSkipSegment("internal")
		}
	})

	b.Run("Miss", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = config.ShouldSkipSegment("api")
		}
	})
}

// BenchmarkNewEntry measures log entry creation
func BenchmarkNewEntry(b *testing.B) {
	b.Run("WithoutMetadata", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = NewEntry(LevelInfo, LayerHTTP, "test message")
		}
	})

	b.Run("WithMetadata", func(b *testing.B) {
		meta := NewMetadata()
		meta.Add("key", "value")

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = NewEntry(LevelInfo, LayerHTTP, "test message", *meta)
		}
	})
}

// BenchmarkConcurrentResolution measures concurrent layer resolution
func BenchmarkConcurrentResolution(b *testing.B) {
	resetLogger()

	logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
		DefaultDepth: 2,
		SkipSegments: []string{"internal"},
	})

	packagePaths := []string{
		"github.com/myapp/api/handlers",
		"github.com/myapp/db/postgres",
		"github.com/myapp/cache/redis",
		"github.com/myapp/services/auth",
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			pkg := packagePaths[i%len(packagePaths)]
			_ = resolveLayer(logger, pkg)
			i++
		}
	})
}

// BenchmarkEndToEndLogging measures complete logging pipeline
func BenchmarkEndToEndLogging(b *testing.B) {
	b.Run("PlainText", func(b *testing.B) {
		resetLogger()

		logger := InitWithConfig(&PlainTextFormatter{}, LevelInfo, Config{
			DefaultDepth: 2,
		})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message")
		}
	})

	b.Run("JSON", func(b *testing.B) {
		resetLogger()

		logger := InitWithConfig(&JSONFormatter{}, LevelInfo, Config{
			DefaultDepth: 2,
		})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message")
		}
	})
}

// ============================================================================
// Comparative Benchmarks
// ============================================================================

// BenchmarkCompareDepthStrategies compares different depth values
func BenchmarkCompareDepthStrategies(b *testing.B) {
	packagePath := "github.com/myapp/internal/api/v1/handlers"

	depths := []int{1, 2, 3, 4}

	for _, depth := range depths {
		b.Run(fmt.Sprintf("Depth-%d", depth), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = extractFromDepth(packagePath, depth, nil)
			}
		})
	}
}

// BenchmarkCompareSkipSegmentSizes compares different skip list sizes
func BenchmarkCompareSkipSegmentSizes(b *testing.B) {
	packagePath := "github.com/myapp/internal/pkg/adapters/api/handlers"

	skipLists := map[string][]string{
		"None":   nil,
		"Small":  {"internal"},
		"Medium": {"internal", "pkg", "adapters"},
		"Large":  {"internal", "pkg", "adapters", "primary", "secondary", "cmd"},
	}

	for name, skipList := range skipLists {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = extractFromDepth(packagePath, 3, skipList)
			}
		})
	}
}
