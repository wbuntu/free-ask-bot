package log

import (
	"context"

	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/config"
)

// Setup configure log setting
func Setup(c *config.Config) error {
	logger, err := NewLogrusAdaptor(c.General.LogLevel, c.General.LogFormat)
	if err != nil {
		return err
	}
	sharedLogger = logger
	return nil
}

type loggerKey int

var userKey loggerKey

// G is an alias for GetLogger.
var G = GetLogger

// GetLogger retrieves the current logger from the context. If no logger is
// available, the default logger is returned.
func GetLogger(ctx context.Context) Logger {
	logger := ctx.Value(userKey)
	if logger == nil {
		return sharedLogger.WithFields(Fields{})
	}
	return logger.(Logger)
}

// S is an alias for SetLogger.
var S = SetLogger

// SetLogger add logger to context
func SetLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, userKey, logger)
}

// Fields allows setting multiple fields on a logger at one time.
type Fields map[string]interface{}

// Logger interface for wrapping external logger
type Logger interface {
	// Standard function
	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	// Structure function
	WithField(string, interface{}) Logger
	WithFields(Fields) Logger
}

// sharedLogger can be used for package global logging after Setup is called
var sharedLogger Logger

func Trace(args ...interface{}) {
	sharedLogger.Debug(args...)
}
func Tracef(format string, args ...interface{}) {
	sharedLogger.Debugf(format, args...)
}
func Debug(args ...interface{}) {
	sharedLogger.Debug(args...)
}
func Debugf(format string, args ...interface{}) {
	sharedLogger.Debugf(format, args...)
}
func Info(args ...interface{}) {
	sharedLogger.Info(args...)
}
func Infof(format string, args ...interface{}) {
	sharedLogger.Infof(format, args...)
}
func Warn(args ...interface{}) {
	sharedLogger.Warn(args...)
}
func Warnf(format string, args ...interface{}) {
	sharedLogger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	sharedLogger.Error(args...)
}
func Errorf(format string, args ...interface{}) {
	sharedLogger.Errorf(format, args...)
}
func Fatal(args ...interface{}) {
	sharedLogger.Fatal(args...)
}
func Fatalf(format string, args ...interface{}) {
	sharedLogger.Fatalf(format, args...)
}
func WithField(key string, value interface{}) Logger {
	return sharedLogger.WithField(key, value)
}
func WithFields(fields Fields) Logger {
	return sharedLogger.WithFields(fields)
}
