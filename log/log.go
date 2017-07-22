package log

import (
	"fmt"
)

// Logger is used to record different levels of logs.
// Implementing a logger requires that all methods are thread-safe.
type Logger interface {
	Verbose(depth int, args ...interface{})
	Debug(depth int, args ...interface{})
	Info(depth int, args ...interface{})
	Warning(depth int, args ...interface{})
	Error(depth int, args ...interface{})
	Fatal(depth int, args ...interface{})
}

// Default logger
var l Logger

// Use custom logger, Non-thread safe.
func Use(logger Logger) {
	l = logger
}

// Verbose logs detailed log information.
func Verbose(args ...interface{}) {
	if l != nil {
		l.Verbose(1, args...)
	}
}

// Verbosef logs detailed log information.
func Verbosef(format string, args ...interface{}) {
	if l != nil {
		l.Verbose(1, fmt.Sprintf(format, args...))
	}
}

// Debug logs debug level log information.
func Debug(args ...interface{}) {
	if l != nil {
		l.Debug(1, args...)
	}
}

// Debugf logs debug level log information.
func Debugf(format string, args ...interface{}) {
	if l != nil {
		l.Debug(1, fmt.Sprintf(format, args...))
	}
}

// Info logs info level log information.
func Info(args ...interface{}) {
	if l != nil {
		l.Info(1, args...)
	}
}

// Infof logs info level log information.
func Infof(format string, args ...interface{}) {
	if l != nil {
		l.Info(1, fmt.Sprintf(format, args...))
	}
}

// Warning logs warning level log information.
func Warning(args ...interface{}) {
	if l != nil {
		l.Warning(1, args...)
	}
}

// Warningf logs warning level log information.
func Warningf(format string, args ...interface{}) {
	if l != nil {
		l.Warning(1, fmt.Sprintf(format, args...))
	}
}

// Error logs error level log information.
func Error(args ...interface{}) {
	if l != nil {
		l.Error(1, args...)
	}
}

// Errorf logs error level log information.
func Errorf(format string, args ...interface{}) {
	if l != nil {
		l.Error(1, fmt.Sprintf(format, args...))
	}
}

// Fatal logs fatal level log information.
func Fatal(args ...interface{}) {
	if l != nil {
		l.Fatal(1, args...)
	}
}

// Fatalf logs fatal level log information.
func Fatalf(format string, args ...interface{}) {
	if l != nil {
		l.Fatal(1, fmt.Sprintf(format, args...))
	}
}
