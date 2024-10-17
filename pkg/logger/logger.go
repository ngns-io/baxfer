package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger to provide a simpler interface
type Logger struct {
	*zap.Logger
}

// New creates a new Logger instance
func New(filename string) (*Logger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout", filename}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{logger}, nil
}

// Close flushes any buffered log entries
func (l *Logger) Close() error {
	return l.Sync()
}

// Info logs an info message
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Infow(msg, keysAndValues...)
}

// Error logs an error message
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Errorw(msg, keysAndValues...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Warnw(msg, keysAndValues...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Debugw(msg, keysAndValues...)
}

// Fatal logs a fatal message and then calls os.Exit(1)
func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Fatalw(msg, keysAndValues...)
	os.Exit(1)
}