package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the interface that defines the logging methods
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	Close() error
}

// ZapLogger is the concrete implementation of the Logger interface
type ZapLogger struct {
	*zap.SugaredLogger
	quietMode bool
}

// New creates a new Logger instance
func New(filename string, quietMode bool) (Logger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{filepath.Clean(filename)}

	if quietMode {
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &ZapLogger{
		SugaredLogger: logger.Sugar(),
		quietMode:     quietMode,
	}, nil
}

func (l *ZapLogger) Close() error {
	return l.Sync()
}

func (l *ZapLogger) Info(msg string, keysAndValues ...interface{}) {
	if !l.quietMode {
		l.SugaredLogger.Infow(msg, keysAndValues...)
	}
}

func (l *ZapLogger) Error(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Errorw(msg, keysAndValues...)
}

func (l *ZapLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Warnw(msg, keysAndValues...)
}

func (l *ZapLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Debugw(msg, keysAndValues...)
}

func (l *ZapLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Fatalw(msg, keysAndValues...)
	os.Exit(1)
}
