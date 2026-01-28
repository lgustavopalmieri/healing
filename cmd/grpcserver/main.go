package main

import (
	"context"
	"log"
	"os"

	"github.com/lgustavopalmieri/healing-specialist/cmd/grpcserver/bootstrap"
	"github.com/lgustavopalmieri/healing-specialist/cmd/grpcserver/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/telemetry"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	log.Println("🚀 Starting Healing Specialist Service...")
	log.Println("📋 Loading configuration...")

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	observability, err := bootstrap.InitObservability(ctx, cfg)
	if err != nil {
		return err
	}

	db, err := bootstrap.InitDatabase(cfg)
	if err != nil {
		return err
	}

	kafkaProducer, err := bootstrap.InitKafkaProducer(cfg)
	if err != nil {
		return err
	}

	grpcServer, err := server.NewGRPCServer(server.Config{
		Port:              cfg.Server.GRPCPort,
		MaxConnections:    cfg.Server.MaxConnections,
		ConnectionTimeout: cfg.Server.ConnectionTimeout,
		Interceptors:      []grpc.UnaryServerInterceptor{observability.GRPCMetrics.UnaryServerInterceptor()},
	})
	if err != nil {
		return err
	}

	// Initialize metrics server
	metricsServer := server.NewMetricsServer(server.MetricsConfig{
		Port:     9090, // Prometheus default port
		Registry: observability.Metrics.(*telemetry.PrometheusMetrics).Registry(),
	})

	// Start metrics server in background
	go func() {
		if err := metricsServer.Start(); err != nil {
			log.Printf("❌ Metrics server error: %v", err)
		}
	}()

	bootstrap.RegisterServices(grpcServer, bootstrap.ServiceDependencies{
		DB:             db,
		EventPublisher: kafkaProducer,
		Tracer:         observability.Tracer,
		Logger:         observability.Logger,
	})

	shutdownManager := bootstrap.NewShutdownManager(cfg.Server.ShutdownTimeout)

	go func() {
		shutdownManager.Wait()
		// Mark application as unhealthy during shutdown
		observability.GRPCMetrics.SetUnhealthy()

		// Shutdown metrics server first
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("⚠️ Metrics server shutdown error: %v", err)
		}

		if err := shutdownManager.Shutdown(grpcServer, db, observability.TracerProvider, kafkaProducer); err != nil {
			log.Printf("❌ Shutdown error: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	return grpcServer.Start()
}
