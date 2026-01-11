package telemetry

import "context"

type Tracer interface {
	Start(ctx context.Context, spanName string) (context.Context, Span)
}

type Span interface {
	End()
	RecordError(err error)
	SetAttribute(key string, value interface{})
}
