package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer serves Prometheus metrics via HTTP
type MetricsServer struct {
	server   *http.Server
	port     int
	registry *prometheus.Registry
}

// MetricsConfig holds configuration for the metrics server
type MetricsConfig struct {
	Port     int
	Registry *prometheus.Registry
}

// NewMetricsServer creates a new HTTP server for Prometheus metrics
func NewMetricsServer(cfg MetricsConfig) *MetricsServer {
	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.HandlerFor(
		cfg.Registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("📊 Initializing Metrics server (Port: %d)...", cfg.Port)

	return &MetricsServer{
		server:   server,
		port:     cfg.Port,
		registry: cfg.Registry,
	}
}

// Start begins serving metrics
func (s *MetricsServer) Start() error {
	log.Printf("📈 Starting Metrics server on port %d...", s.port)
	log.Printf("🔗 Metrics available at http://localhost:%d/metrics", s.port)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("metrics server failed: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the metrics server
func (s *MetricsServer) Shutdown(ctx context.Context) error {
	log.Println("Shutting down metrics server...")
	return s.server.Shutdown(ctx)
}
