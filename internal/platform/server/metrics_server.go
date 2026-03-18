package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type MetricsServer struct {
	Server *http.Server
	Port   int
}

type MetricsConfig struct {
	Port           int
	MetricsHandler http.Handler
}

func NewMetricsServer(cfg MetricsConfig) *MetricsServer {
	mux := http.NewServeMux()

	mux.Handle("/metrics", cfg.MetricsHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Initializing Metrics server (Port: %d)...", cfg.Port)

	return &MetricsServer{
		Server: srv,
		Port:   cfg.Port,
	}
}

func (s *MetricsServer) Start() error {
	log.Printf("Starting Metrics server on port %d...", s.Port)
	if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("metrics server failed: %w", err)
	}
	return nil
}

func (s *MetricsServer) Shutdown(ctx context.Context) error {
	log.Println("Shutting down metrics server...")
	return s.Server.Shutdown(ctx)
}
