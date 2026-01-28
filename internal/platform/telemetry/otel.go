package telemetry

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// OtelTracer wraps an OpenTelemetry tracer to implement the observability.Tracer interface
type OtelTracer struct {
	tracer trace.Tracer
}

// NewOtelTracer creates a new OtelTracer with the given OpenTelemetry tracer
func NewOtelTracer(tracer trace.Tracer) *OtelTracer {
	return &OtelTracer{tracer: tracer}
}

// Start creates a new span with the given name and returns the updated context and span
func (t *OtelTracer) Start(ctx context.Context, spanName string) (context.Context, observability.Span) {
	ctx, span := t.tracer.Start(ctx, spanName)
	return ctx, &OtelSpan{span: span}
}

// OtelSpan wraps an OpenTelemetry span to implement the observability.Span interface
type OtelSpan struct {
	span trace.Span
}

// End finalizes the span
func (s *OtelSpan) End() {
	s.span.End()
}

// RecordError records an error on the span and sets the span status to Error
func (s *OtelSpan) RecordError(err error) {
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

// SetAttribute adds one or more attributes to the span
func (s *OtelSpan) SetAttribute(key string, attributes ...observability.Attribute) {
	for _, attr := range attributes {
		s.span.SetAttributes(attribute.String(attr.Key, attr.Value))
	}
}
