package impl

import (
	"github.com/golang/glog"
)

// Logger is an implementation of log.Logger.
type Logger struct{}

// Verbose logs detailed log information.
func (l *Logger) Verbose(depth int, args ...interface{}) {
	if glog.V(20) {
		glog.InfoDepth(depth+1, args...)
	}
}

// Debug logs debug level log information.
func (l *Logger) Debug(depth int, args ...interface{}) {
	if glog.V(10) {
		glog.InfoDepth(depth+1, args...)
	}
}

// Info logs info level log information.
func (l *Logger) Info(depth int, args ...interface{}) {
	glog.InfoDepth(depth+1, args...)
}

// Warning logs warning level log information.
func (l *Logger) Warning(depth int, args ...interface{}) {
	glog.WarningDepth(depth+1, args...)
}

// Error logs error level log information.
func (l *Logger) Error(depth int, args ...interface{}) {
	glog.ErrorDepth(depth+1, args...)
}

// Fatal logs fatal level log information.
func (l *Logger) Fatal(depth int, args ...interface{}) {
	glog.FatalDepth(depth+1, args...)
}
