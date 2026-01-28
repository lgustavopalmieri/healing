package telemetry

import (
	"context"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetrics struct {
	registry   *prometheus.Registry
	counters   sync.Map
	histograms sync.Map
	gauges     sync.Map
}

func (p *PrometheusMetrics) Counter(name string) observability.Counter {
	if counter, ok := p.counters.Load(name); ok {
		return counter.(*prometheusCounter)
	}

	counterVec := promauto.With(p.registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
		},
		[]string{"endpoint", "status"},
	)

	counter := &prometheusCounter{vec: counterVec}
	p.counters.Store(name, counter)
	return counter
}

func (p *PrometheusMetrics) Histogram(name string) observability.Histogram {
	if histogram, ok := p.histograms.Load(name); ok {
		return histogram.(*prometheusHistogram)
	}

	buckets := []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}

	histogramVec := promauto.With(p.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Buckets: buckets,
		},
		[]string{},
	)

	histogram := &prometheusHistogram{vec: histogramVec}
	p.histograms.Store(name, histogram)
	return histogram
}

func (p *PrometheusMetrics) Gauge(name string) observability.Gauge {
	if gauge, ok := p.gauges.Load(name); ok {
		return gauge.(*prometheusGauge)
	}

	gaugeVec := promauto.With(p.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
		},
		[]string{},
	)

	gauge := &prometheusGauge{vec: gaugeVec}
	p.gauges.Store(name, gauge)
	return gauge
}

type prometheusCounter struct {
	vec *prometheus.CounterVec
}

func (c *prometheusCounter) Add(ctx context.Context, value float64, labels ...observability.Label) {
	labelNames, labelValues := extractLabels(labels)
	c.vec.WithLabelValues(labelValues...).Add(value)
	_ = labelNames
}

type prometheusHistogram struct {
	vec *prometheus.HistogramVec
}

func (h *prometheusHistogram) Record(ctx context.Context, value float64, labels ...observability.Label) {
	labelNames, labelValues := extractLabels(labels)
	h.vec.WithLabelValues(labelValues...).Observe(value)
	_ = labelNames
}

type prometheusGauge struct {
	vec *prometheus.GaugeVec
}

func (g *prometheusGauge) Set(ctx context.Context, value float64, labels ...observability.Label) {
	labelNames, labelValues := extractLabels(labels)
	g.vec.WithLabelValues(labelValues...).Set(value)
	_ = labelNames
}

func extractLabels(labels []observability.Label) ([]string, []string) {
	names := make([]string, len(labels))
	values := make([]string, len(labels))
	for i, label := range labels {
		names[i] = label.Key
		values[i] = label.Value
	}
	return names, values
}
