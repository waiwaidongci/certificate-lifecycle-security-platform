package logger

import (
	"context"
	"log/slog"
	"os"
)

type Logger struct {
	inner *slog.Logger
}

func New(level string) *Logger {
	var parsed slog.Level
	switch level {
	case "debug":
		parsed = slog.LevelDebug
	case "warn":
		parsed = slog.LevelWarn
	case "error":
		parsed = slog.LevelError
	default:
		parsed = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: parsed}
	return &Logger{inner: slog.New(slog.NewJSONHandler(os.Stdout, opts))}
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{inner: l.inner.With(args...)}
}

func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.inner.DebugContext(ctx, msg, args...)
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.inner.InfoContext(ctx, msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.inner.WarnContext(ctx, msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.inner.ErrorContext(ctx, msg, args...)
}
