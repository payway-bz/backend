package logx

import (
	"context"
	"log/slog"
	"runtime"
)

func callerFunc(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	if f := runtime.FuncForPC(pc); f != nil {
		return f.Name()
	}
	return ""
}

// Debug logs with function name attribute automatically.
func Debug(ctx context.Context, msg string, args ...any) {
	l := slog.Default().With(slog.String("func", callerFunc(2)))
	l.DebugContext(ctx, msg, args...)
}

// Info logs with function name attribute automatically.
func Info(ctx context.Context, msg string, args ...any) {
	l := slog.Default().With(slog.String("func", callerFunc(2)))
	l.InfoContext(ctx, msg, args...)
}

// Warn logs with function name attribute automatically.
func Warn(ctx context.Context, msg string, args ...any) {
	l := slog.Default().With(slog.String("func", callerFunc(2)))
	l.WarnContext(ctx, msg, args...)
}

// Error logs with function name attribute automatically.
func Error(ctx context.Context, msg string, args ...any) {
	l := slog.Default().With(slog.String("func", callerFunc(2)))
	l.ErrorContext(ctx, msg, args...)
}
