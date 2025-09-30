package logr

import (
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

	if entry.Metadata != nil && len(entry.Metadata.data) > 0 {
		var metadataStr []string
		for key, value := range entry.Metadata.data {
			metadataStr = append(metadataStr, fmt.Sprintf("%s=%v", key, value))
		}
		metadataJoined := strings.Join(metadataStr, " ")
		baseStr = baseStr + " " + metadataJoined
	}
	return baseStr
}
