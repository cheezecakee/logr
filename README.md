# logr

A lightweight, layer-aware, structured logging library for Go with automatic package detection and zero-configuration defaults.

## Features

- **Zero Configuration**: Works out of the box with sensible defaults
- **Automatic Layer Detection**: Logs are automatically tagged with the package they originate from
- **Hierarchical Layers**: Supports nested package structures (`API/HANDLERS`, `DB/POSTGRES`)
- **Flexible Overrides**: Customize layers and extraction depth per package
- **Multiple Formatters**: Plain text and JSON formatters included
- **Log Levels**: Debug, Info, Warn, Error, and Test
- **Structured Metadata**: Attach key-value pairs to any log entry
- **Thread-Safe**: Safe for concurrent use across goroutines
- **Performance**: Built-in caching for fast repeated resolutions

---

## Installation

```bash
go get github.com/cheezecakee/logr
```

---

## Quick Start

```go
package main

import "github.com/cheezecakee/logr"

func main() {
    // Initialize with plain text formatter at Info level
    logr.Init(&logr.PlainTextFormatter{}, logr.LevelInfo, nil)
    
    // Start logging - layer is automatically detected
    logr.Get().Info("Application started")
    logr.Get().Error("Something went wrong")
}
```

**Output:**
```
[INFO] [MAIN] [2025-09-30T19:12:02-03:00] Application started
[ERROR] [MAIN] [2025-09-30T19:12:03-03:00] Something went wrong
```

---

## How It Works

### Automatic Layer Detection

The logger automatically detects which package is logging based on your project structure:

```
myproject/
  internal/
    api/
      handlers/
        user.go      → Logs as [API/HANDLERS]
    db/
      postgres/
        connection.go → Logs as [DB/POSTGRES]
```

**In `user.go`:**
```go
package handlers

import "github.com/yourusername/myproject/logr"

func GetUser(id string) {
    logr.Get().Info("Fetching user")  // → [INFO] [API/HANDLERS] Fetching user
}
```

### Configuration

Customize the logger behavior with `InitWithConfig`:

```go
config := logr.Config{
    DefaultDepth: 2,                              // Take last 2 segments of package path
    SkipSegments: []string{"internal", "pkg"},    // Filter out these segments
    StrictMode: false,                            // Allow any layers
}

logr.InitWithConfig(&logr.PlainTextFormatter{}, logr.LevelInfo, config)
```

**How DefaultDepth works:**
```
Package path: github.com/myapp/internal/api/handlers
DefaultDepth: 2 → Takes last 2 segments → [API/HANDLERS]
DefaultDepth: 1 → Takes last 1 segment  → [HANDLERS]
```

**How SkipSegments works:**
```
Package path: github.com/myapp/internal/api/handlers
SkipSegments: ["internal"]
Result: [API/HANDLERS] (internal is filtered out)
```

---

## Advanced Usage

### Custom Layer Names

Override the automatic layer detection for a specific package:

```go
package database

import "github.com/yourusername/myproject/logr"

func init() {
    logr.Get().SetLayerForPackage("DataLayer")
}

func Connect() {
    logr.Get().Info("Connecting...")  // → [INFO] [DataLayer] Connecting...
}
```

### Custom Depth

Control how many path segments to include:

```go
package handlers

import "github.com/yourusername/myproject/logr"

func init() {
    logr.Get().SetDepth(1)  // Only use last segment
}

func Handle() {
    logr.Get().Info("Handling request")  // → [INFO] [HANDLERS] Handling request
}
```

### Layer Inheritance

Child packages inherit parent configurations:

```go
// In db/db.go
func init() {
    logr.Get().SetLayerForPackage("Database")
}

// In db/postgres/postgres.go - inherits "Database" layer
func Query() {
    logr.Get().Info("Executing query")  // → [INFO] [Database] Executing query
}
```

### Metadata

Attach structured data to log entries:

```go
meta := logr.NewMetadata()
meta.Add("userID", 12345)
meta.Add("action", "login")

entry := logr.NewEntry(logr.LevelInfo, logr.LayerHTTP, "User logged in", *meta)
formatted := logger.formatter.Format(*entry)
```

