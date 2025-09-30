package logr

type Level int

const (
	LevelDebug Level = iota // 0 - most verbose
	LevelInfo               // 1
	LevelWarn               // 2
	LevelError              // 3
	LevelTest               // 4 - special test level
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelTest:
		return "TEST"
	default:
		return "UNKNOWN"
	}
}
