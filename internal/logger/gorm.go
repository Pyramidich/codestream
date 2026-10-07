package logger

import (
	"context"
	"log/slog"
	"time"

	gormLogger "gorm.io/gorm/logger"
)

type gormSlogLogger struct {
	logger *slog.Logger
}

// NewGORMLogger creates a GORM logger that uses slog.
func NewGORMLogger(logger *slog.Logger) gormLogger.Interface {
	return &gormSlogLogger{logger: logger}
}

func (l *gormSlogLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return l
}

func (l *gormSlogLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.logger.Info(msg, slog.Any("args", args))
}

func (l *gormSlogLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.logger.Warn(msg, slog.Any("args", args))
}

func (l *gormSlogLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.logger.Error(msg, slog.Any("args", args))
}

func (l *gormSlogLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	 elapsed := time.Since(begin)
	sql, rows := fc()
	l.logger.Debug("sql trace", slog.Duration("elapsed", elapsed), slog.String("sql", sql), slog.Int64("rows", rows))
}
