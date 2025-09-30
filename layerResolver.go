package logr

import (
	"runtime"
	"slices"
	"strings"
)

// ResolveLayer Core resolution functions
func resolveLayer(logger *Logger, packagePath string) string {
	cachedLayer, ok := logger.getCachedLayer(packagePath)
	if ok {
		return cachedLayer
	}

	inheritadLayer := findInheritedLayer(logger, packagePath)
	if inheritadLayer != nil {
		logger.setCachedLayer(packagePath, *inheritadLayer)
		return *inheritadLayer
	}

	logger.registryMu.RLock()
	depthValue := logger.config.DefaultDepth

	if logger.registry[packagePath] != nil && logger.registry[packagePath].explicitDepth != nil {
		depthValue = *logger.registry[packagePath].explicitDepth
	}

	logger.registryMu.RUnlock()

	result := extractFromDepth(packagePath, depthValue, logger.config.SkipSegments)

	logger.setCachedLayer(packagePath, result)

	return result
}

func getCurrentPackage(skip int) string {
	// Get program counter of caller
	// skip: how many stack frames to skip
	//   0 = GetCurrentPackage itself
	//   1 = function that called GetCurrentPackage
	//   2 = function that called that function, etc.

	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown" // Couldn't get caller
	}

	// Get function info from program counter
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	// Function name format: "github.com/user/pkg/subpkg.FuncName"
	// or with receiver: "github.com/user/pkg.(*Type).Method"
	fullName := fn.Name()

	// fmt.Printf("DEBUG: fullName = %s, skip = %d\n", fullName, skip)
	// Extract package path (everything before last dot)
	// "github.com/user/pkg.FuncName" -> "github.com/user/pkg"
	lastDot := strings.LastIndex(fullName, ".")
	if lastDot == -1 {
		return "unknown"
	}

	packagePath := fullName[:lastDot]

	// Clean up method receivers: "pkg.(*Type)" -> "pkg"
	if idx := strings.Index(packagePath, ".("); idx != -1 {
		packagePath = packagePath[:idx]
	}

	return packagePath
}

// Finding the right skip value:
// GetCurrentPackage(3) typical for:
//
//	[0] runtime.Caller
//	[1] GetCurrentPackage
//	[2] resolveLayer or Log
//	[3] Info/Error/Debug <- actual caller we want
func extractFromDepth(packagePath string, depth int, skipSegments []string) string {
	// Split path: "a/b/c/d" -> ["a", "b", "c", "d"]
	segments := strings.Split(packagePath, "/")

	// Safety: ensure depth is valid
	if depth < 0 {
		depth = 0
	}
	if depth > len(segments) {
		depth = len(segments)
	}

	// Take LAST N segments
	// Example: ["github.com", "myapp", "internal", "db", "postgres"]
	//          depth=2 → take last 2 → ["db", "postgres"]
	startIndex := len(segments) - depth
	if startIndex < 0 {
		startIndex = 0
	}
	relevant := segments[startIndex:]

	// Filter out skipped segments
	filtered := []string{}
	for _, seg := range relevant {
		if !slices.Contains(skipSegments, seg) {
			filtered = append(filtered, seg)
		}
	}

	// Handler empty result
	if len(filtered) == 0 {
		return "UNKNOWN"
	}

	// Join segments and conver to uppercase
	// ["db", "postgres"] → "db/postgres" → "DB/POSTGRES"
	result := strings.Join(filtered, "/")
	return strings.ToUpper(result)
}

func findInheritedLayer(logger *Logger, packagePath string) *string {
	logger.registryMu.RLock()
	defer logger.registryMu.RUnlock()

	current := packagePath

	for current != "" {
		// Check if current package has explicit layer
		if logger.registry[current] != nil && logger.registry[current].explicitLayer != nil {
			return logger.registry[current].explicitLayer
		}

		// Move to parent package
		current = parentPath(current)
	}
	return nil
}

func parentPath(path string) string {
	lastIndex := strings.LastIndex(path, "/")
	if lastIndex == -1 {
		return ""
	}
	return path[:lastIndex]
}

// GetCachedLayer optional caching
func (l *Logger) getCachedLayer(pkgPath string) (string, bool) {
	l.registryMu.Lock()
	defer l.registryMu.Unlock()

	cachedValue, ok := l.layerCache[pkgPath]
	if ok {
		return cachedValue, true
	}

	return "", false
}

func (l *Logger) setCachedLayer(pkgPath string, layer string) {
	l.registryMu.Lock()
	defer l.registryMu.Unlock()

	l.layerCache[pkgPath] = layer
}
