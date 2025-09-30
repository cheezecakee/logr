package logr

import (
	"fmt"
	"time"
)

const TimeFormat = time.RFC3339

type Formatter interface {
	Format(entry LogEntry) string
}

type PlainTextFormatter struct{}

func (f *PlainTextFormatter) Format(entry LogEntry) string {
	return fmt.Sprintf("[%s] [%s] [%v] %s", entry.Level, entry.Layer, entry.Timestamp.Format(TimeFormat), entry.Message)
}
