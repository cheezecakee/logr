package logr

import (
	"fmt"
	"sync"
)

var once sync.Once

const (
	skipForSetMethods = 3 // SetLayerForPackage/SetDepth → user code
	skipForLogging    = 4 // Info/Error/etc → log → getOrResolveLayer → getCurrentPackage → user
)

type Logger struct {
	formatter     Formatter
	level         Level
	defaultLayer  Layer
	allowedLayers map[Layer]int

	config     Config
	registry   map[string]*packageConfig
	layerCache map[string]string
	registryMu sync.RWMutex

	mu sync.Mutex
}

var defaultLogger *Logger

func Init(formatter Formatter, level Level, allowedLayers map[Layer]int) *Logger {
	once.Do(func() {
		defaultLogger = &Logger{
			formatter:     formatter,
			level:         level,
			allowedLayers: allowedLayers,

			config:     DefaultConfig(),
			registry:   make(map[string]*packageConfig),
			layerCache: make(map[string]string),
		}
	})
	return defaultLogger
}

func Get() *Logger {
	if defaultLogger == nil {
		panic("Logger not initialized: call Init() before Get()")
	}
	return defaultLogger
}

func (l *Logger) SetLayer(layer Layer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.allowedLayers[layer]; !ok {
		panic("Layer not found: create a new layer RegisterLayer()")
	} else {
		l.defaultLayer = layer
	}
}

func (l *Logger) Info(msg string) {
	l.log(LevelInfo, msg)
}

func (l *Logger) Error(msg string) {
	l.log(LevelError, msg)
}

func (l *Logger) Debug(msg string) {
	l.log(LevelDebug, msg)
}

func (l *Logger) Warn(msg string) {
	l.log(LevelWarn, msg)
}

func (l *Logger) Test(msg string) {
	l.log(LevelTest, msg)
}

// Dynamic context

func (l *Logger) Errorf(format string, args ...any) {
	l.log(LevelError, fmt.Sprintf(format, args...))
}

func (l *Logger) Infof(format string, args ...any) {
	l.log(LevelInfo, fmt.Sprintf(format, args...))
}

func (l *Logger) Debugf(format string, args ...any) {
	l.log(LevelDebug, fmt.Sprintf(format, args...))
}

func (l *Logger) Warnf(format string, args ...any) {
	l.log(LevelWarn, fmt.Sprintf(format, args...))
}

func (l *Logger) log(level Level, msg string) {
	if l.level <= level {
		layerStr := l.getOrResolveLayer()
		layer := Layer(layerStr)

		entry := NewEntry(level, layer, msg)
		formatted := l.formatter.Format(*entry)
		fmt.Println(formatted)
	}
}

func InitWithConfig(formatter Formatter, level Level, config Config) *Logger {
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid config: %v", err))
	}

	once.Do(func() {
		defaultLogger = &Logger{
			formatter: formatter,
			level:     level,

			config:     config,
			registry:   make(map[string]*packageConfig),
			layerCache: make(map[string]string),

			// Note: allowedLayers comes from config.allowedLayers
			allowedLayers: make(map[Layer]int),
		}

		// If useing StrictMode, populate allowedLayers from config
		if config.StrictMode {
			for _, layer := range config.AllowedLayers {
				defaultLogger.allowedLayers[layer] = 1
			}
		}
	})
	return defaultLogger
}

// SetLayerForPackage stores a custom layer name for a specific package.
// This is called by the user at the top of their package file.
func (l *Logger) SetLayerForPackage(layer string) {
	// Detect which package is calling this function
	// We skip 2 frames: [0]=runtime.Caller, [1]=getCurrentPackage, [2]=SetLayerForPackage, [3]=actual caller
	packagePath := getCurrentPackage(skipForSetMethods)

	// Thread-safe write to registry
	l.registryMu.Lock()
	defer l.registryMu.Unlock()

	// Get or create config for this package
	if l.registry[packagePath] == nil {
		l.registry[packagePath] = &packageConfig{}
	}

	// Store the layer name
	l.registry[packagePath].explicitLayer = &layer

	// Invalidate cache for this package (it needs to be recalculated)
	delete(l.layerCache, packagePath)
}

// SetDepth sets a custom depth for layer extraction in the calling package.
// Unlike SetLayerForPackage, this does NOT inherit to child packages.
func (l *Logger) SetDepth(depth int) {
	// Validate depth
	if depth < 0 {
		panic(fmt.Sprintf("SetDepth: depth must be >= 0, got %d", depth))
	}

	// Detect calling package
	packagePath := getCurrentPackage(skipForSetMethods)

	// Thread-safe write
	l.registryMu.Lock()
	defer l.registryMu.Unlock()

	// Get or create config
	if l.registry[packagePath] == nil {
		l.registry[packagePath] = &packageConfig{}
	}

	// Store the depth
	l.registry[packagePath].explicitDepth = &depth

	// Invalidate cache
	delete(l.layerCache, packagePath)
}

// GetOrResolveLayer resolves the layer for the calling package.
// This is an internal helper used by Log() method.
func (l *Logger) getOrResolveLayer() string {
	// Detect calling package (adjust skip as needed based on call stack)
	packagePath := getCurrentPackage(skipForLogging)

	// fmt.Printf("DEBUG: Detected package: %s\n", packagePath) // Add this temporarily

	// Try to resolve the layer
	// (We'll implement resolveLayer in Phase 2, for now return placeholder)
	layer := resolveLayer(l, packagePath)

	return layer
}
