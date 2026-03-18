package telemetry

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type Factory struct {
	ServiceName string
}

func NewFactory(serviceName string) *Factory {
	return &Factory{
		ServiceName: serviceName,
	}
}

func (f *Factory) Tracer(scope string) observability.Tracer {
	return NewOtelTracer(scope)
}

func (f *Factory) Logger(scope string) observability.Logger {
	return NewSlogLogger(scope)
}

func (f *Factory) Metrics(scope string) observability.Metrics {
	return NewOtelMetrics(scope)
}
