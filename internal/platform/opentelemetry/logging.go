package opentelemetry

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

type OtelLogger struct {
	logger log.Logger
}

func NewLogger(name string) observability.Logger {
	return &OtelLogger{
		logger: global.GetLoggerProvider().Logger(name),
	}
}

func (l *OtelLogger) Debug(ctx context.Context, msg string, fields ...observability.Field) {
	l.log(ctx, log.SeverityDebug, msg, fields)
}

func (l *OtelLogger) Info(ctx context.Context, msg string, fields ...observability.Field) {
	l.log(ctx, log.SeverityInfo, msg, fields)
}

func (l *OtelLogger) Warn(ctx context.Context, msg string, fields ...observability.Field) {
	l.log(ctx, log.SeverityWarn, msg, fields)
}

func (l *OtelLogger) Error(ctx context.Context, msg string, fields ...observability.Field) {
	l.log(ctx, log.SeverityError, msg, fields)
}

func (l *OtelLogger) log(ctx context.Context, severity log.Severity, msg string, fields []observability.Field) {
	attrs := make([]log.KeyValue, len(fields))
	for i, field := range fields {
		attrs[i] = log.String(field.Key, field.Value)
	}

	record := log.Record{}
	record.SetSeverity(severity)
	record.SetBody(log.StringValue(msg))
	record.AddAttributes(attrs...)

	l.logger.Emit(ctx, record)
}
