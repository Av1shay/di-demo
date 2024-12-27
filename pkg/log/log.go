package log

import (
	"context"
	"fmt"
	"log/slog"
)

type ctxKey string

var traceIDCtxKey = ctxKey("trace_id")

func Logf(ctx context.Context, lvl slog.Level, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	traceID := TraceIDFromContext(ctx)
	var logArgs []any
	if traceID != "" {
		logArgs = append(logArgs, slog.String("trace_id", traceID))
	}
	slog.Log(ctx, lvl, msg, logArgs...)
}

func Infof(ctx context.Context, format string, args ...any) {
	Logf(ctx, slog.LevelInfo, format, args...)
}

func Errorf(ctx context.Context, format string, args ...any) {
	Logf(ctx, slog.LevelError, format, args...)
}

func Debugf(ctx context.Context, format string, args ...any) {
	Logf(ctx, slog.LevelDebug, format, args...)
}

func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDCtxKey, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDCtxKey).(string); ok {
		return traceID
	}
	return ""
}
