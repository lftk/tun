package log

import (
	"fmt"

	"github.com/golang/glog"
)

func Verbose(args ...interface{}) {
	if glog.V(20) {
		glog.InfoDepth(1, args...)
	}
}

func Verbosef(format string, args ...interface{}) {
	if glog.V(20) {
		glog.InfoDepth(1, fmt.Sprintf(format, args...))
	}
}

func VerboseDepth(depth int, args ...interface{}) {
	if glog.V(20) {
		glog.InfoDepth(depth+1, args...)
	}
}

func Debug(args ...interface{}) {
	if glog.V(10) {
		glog.InfoDepth(1, args...)
	}
}

func Debugf(format string, args ...interface{}) {
	if glog.V(10) {
		glog.InfoDepth(1, fmt.Sprintf(format, args...))
	}
}

func DebugDepth(depth int, args ...interface{}) {
	if glog.V(10) {
		glog.InfoDepth(depth+1, args...)
	}
}

func Info(args ...interface{}) {
	glog.InfoDepth(1, args...)
}

func Infof(format string, args ...interface{}) {
	glog.InfoDepth(1, fmt.Sprintf(format, args...))
}

func InfoDepth(depth int, args ...interface{}) {
	glog.InfoDepth(depth+1, args...)
}

func Warning(args ...interface{}) {
	glog.WarningDepth(1, args...)
}

func Warningf(format string, args ...interface{}) {
	glog.WarningDepth(1, fmt.Sprintf(format, args...))
}

func WarningDepth(depth int, args ...interface{}) {
	glog.WarningDepth(depth+1, args...)
}

func Error(args ...interface{}) {
	glog.ErrorDepth(1, args...)
}

func Errorf(format string, args ...interface{}) {
	glog.ErrorDepth(1, fmt.Sprintf(format, args...))
}

func ErrorDepth(depth int, args ...interface{}) {
	glog.ErrorDepth(depth+1, args...)
}

func Fatal(args ...interface{}) {
	glog.FatalDepth(1, args...)
}

func Fatalf(format string, args ...interface{}) {
	glog.FatalDepth(1, fmt.Sprintf(format, args...))
}
