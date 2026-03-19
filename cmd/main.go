package main

import (
	"context"
	"log"
	"os"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/bootstrap"
	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	_ "github.com/lgustavopalmieri/healing-specialist/docs"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
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
		Port:               cfg.Server.GRPCPort,
		MaxConnections:     cfg.Server.MaxConnections,
		ConnectionTimeout:  cfg.Server.ConnectionTimeout,
		Interceptors:       nil,
		StreamInterceptors: nil,
	})
	if err != nil {
		return err
	}

	httpServer := server.NewHTTPServer(server.HTTPConfig{
		Port: cfg.Server.HTTPPort,
	})

	serviceDeps := bootstrap.ServiceDependencies{
		DB:             db,
		ESFactory:      esFactory,
		EventPublisher: kafkaProducer,
	}

	bootstrap.RegisterServices(grpcServer, serviceDeps)
	bootstrap.RegisterHTTPServices(httpServer, serviceDeps)

	if err := bootstrap.InitKafkaConsumers(context.Background(), bootstrap.ConsumerDependencies{
		DB:             db,
		ESFactory:      esFactory,
		EventPublisher: kafkaProducer,
		Config:         cfg,
	}); err != nil {
		return err
	}

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

	if err := shutdownManager.Shutdown(grpcServer, db, kafkaProducer); err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	return nil
}
