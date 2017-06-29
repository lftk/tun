package log

import (
	"fmt"
)

type logger interface {
	Verbose(depth int, args ...interface{})
	Debug(depth int, args ...interface{})
	Info(depth int, args ...interface{})
	Warning(depth int, args ...interface{})
	Error(depth int, args ...interface{})
	Fatal(depth int, args ...interface{})
}

var Logger logger

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
