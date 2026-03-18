package telemetry

import (
	"context"
	"log/slog"
	"os"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"go.opentelemetry.io/otel/trace"
)

type SlogLogger struct {
	Logger      *slog.Logger
	ServiceName string
}

func NewSlogLogger(serviceName string) *SlogLogger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	return &SlogLogger{
		Logger:      slog.New(handler),
		ServiceName: serviceName,
	}
}

func (l *SlogLogger) Debug(ctx context.Context, msg string, fields ...observability.Field) {
	l.log(ctx, slog.LevelDebug, msg, fields...)
}

func (l *SlogLogger) Info(ctx context.Context, msg string, fields ...observability.Field) {
	l.log(ctx, slog.LevelInfo, msg, fields...)
}

func (l *SlogLogger) Warn(ctx context.Context, msg string, fields ...observability.Field) {
	l.log(ctx, slog.LevelWarn, msg, fields...)
}

func (l *SlogLogger) Error(ctx context.Context, msg string, fields ...observability.Field) {
	l.log(ctx, slog.LevelError, msg, fields...)
}

func (l *SlogLogger) log(ctx context.Context, level slog.Level, msg string, fields ...observability.Field) {
	attrs := make([]slog.Attr, 0, len(fields)+3)
	attrs = append(attrs, slog.String("service", l.ServiceName))

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		attrs = append(attrs,
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	for _, field := range fields {
		attrs = append(attrs, slog.String(field.Key, field.Value))
	}

	l.Logger.LogAttrs(ctx, level, msg, attrs...)
}
