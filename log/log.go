package log

import (
	"fmt"
)

var (
	FnInfo    func(depth int, args ...interface{})
	FnWarning func(depth int, args ...interface{})
	FnError   func(depth int, args ...interface{})
	FnFatal   func(depth int, args ...interface{})
)

func Info(args ...interface{}) {
	if FnInfo != nil {
		FnInfo(1, args...)
	}
}

func Infof(format string, args ...interface{}) {
	if FnInfo != nil {
		FnInfo(1, fmt.Sprintf(format, args...))
	}
}

func Warning(args ...interface{}) {
	if FnWarning != nil {
		FnWarning(1, args...)
	}
}

func Warningf(format string, args ...interface{}) {
	if FnWarning != nil {
		FnWarning(1, fmt.Sprintf(format, args...))
	}
}

func Error(args ...interface{}) {
	if FnError != nil {
		FnError(1, args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if FnError != nil {
		FnError(1, fmt.Sprintf(format, args...))
	}
}

func Fatal(args ...interface{}) {
	if FnFatal != nil {
		FnFatal(1, args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	if FnFatal != nil {
		FnFatal(1, fmt.Sprintf(format, args...))
	}
}
