package log

import (
	"fmt"

	"github.com/golang/glog"
)

type logger interface {
	Verbose(depth int, args ...interface{})
	Debug(depth int, args ...interface{})
	Info(depth int, args ...interface{})
	Warning(depth int, args ...interface{})
	Error(depth int, args ...interface{})
	Fatal(depth int, args ...interface{})
}

type glogger struct{}

func (l *glogger) Verbose(depth int, args ...interface{}) {
	if glog.V(20) {
		glog.InfoDepth(depth+1, args...)
	}
}

func (l *glogger) Debug(depth int, args ...interface{}) {
	if glog.V(10) {
		glog.InfoDepth(depth+1, args...)
	}
}

func (l *glogger) Info(depth int, args ...interface{}) {
	glog.InfoDepth(depth+1, args...)
}

func (l *glogger) Warning(depth int, args ...interface{}) {
	glog.WarningDepth(depth+1, args...)
}

func (l *glogger) Error(depth int, args ...interface{}) {
	glog.ErrorDepth(depth+1, args...)
}

func (l *glogger) Fatal(depth int, args ...interface{}) {
	glog.FatalDepth(depth+1, args...)
}

var Logger = new(glogger)

func Verbose(args ...interface{}) {
	if Logger != nil {
		Logger.Verbose(1, args...)
	}
}

func Verbosef(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Verbose(1, fmt.Sprintf(format, args...))
	}
}

func Debug(args ...interface{}) {
	if Logger != nil {
		Logger.Debug(1, args...)
	}
}

func Debugf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Debug(1, fmt.Sprintf(format, args...))
	}
}

func Info(args ...interface{}) {
	if Logger != nil {
		Logger.Info(1, args...)
	}
}

func Infof(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Info(1, fmt.Sprintf(format, args...))
	}
}

func Warning(args ...interface{}) {
	if Logger != nil {
		Logger.Warning(1, args...)
	}
}

func Warningf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Warning(1, fmt.Sprintf(format, args...))
	}
}

func Error(args ...interface{}) {
	if Logger != nil {
		Logger.Error(1, args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Error(1, fmt.Sprintf(format, args...))
	}
}

func Fatal(args ...interface{}) {
	if Logger != nil {
		Logger.Fatal(1, args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Fatal(1, fmt.Sprintf(format, args...))
	}
}
