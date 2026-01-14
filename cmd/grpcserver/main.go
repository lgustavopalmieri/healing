package main

import (
	"context"
	"log"
	"os"

	"github.com/lgustavopalmieri/healing-specialist/cmd/grpcserver/bootstrap"
	"github.com/lgustavopalmieri/healing-specialist/cmd/grpcserver/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/opentelemetry"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	otelProvider, err := bootstrap.InitObservability(ctx, cfg)
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
	})
	if err != nil {
		return err
	}

	logger := opentelemetry.NewLogger(cfg.Observability.ServiceName)
	tracer := opentelemetry.NewTracer(cfg.Observability.ServiceName)

	bootstrap.RegisterServices(grpcServer, bootstrap.ServiceDependencies{
		DB:             db,
		EventPublisher: kafkaProducer,
		Tracer:         tracer,
		Logger:         logger,
	})

	shutdownManager := bootstrap.NewShutdownManager(cfg.Server.ShutdownTimeout)

	go func() {
		shutdownManager.Wait()
		if err := shutdownManager.Shutdown(grpcServer, db, otelProvider, kafkaProducer); err != nil {
			log.Printf("Shutdown error: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	log.Println("Application started successfully")
	return grpcServer.Start()
}
