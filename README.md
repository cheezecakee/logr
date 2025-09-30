# logr
**logr** is a lightweight, layer-aware, structured logging library for Go.
It provides:
- Singleton global logger
- Pluggable formatters (Text, JSON, custom)
- Layer-aware logging (HTTP, Core, DB, etc.)
- Log levels including Debug, Info, Warn, Error, and Test
- Optional metadata fields per log entry

---

## Installation 
```bash
go get github.com/yourusername/logr
```

---

## Basic Usage
```go
package main

import (
    "github.com/yourusername/logr"
)

func main() {
    // Initialize logger (TextFormatter, Info level, allowed layers)
    logr.Init(logr.TextFormatter{}, logr.LevelInfo, []string{"HTTP","Core","DB"})

    // Set default layer for this package
    logr.SetLayer("HTTP")

    // Log messages
    logr.Get().Info("Server started")
    logr.Get().Error("Failed to insert workout")
}
```
**Output (TextFormatter):**
```less
[INFO] [HTTP] 2025-09-29T15:23:45Z Server started
[ERROR] [HTTP] 2025-09-29T15:23:46Z Failed to insert workout
```

---

## Planned Features
- JSON and custom formatters
- Layer inheritance for subpackages
- Optional per-log metadata fields
- Dynamic log level filtering
- Multiple output destinations (stdout, file, remote)

## License 
MIT © Anabel
