package main

import (
	"context"
	"log"
	"os"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/bootstrap"
	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	_ "github.com/lgustavopalmieri/healing-specialist/docs"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// @title           Healing Specialist API
// @version         1.0
// @description     API for managing healthcare specialists on the Healing platform. Supports registration, search with cursor-based pagination, and profile updates.

// @BasePath  /

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	log.Println("Starting Healing Specialist Service...")

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	obs, err := bootstrap.InitObservability(ctx, cfg)
	if err != nil {
		return err
	}

	db, err := bootstrap.InitDatabase(cfg)
	if err != nil {
		return err
	}

	esFactory, err := bootstrap.InitElasticsearch(cfg)
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
		Interceptors: []grpc.UnaryServerInterceptor{
			obs.GRPCMetrics.UnaryServerInterceptor(),
		},
		StreamInterceptors: []grpc.StreamServerInterceptor{},
		StatsHandler:       otelgrpc.NewServerHandler(),
	})
	if err != nil {
		return err
	}

	httpServer := server.NewHTTPServer(server.HTTPConfig{
		Port: cfg.Server.HTTPPort,
	})

	metricsServer := server.NewMetricsServer(server.MetricsConfig{
		Port:           cfg.Server.MetricsPort,
		MetricsHandler: obs.Provider.MetricsHandler(),
	})

	serviceDeps := bootstrap.ServiceDependencies{
		DB:             db,
		ESFactory:      esFactory,
		EventPublisher: kafkaProducer,
		Factory:        obs.Factory,
	}

	bootstrap.RegisterServices(grpcServer, serviceDeps)
	bootstrap.RegisterHTTPServices(httpServer, serviceDeps)

	if err := bootstrap.InitKafkaConsumers(ctx, bootstrap.ConsumerDependencies{
		DB:             db,
		ESFactory:      esFactory,
		Factory:        obs.Factory,
		EventPublisher: kafkaProducer,
		Config:         cfg,
	}); err != nil {
		return err
	}

	serverErrors := make(chan error, 3)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("gRPC server panic recovered: %v", r)
			}
		}()
		log.Println("gRPC server starting...")
		if err := grpcServer.Start(); err != nil {
			log.Printf("gRPC server error: %v", err)
			serverErrors <- err
		}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("HTTP server panic recovered: %v", r)
			}
		}()
		log.Println("HTTP server starting...")
		if err := httpServer.Start(); err != nil {
			log.Printf("HTTP server error: %v", err)
			serverErrors <- err
		}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Metrics server panic recovered: %v", r)
			}
		}()
		if err := metricsServer.Start(); err != nil {
			log.Printf("Metrics server error: %v", err)
			serverErrors <- err
		}
	}()

	shutdownManager := bootstrap.NewShutdownManager(cfg.Server.ShutdownTimeout)
	shutdownManager.Wait()

	log.Println("Initiating graceful shutdown...")
	obs.GRPCMetrics.SetUnhealthy()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Metrics server shutdown error: %v", err)
	}

	if err := shutdownManager.Shutdown(grpcServer, db, obs.Provider, kafkaProducer); err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	return nil
}
