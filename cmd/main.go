package main

import (
	"context"
	"log"
	"os"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/bootstrap"
	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	_ "github.com/lgustavopalmieri/healing-specialist/docs"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/telemetry"
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
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log.Println("Starting Healing Specialist Service...")

	ctx := context.Background()

	otelProvider, err := bootstrap.InitOtel(ctx, cfg)
	if err != nil {
		return err
	}

	db, err := bootstrap.InitDatabase(cfg)
	if err != nil {
		return err
	}

	osFactory, err := bootstrap.InitOpenSearch(cfg)
	if err != nil {
		return err
	}

	sqsResources, err := bootstrap.InitSQS(ctx, cfg)
	if err != nil {
		return err
	}

	serviceName := cfg.Otel.ServiceName
	grpcMetrics := telemetry.NewGRPCMetrics(serviceName)

	grpcServer, err := server.NewGRPCServer(server.Config{
		Port:              cfg.Server.GRPCPort,
		MaxConnections:    cfg.Server.MaxConnections,
		ConnectionTimeout: cfg.Server.ConnectionTimeout,
		Interceptors:      []grpc.UnaryServerInterceptor{grpcMetrics.UnaryServerInterceptor()},
	})
	if err != nil {
		return err
	}

	httpMetrics := telemetry.NewHTTPMetrics(serviceName)

	httpServer := server.NewHTTPServer(server.HTTPConfig{
		Port: cfg.Server.HTTPPort,
	})
	httpServer.Engine.Use(httpMetrics.Middleware())

	serviceDeps := bootstrap.ServiceDependencies{
		DB:             db,
		OSFactory:      osFactory,
		EventPublisher: sqsResources.Producer,
		Logger:         telemetry.NewSlogLogger("healing-specialist"),
	}

	bootstrap.RegisterServices(grpcServer, serviceDeps)
	bootstrap.RegisterHTTPServices(httpServer, serviceDeps)

	bootstrap.InitSQSConsumers(ctx, bootstrap.SQSConsumerDependencies{
		DB:             db,
		OSFactory:      osFactory,
		EventPublisher: sqsResources.Producer,
		SQS:            sqsResources,
		Config:         cfg,
	})

	serverErrors := make(chan error, 2)

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

	shutdownManager := bootstrap.NewShutdownManager(cfg.Server.ShutdownTimeout)
	shutdownManager.Wait()

	log.Println("Initiating graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if otelProvider != nil {
		if err := otelProvider.Shutdown(shutdownCtx); err != nil {
			log.Printf("OTel shutdown error: %v", err)
		}
	}

	if err := shutdownManager.Shutdown(grpcServer, db); err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	return nil
}
