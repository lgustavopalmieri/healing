package opentelemetry

import (
	"context"
	"sync"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type OtelMetrics struct {
	meter    metric.Meter
	counters map[string]metric.Float64Counter
	histos   map[string]metric.Float64Histogram
	gauges   map[string]metric.Float64Gauge
	mu       sync.RWMutex
}

type OtelCounter struct {
	counter metric.Float64Counter
}

type OtelHistogram struct {
	histogram metric.Float64Histogram
}

type OtelGauge struct {
	gauge metric.Float64Gauge
}

func NewMetrics(name string) observability.Metrics {
	return &OtelMetrics{
		meter:    otel.Meter(name),
		counters: make(map[string]metric.Float64Counter),
		histos:   make(map[string]metric.Float64Histogram),
		gauges:   make(map[string]metric.Float64Gauge),
	}
}

func (m *OtelMetrics) Counter(name string) observability.Counter {
	m.mu.RLock()
	counter, exists := m.counters[name]
	m.mu.RUnlock()

	if exists {
		return &OtelCounter{counter: counter}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if counter, exists := m.counters[name]; exists {
		return &OtelCounter{counter: counter}
	}

	counter, err := m.meter.Float64Counter(name)
	if err != nil {
		// Return a no-op counter to prevent application crashes
		return &OtelCounter{counter: nil}
	}

	m.counters[name] = counter
	return &OtelCounter{counter: counter}
}

func (m *OtelMetrics) Histogram(name string) observability.Histogram {
	m.mu.RLock()
	histo, exists := m.histos[name]
	m.mu.RUnlock()

	if exists {
		return &OtelHistogram{histogram: histo}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if histo, exists := m.histos[name]; exists {
		return &OtelHistogram{histogram: histo}
	}

	histo, err := m.meter.Float64Histogram(name)
	if err != nil {
		// Return a no-op histogram to prevent application crashes
		return &OtelHistogram{histogram: nil}
	}

	m.histos[name] = histo
	return &OtelHistogram{histogram: histo}
}

func (m *OtelMetrics) Gauge(name string) observability.Gauge {
	m.mu.RLock()
	gauge, exists := m.gauges[name]
	m.mu.RUnlock()

	if exists {
		return &OtelGauge{gauge: gauge}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if gauge, exists := m.gauges[name]; exists {
		return &OtelGauge{gauge: gauge}
	}

	gauge, err := m.meter.Float64Gauge(name)
	if err != nil {
		// Return a no-op gauge to prevent application crashes
		return &OtelGauge{gauge: nil}
	}

	m.gauges[name] = gauge
	return &OtelGauge{gauge: gauge}
}

func (c *OtelCounter) Add(ctx context.Context, value float64, labels ...observability.Label) {
	if c.counter == nil {
		return
	}
	attrs := make([]attribute.KeyValue, len(labels))
	for i, label := range labels {
		attrs[i] = attribute.String(label.Key, label.Value)
	}
	c.counter.Add(ctx, value, metric.WithAttributes(attrs...))
}

func (h *OtelHistogram) Record(ctx context.Context, value float64, labels ...observability.Label) {
	if h.histogram == nil {
		return
	}
	attrs := make([]attribute.KeyValue, len(labels))
	for i, label := range labels {
		attrs[i] = attribute.String(label.Key, label.Value)
	}
	h.histogram.Record(ctx, value, metric.WithAttributes(attrs...))
}

func (g *OtelGauge) Set(ctx context.Context, value float64, labels ...observability.Label) {
	if g.gauge == nil {
		return
	}
	attrs := make([]attribute.KeyValue, len(labels))
	for i, label := range labels {
		attrs[i] = attribute.String(label.Key, label.Value)
	}
	g.gauge.Record(ctx, value, metric.WithAttributes(attrs...))
}
