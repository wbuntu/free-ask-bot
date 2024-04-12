package log

import (
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

func NewLogrusAdaptor(logLevel int, format string) (*LogrusLoggerAdapter, error) {
	logger := &LogrusLoggerAdapter{logrus.New()}
	var customFormatter logrus.Formatter
	// 只打印文件和行数
	callerPrettyfier := func(f *runtime.Frame) (string, string) {
		return "", fmt.Sprintf("%s:%d", f.File, f.Line)
	}
	switch format {
	case "json":
		// 设置JSONFormatter
		formatter := new(logrus.JSONFormatter)
		formatter.TimestampFormat = time.StampMilli
		formatter.CallerPrettyfier = callerPrettyfier
		customFormatter = formatter
	default:
		// 其他情况默认使用TextFormatter
		format = "text"
		formatter := new(logrus.TextFormatter)
		formatter.TimestampFormat = time.StampMilli
		formatter.FullTimestamp = true
		formatter.CallerPrettyfier = callerPrettyfier
		customFormatter = formatter
	}
	logger.SetFormatter(customFormatter)
	logger.SetReportCaller(true)
	logger.SetLevel(logrus.Level(logLevel))
	logger.WithFields(Fields{"logLevel": logger.Level, "formatter": format}).Info("setup logrus success")
	return logger, nil
}

// logger = logrus.Logger
var _ Logger = (*LogrusLoggerAdapter)(nil)

type LogrusLoggerAdapter struct {
	*logrus.Logger
}

func (l *LogrusLoggerAdapter) WithField(key string, value interface{}) Logger {
	return &LogrusEntryAdapter{l.Logger.WithField(key, value)}
}

func (l *LogrusLoggerAdapter) WithFields(fields Fields) Logger {
	return &LogrusEntryAdapter{l.Logger.WithFields(logrus.Fields(fields))}
}

func (l *LogrusLoggerAdapter) WithError(err error) Logger {
	return &LogrusEntryAdapter{l.Logger.WithError(err)}
}

// Entry = logrus.Logger + logrus.Fields
var _ Logger = (*LogrusEntryAdapter)(nil)

type LogrusEntryAdapter struct {
	*logrus.Entry
}

func (e *LogrusEntryAdapter) WithField(key string, value interface{}) Logger {
	return &LogrusEntryAdapter{e.Entry.WithField(key, value)}
}

func (e *LogrusEntryAdapter) WithFields(fields Fields) Logger {
	return &LogrusEntryAdapter{e.Entry.WithFields(logrus.Fields(fields))}
}

func (e *LogrusEntryAdapter) WithError(err error) Logger {
	return &LogrusEntryAdapter{e.Entry.WithError(err)}
}
