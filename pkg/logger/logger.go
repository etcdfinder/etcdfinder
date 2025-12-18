package logger

import (
	"context"

	"github.com/etcdfinder/etcdfinder/internal/config"
	"github.com/etcdfinder/etcdfinder/internal/lib"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.SugaredLogger to provide logging functionality
type Logger struct {
	*zap.SugaredLogger
}

// Global logger for convenience
var L *Logger

// NewLogger creates and returns a new Logger instance
func NewLogger(cfg *config.Config) error {
	config := zap.NewProductionConfig()

	if cfg.Log.Level == lib.LogLevelDebug {
		config = zap.NewDevelopmentConfig()
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := config.Build()
	if err != nil {
		return err
	}

	L = &Logger{
		SugaredLogger: zapLogger.Sugar(),
	}

	return nil
}

// Helper methods to make logging more convenient
func Debugf(template string, args ...any) {
	L.Debugf(template, args...)
}

func Infof(template string, args ...any) {
	L.Infof(template, args...)
}

func Warnf(template string, args ...any) {
	L.Warnf(template, args...)
}

func Errorf(template string, args ...any) {
	L.Errorf(template, args...)
}

func Fatalf(template string, args ...any) {
	L.Fatalf(template, args...)
}

func WithContext(ctx context.Context) *Logger {
	return &Logger{
		SugaredLogger: L.With(
			"request_id", lib.GetRequestID(ctx),
		),
	}
}
