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

	esClient, err := bootstrap.InitElasticsearch(cfg)
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

	httpServer := server.NewHTTPServer(server.HTTPConfig{
		Port: cfg.Server.HTTPPort,
	})

	metricsServer := server.NewMetricsServer(server.MetricsConfig{
		Port:     9090,
		Registry: observability.Metrics.(*telemetry.PrometheusMetrics).Registry(),
	})

	go func() {
		if err := metricsServer.Start(); err != nil {
			log.Printf("❌ Metrics server error: %v", err)
		}
	}()

	serviceDeps := bootstrap.ServiceDependencies{
		DB:                 db,
		ESClient:           esClient,
		EventPublisher:     kafkaProducer,
		Tracer:             observability.Tracer,
		Logger:             observability.Logger,
		ESIndexSpecialists: cfg.Elasticsearch.IndexSpecialists,
	}

	bootstrap.RegisterServices(grpcServer, serviceDeps)
	bootstrap.RegisterHTTPServices(httpServer, serviceDeps)

	if err := bootstrap.InitKafkaConsumers(ctx, bootstrap.ConsumerDependencies{
		DB:                 db,
		ESClient:           esClient,
		ESIndexSpecialists: cfg.Elasticsearch.IndexSpecialists,
		Tracer:             observability.Tracer,
		Logger:             observability.Logger,
		EventPublisher:     kafkaProducer,
		Config:             cfg,
	}); err != nil {
		return err
	}

	shutdownManager := bootstrap.NewShutdownManager(cfg.Server.ShutdownTimeout)

	go func() {
		shutdownManager.Wait()
		observability.GRPCMetrics.SetUnhealthy()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("⚠️ HTTP server shutdown error: %v", err)
		}

		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("⚠️ Metrics server shutdown error: %v", err)
		}

		if err := shutdownManager.Shutdown(grpcServer, db, observability.TracerProvider, kafkaProducer); err != nil {
			log.Printf("❌ Shutdown error: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	go func() {
		if err := httpServer.Start(); err != nil {
			log.Printf("❌ HTTP server error: %v", err)
		}
	}()

	return grpcServer.Start()
}
