package logr

import (
	"fmt"
	"sync"
)

var once sync.Once

type Logger struct {
	formatter     Formatter
	level         Level
	defaultLayer  Layer
	allowedLayers map[Layer]int
	mu            sync.Mutex
}

var defaultLogger *Logger

func Init(formatter Formatter, level Level, allowedLayers map[Layer]int) *Logger {
	once.Do(func() {
		fmt.Println("Creating single Logger instance now.")
		defaultLogger = &Logger{
			formatter:     formatter,
			level:         level,
			allowedLayers: allowedLayers,
		}
	})
	return defaultLogger
}

func Get() *Logger {
	if defaultLogger == nil {
		panic("Logger not initialized: call Init() before Get()")
	}
	return defaultLogger
}

func (l *Logger) SetLayer(layer Layer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.allowedLayers[layer]; !ok {
		panic("Layer not found: create a new layer RegisterLayer()")
	} else {
		l.defaultLayer = layer
	}
}

func (l *Logger) Info(msg string) {
	l.Log(LevelInfo, msg)
}

func (l *Logger) Error(msg string) {
	l.Log(LevelError, msg)
}

func (l *Logger) Debug(msg string) {
	l.Log(LevelDebug, msg)
}

func (l *Logger) Warn(msg string) {
	l.Log(LevelWarn, msg)
}

func (l *Logger) Test(msg string) {
	l.Log(LevelTest, msg)
}

func (l *Logger) Log(level Level, msg string) {
	if l.level <= level {
		entry := NewEntry(level, l.defaultLayer, msg)
		formatted := l.formatter.Format(*entry)
		fmt.Println(formatted)
	}
}
