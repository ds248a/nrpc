package log

import (
	"fmt"
	"os"
	"time"
)

var (
	TimeFormat           = "2006/01/02 15:04:05.000"
	Output               = os.Stdout
	DefaultLogger Logger = &logger{level: LevelInfo}
)

const (
	LevelAll = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelNone
)

// Logger defines log interface
type Logger interface {
	SetLevel(lvl int)
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

// Sets default logger.
func SetLogger(l Logger) {
	DefaultLogger = l
}

// Sets default logger's priority.
func SetLevel(lvl int) {
	switch lvl {
	case LevelAll, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelNone:
		DefaultLogger.SetLevel(lvl)
		break
	default:
		fmt.Fprintf(Output, "invalid log level: %v", lvl)
	}
}

// Logger by default.
type logger struct {
	level int
}

// Sets logs priority.
func (l *logger) SetLevel(lvl int) {
	switch lvl {
	case LevelAll, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelNone:
		l.level = lvl
		break
	default:
		fmt.Fprintf(Output, "invalid log level: %v", lvl)
	}
}

// Uses fmt.Printf to log a message at LevelDebug.
func (l *logger) Debug(format string, v ...interface{}) {
	if LevelDebug >= l.level {
		fmt.Fprintf(Output, time.Now().Format(TimeFormat)+" [DBG] "+format+"\n", v...)
	}
}

// Uses fmt.Printf to log a message at LevelInfo.
func (l *logger) Info(format string, v ...interface{}) {
	if LevelInfo >= l.level {
		fmt.Fprintf(Output, time.Now().Format(TimeFormat)+" [INF] "+format+"\n", v...)
	}
}

// Warn uses fmt.Printf to log a message at LevelWarn.
func (l *logger) Warn(format string, v ...interface{}) {
	if LevelWarn >= l.level {
		fmt.Fprintf(Output, time.Now().Format(TimeFormat)+" [WRN] "+format+"\n", v...)
	}
}

// Uses fmt.Printf to log a message at LevelError.
func (l *logger) Error(format string, v ...interface{}) {
	if LevelError >= l.level {
		fmt.Fprintf(Output, time.Now().Format(TimeFormat)+" [ERR] "+format+"\n", v...)
	}
}

// Uses DefaultLogger to log a message at LevelDebug.
func Debug(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Debug(format, v...)
	}
}

// Uses DefaultLogger to log a message at LevelInfo.
func Info(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Info(format, v...)
	}
}

// Uses DefaultLogger to log a message at LevelWarn.
func Warn(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Warn(format, v...)
	}
}

// Uses DefaultLogger to log a message at LevelError.
func Error(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Error(format, v...)
	}
}
