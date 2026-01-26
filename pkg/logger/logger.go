package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	Filename     string
	MaxSize      int // megabytes
	MaxAge       int // days
	MaxBackups   int
	Compress     bool
	ClearOnStart bool
}

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
func New(config LogConfig, quietMode bool) (Logger, error) {
	if config.ClearOnStart {
		err := os.Remove(config.Filename)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to clear log file: %w", err)
		}
	}

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Clean(config.Filename),
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	})

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		w,
		getLogLevel(quietMode),
	)

	logger := zap.New(core)

	return &ZapLogger{
		SugaredLogger: logger.Sugar(),
		quietMode:     quietMode,
	}, nil
}

func getLogLevel(quietMode bool) zapcore.Level {
	if quietMode {
		return zapcore.ErrorLevel
	}
	return zapcore.InfoLevel
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
	// Note: Fatalw already calls os.Exit(1), no need for explicit exit
}
