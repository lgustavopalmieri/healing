package telemetry

import (
	"context"
	"sync"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

type OtelMetrics struct {
	Meter      otelmetric.Meter
	Counters   sync.Map
	Histograms sync.Map
	Gauges     sync.Map
}

func NewOtelMetrics(serviceName string) *OtelMetrics {
	return &OtelMetrics{
		Meter: otel.Meter(serviceName),
	}
}

func (m *OtelMetrics) Counter(name string) observability.Counter {
	if c, ok := m.Counters.Load(name); ok {
		return c.(*otelCounter)
	}

	counter, err := m.Meter.Float64Counter(name)
	if err != nil {
		return &noopCounter{}
	}

	c := &otelCounter{Counter: counter}
	m.Counters.Store(name, c)
	return c
}

func (m *OtelMetrics) Histogram(name string) observability.Histogram {
	if h, ok := m.Histograms.Load(name); ok {
		return h.(*otelHistogram)
	}

	histogram, err := m.Meter.Float64Histogram(name)
	if err != nil {
		return &noopHistogram{}
	}

	h := &otelHistogram{Histogram: histogram}
	m.Histograms.Store(name, h)
	return h
}

func (m *OtelMetrics) Gauge(name string) observability.Gauge {
	if g, ok := m.Gauges.Load(name); ok {
		return g.(*otelGauge)
	}

	gauge, err := m.Meter.Float64Gauge(name)
	if err != nil {
		return &noopGauge{}
	}

	g := &otelGauge{Gauge: gauge}
	m.Gauges.Store(name, g)
	return g
}

type otelCounter struct {
	Counter otelmetric.Float64Counter
}

func (c *otelCounter) Add(ctx context.Context, value float64, labels ...observability.Label) {
	c.Counter.Add(ctx, value, otelmetric.WithAttributes(toAttributes(labels)...))
}

type otelHistogram struct {
	Histogram otelmetric.Float64Histogram
}

func (h *otelHistogram) Record(ctx context.Context, value float64, labels ...observability.Label) {
	h.Histogram.Record(ctx, value, otelmetric.WithAttributes(toAttributes(labels)...))
}

type otelGauge struct {
	Gauge otelmetric.Float64Gauge
}

func (g *otelGauge) Set(ctx context.Context, value float64, labels ...observability.Label) {
	g.Gauge.Record(ctx, value, otelmetric.WithAttributes(toAttributes(labels)...))
}

func toAttributes(labels []observability.Label) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, len(labels))
	for i, l := range labels {
		attrs[i] = attribute.String(l.Key, l.Value)
	}
	return attrs
}

type noopCounter struct{}

func (n *noopCounter) Add(_ context.Context, _ float64, _ ...observability.Label) {}

type noopHistogram struct{}

func (n *noopHistogram) Record(_ context.Context, _ float64, _ ...observability.Label) {}

type noopGauge struct{}

func (n *noopGauge) Set(_ context.Context, _ float64, _ ...observability.Label) {}
