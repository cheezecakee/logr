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

func TestLoggerDebugWarnTest(t *testing.T) {
	// Reset singleton for fresh initialization
	defaultLogger = nil
	once = sync.Once{}

	mock := &MockFormatter{}
	allowedLayers := map[Layer]int{"CORE": 0}

	// Logger level set to Debug
	logger := Init(mock, LevelDebug, allowedLayers)
	logger.SetLayer("CORE")

	// Info should not log (below Debug)
	logger.Info("info message")
	if mock.LastFormatted == "info message" {
		t.Errorf("expected Info not to be logged at LevelDebug")
	}

	// Debug should log
	debugMsg := "debugging"
	logger.Debug(debugMsg)
	if mock.LastFormatted != debugMsg {
		t.Errorf("expected %q, got %q", debugMsg, mock.LastFormatted)
	}

	// Warn should log (higher than Debug)
	warnMsg := "warning!"
	logger.Warn(warnMsg)
	if mock.LastFormatted != warnMsg {
		t.Errorf("expected %q, got %q", warnMsg, mock.LastFormatted)
	}

	// Test should log (highest level)
	testMsg := "test level"
	logger.Test(testMsg)
	if mock.LastFormatted != testMsg {
		t.Errorf("expected %q, got %q", testMsg, mock.LastFormatted)
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

func TestLoggerSetLayer(t *testing.T) {
	// Reset singleton
	defaultLogger = nil
	once = sync.Once{}

	mock := &MockFormatter{}
	allowedLayers := map[Layer]int{"HTTP": 0, "DB": 1}

	logger := Init(mock, LevelInfo, allowedLayers)

	// Switch to a registered layer
	logger.SetLayer("DB")
	if logger.defaultLayer != "DB" {
		t.Errorf("expected defaultLayer to be 'DB', got %q", logger.defaultLayer)
	}

	// Attempt to set an unregistered layer and expect panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when setting unregistered layer")
		}
	}()
	logger.SetLayer("UNKNOWN")
}