### JSON Formatter

Use JSON output for structured logging:

```go
logr.Init(&logr.JSONFormatter{}, logr.LevelInfo, nil)

logr.Get().Info("Request processed")
```

**Output:**
```json
{"level":"INFO","layer":"API","message":"Request processed","timestamp":"2025-09-30T19:12:02-03:00"}
```

---

## Log Levels

Levels are ordered from most to least verbose:

```go
logr.LevelDebug  // 0 - Most verbose
logr.LevelInfo   // 1
logr.LevelWarn   // 2
logr.LevelError  // 3
logr.LevelTest   // 4 - Special test level
```

Set the minimum level when initializing:

```go
logr.Init(&logr.PlainTextFormatter{}, logr.LevelWarn, nil)

logger.Debug("This won't log")  // Below Warn
logger.Info("This won't log")   // Below Warn
logger.Warn("This will log")    // At Warn level
logger.Error("This will log")   // Above Warn
```

---

## Predefined Layers

Common layers are provided as constants:

```go
logr.LayerHTTP   // "HTTP"
logr.LayerDB     // "DB"
logr.LayerCORE   // "CORE"
```

Create custom layers:

```go
customLayer := logr.RegisterLayer("cache")  // Returns Layer("CACHE")
```

---

## Strict Mode

Restrict logging to specific layers:

```go
config := logr.Config{
    DefaultDepth: 2,
    StrictMode: true,
    AllowedLayers: []logr.Layer{
        logr.LayerHTTP,
        logr.LayerDB,
        logr.LayerCORE,
    },
}

logr.InitWithConfig(&logr.PlainTextFormatter{}, logr.LevelInfo, config)

// Only HTTP, DB, and CORE layers are allowed
// Other layers will cause validation errors
```

---

## Thread Safety

All operations are thread-safe:

```go
func handler1() {
    logr.Get().Info("Request 1")
}

func handler2() {
    logr.Get().Info("Request 2")
}

// Safe to call concurrently
go handler1()
go handler2()
```

---

## Performance

The logger includes intelligent caching:

1. **First call**: Detects package, resolves layer, caches result
2. **Subsequent calls**: Returns cached layer (very fast)
3. **After SetLayer/SetDepth**: Cache is invalidated automatically

Run benchmarks:

```bash
go test -bench=. -benchmem
```

---

## API Reference

### Initialization

```go
// Simple initialization
Init(formatter Formatter, level Level, allowedLayers map[Layer]int) *Logger

// With configuration
InitWithConfig(formatter Formatter, level Level, config Config) *Logger

// Get singleton instance
Get() *Logger
```

### Logging Methods

```go
logger.Debug(msg string)
logger.Info(msg string)
logger.Warn(msg string)
logger.Error(msg string)
logger.Test(msg string)
```

### Configuration Methods

```go
// Set custom layer for calling package
SetLayerForPackage(layer string)

// Set custom depth for calling package
SetDepth(depth int)
```

### Formatters

```go
&PlainTextFormatter{}  // Human-readable format
&JSONFormatter{}       // Machine-readable JSON
```

---

## Project Structure Example

```
myproject/
├── cmd/
│   └── server/
│       └── main.go           → [SERVER] or [MAIN]
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   └── user.go       → [API/HANDLERS]
│   │   └── middleware/
│   │       └── auth.go       → [API/MIDDLEWARE]
│   ├── db/
│   │   ├── postgres/
│   │   │   └── client.go     → [DB/POSTGRES]
│   │   └── redis/
│   │       └── cache.go      → [DB/REDIS]
│   └── services/
│       └── auth/
│           └── jwt.go        → [SERVICES/AUTH]
└── go.mod
```

With `DefaultDepth: 2` and `SkipSegments: ["internal"]`, each package gets a meaningful, hierarchical layer automatically.

---

## Examples

See the `example/` directory for complete working examples.

---

## Contributing

Contributions welcome! Please open an issue or PR.

---

## License

MIT © Anabel
