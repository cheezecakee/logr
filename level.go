package logr

type Level int

const (
	LevelInfo Level = iota
	LevelError
	LevelDebug
	LevelWarn
	LevelTest
)

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelDebug:
		return "DEBUG"
	case LevelWarn:
		return "WARN"
	case LevelTest:
		return "TEST"
	default:
		return "UNKNOWN"
	}
}
