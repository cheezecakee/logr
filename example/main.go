// example/main.go
package main

import (
	"github.com/cheezecakee/logr"
	"github.com/cheezecakee/logr/example/handlers"
)

func main() {
	logr.Init(&logr.PlainTextFormatter{}, logr.LevelInfo, nil)
	logr.Get().Info("Application started")
	handlers.HandleUser()
}
