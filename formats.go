package logr

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const TimeFormat = time.RFC3339

type Formatter interface {
	Format(entry LogEntry) string
}

type PlainTextFormatter struct{}

func (f *PlainTextFormatter) Format(entry LogEntry) string {
	baseStr := fmt.Sprintf("[%s] [%s] [%v] %s", entry.Level, entry.Layer, entry.Timestamp.Format(TimeFormat), entry.Message)

	if entry.Metadata != nil && len(entry.Metadata.Data) > 0 {
		var metadataStr []string
		for key, value := range entry.Metadata.Data {
			metadataStr = append(metadataStr, fmt.Sprintf("%s=%v", key, value))
		}
		metadataJoined := strings.Join(metadataStr, " ")
		baseStr = baseStr + " " + metadataJoined
	}
	return baseStr
}

type JSONFormatter struct{}

func (f JSONFormatter) Format(entry LogEntry) string {
	jsonLogEntry := struct {
		Level     string    `json:"level"`
		Layer     string    `json:"layer"`
		Message   string    `json:"message"`
		Timestamp string    `json:"timestamp"`
		Metadata  *Metadata `json:"metadata,omitempty"`
	}{
		Level:     entry.Level.String(),
		Layer:     entry.Layer.String(),
		Message:   entry.Message,
		Timestamp: entry.Timestamp.Format(TimeFormat),
		Metadata:  nil,
	}

	if entry.Metadata != nil && len(entry.Metadata.Data) > 0 {
		jsonLogEntry.Metadata = entry.Metadata
	}

	jsonEntry, err := json.Marshal(&jsonLogEntry)
	if err != nil {
		fmt.Printf("failed to encode entry: %s", err)
		return ""
	}

	return string(jsonEntry)
}
