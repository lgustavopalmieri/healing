package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func NewPrometheusMetrics() *PrometheusMetrics {
	registry := prometheus.NewRegistry()

	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	return &PrometheusMetrics{
		registry: registry,
	}
}

func (p *PrometheusMetrics) Registry() *prometheus.Registry {
	return p.registry
}
