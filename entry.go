// Package logr
package logr

import "time"

type LogEntry struct {
	Level     Level
	Layer     Layer
	Message   string
	Timestamp time.Time
}

func NewEntry(level Level, layer Layer, msg string) *LogEntry {
	return &LogEntry{
		Level:     level,
		Layer:     layer,
		Message:   msg,
		Timestamp: time.Now(),
	}
}
