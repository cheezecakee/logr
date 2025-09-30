// Package logr
package logr

import "time"

type LogEntry struct {
	Level     Level
	Layer     Layer
	Message   string
	Timestamp time.Time
	Metadata  *Metadata
}

func NewEntry(level Level, layer Layer, msg string, meta ...Metadata) *LogEntry {
	var metadata *Metadata
	if len(meta) > 0 {
		metadata = &meta[0]
	} else {
		metadata = NewMetadata()
	}
	return &LogEntry{
		Level:     level,
		Layer:     layer,
		Message:   msg,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
}

func (l *LogEntry) AddMetadata(key string, value any) {
	if l.Metadata == nil {
		l.Metadata = NewMetadata()
	}
	l.Metadata.Add(key, value)
}
