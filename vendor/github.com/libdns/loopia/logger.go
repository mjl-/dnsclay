package loopia

import (
	"sync"
)

type iLogger interface {
	Errorw(msg string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatal(args ...interface{})
	Infof(format string, args ...interface{})
	Info(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnw(msg string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugw(msg string, args ...interface{})
}

type loggerWrapper struct {
}

func (logger *loggerWrapper) Errorw(msg string, args ...interface{}) {
	// noop
}
func (logger *loggerWrapper) Errorf(format string, args ...interface{}) {
	// noop
}
func (logger *loggerWrapper) Fatalf(format string, args ...interface{}) {
	// noop
}
func (logger *loggerWrapper) Fatal(args ...interface{}) {
	// noop
}
func (logger *loggerWrapper) Info(args ...interface{}) {
	// noop
}
func (logger *loggerWrapper) Infof(format string, args ...interface{}) {
	// noop
}
func (logger *loggerWrapper) Warnf(format string, args ...interface{}) {
	// noop
}
func (logger *loggerWrapper) Warnw(msg string, args ...interface{}) {
	// noop
}
func (logger *loggerWrapper) Debug(args ...interface{}) {
	// noop
}
func (logger *loggerWrapper) Debugf(format string, args ...interface{}) {
	// noop
}
func (logger *loggerWrapper) Debugw(msg string, args ...interface{}) {
	// noop
}
func (logger *loggerWrapper) Printf(format string, args ...interface{}) {
	// noop
}
func (logger *loggerWrapper) Println(args ...interface{}) {
	// noop
}

func Log() iLogger {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	if defaultLogger == nil {
		defaultLogger = &loggerWrapper{}
	}
	return defaultLogger
}

var (
	defaultLogger   iLogger
	defaultLoggerMu sync.RWMutex
)
