package opentelemetry

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type OtelTracer struct {
	tracer trace.Tracer
}

type OtelSpan struct {
	span trace.Span
}

func NewTracer(name string) observability.Tracer {
	return &OtelTracer{
		tracer: otel.Tracer(name),
	}
}

func (t *OtelTracer) Start(ctx context.Context, spanName string) (context.Context, observability.Span) {
	ctx, span := t.tracer.Start(ctx, spanName)
	return ctx, &OtelSpan{span: span}
}

func (s *OtelSpan) End() {
	s.span.End()
}

func (s *OtelSpan) RecordError(err error) {
	s.span.RecordError(err)
}

func (s *OtelSpan) SetAttribute(key string, attrs ...observability.Attribute) {
	otelAttrs := make([]attribute.KeyValue, len(attrs))
	for i, attr := range attrs {
		otelAttrs[i] = attribute.String(attr.Key, attr.Value)
	}
	s.span.SetAttributes(otelAttrs...)
}
