package telemetry

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type OtelTracer struct {
	Tracer trace.Tracer
}

func NewOtelTracer(serviceName string) *OtelTracer {
	return &OtelTracer{
		Tracer: otel.Tracer(serviceName),
	}
}

func (t *OtelTracer) Start(ctx context.Context, spanName string) (context.Context, observability.Span) {
	ctx, span := t.Tracer.Start(ctx, spanName)
	return ctx, &OtelSpan{Span: span}
}

type OtelSpan struct {
	Span trace.Span
}

func (s *OtelSpan) End() {
	s.Span.End()
}

func (s *OtelSpan) RecordError(err error) {
	s.Span.RecordError(err)
	s.Span.SetStatus(codes.Error, err.Error())
}

func (s *OtelSpan) SetAttribute(key string, attributes ...observability.Attribute) {
	for _, attr := range attributes {
		s.Span.SetAttributes(attribute.String(attr.Key, attr.Value))
	}
}
