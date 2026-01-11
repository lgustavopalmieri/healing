package observability

import "context"

type Tracer interface {
	Start(ctx context.Context, spanName string) (context.Context, Span)
}

type Span interface {
	End()
	RecordError(err error)
	SetAttribute(key string, attribute ...Attribute)
}

type Attribute struct {
	Key   string
	Value string
}