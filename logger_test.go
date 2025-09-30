package logr

import (
	"sync"
	"testing"
)

type MockFormatter struct {
	LastFormatted string
}

func (f *MockFormatter) Format(entry LogEntry) string {
	f.LastFormatted = entry.Message
	return entry.Message
}

func TestLoggerInfo(t *testing.T) {
	mock := &MockFormatter{}
	allowedLayers := map[Layer]int{"HTTP": 0}

	// Initialize the logger
	logger := Init(mock, LevelInfo, allowedLayers)
	logger.SetLayer("HTTP")

	msg := "test info message"
	logger.Info(msg)

	if mock.LastFormatted != msg {
		t.Errorf("expected %q, got %q", msg, mock.LastFormatted)
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	defaultLogger = nil
	once = sync.Once{}

	mock := &MockFormatter{}
	allowedLayers := map[Layer]int{"HTTP": 0}

	// Logger level set to Error
	logger := Init(mock, LevelError, allowedLayers)
	logger.SetLayer("HTTP")

	// Info should be ignored
	logger.Info("should not log")
	if mock.LastFormatted != "" {
		t.Errorf("expected no log, got %q", mock.LastFormatted)
	}

	// Error should log
	errMsg := "error occurred"
	logger.Error(errMsg)
	if mock.LastFormatted != errMsg {
		t.Errorf("expected %q, got %q", errMsg, mock.LastFormatted)
	}
}
