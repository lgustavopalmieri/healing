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

	log.Println("🚀 Starting Healing Specialist Service...")
	log.Println("📋 Loading configuration...")

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log.Printf("✅ Configuration loaded successfully (Environment: %s)", cfg.Observability.Environment)

	log.Println("📊 Initializing observability (OpenTelemetry + Grafana Stack)...")
	otelProvider, err := bootstrap.InitObservability(ctx, cfg)
	if err != nil {
		return err
	}
	log.Printf("✅ Observability initialized (Endpoint: %s)", cfg.Observability.OTLPEndpoint)

	log.Printf("🗄️  Connecting to PostgreSQL database (%s:%d)...", cfg.Database.Host, cfg.Database.Port)
	db, err := bootstrap.InitDatabase(cfg)
	if err != nil {
		return err
	}
	log.Printf("✅ Database connected successfully (Database: %s)", cfg.Database.Database)

	log.Printf("📨 Connecting to Kafka broker (%s)...", cfg.Kafka.BootstrapServers)
	kafkaProducer, err := bootstrap.InitKafkaProducer(cfg)
	if err != nil {
		return err
	}
	log.Println("✅ Kafka producer initialized successfully")

	log.Printf("🌐 Initializing gRPC server (Port: %d)...", cfg.Server.GRPCPort)
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

	log.Println("🔧 Registering gRPC services...")
	bootstrap.RegisterServices(grpcServer, bootstrap.ServiceDependencies{
		DB:             db,
		EventPublisher: kafkaProducer,
		Tracer:         tracer,
		Logger:         logger,
	})
	log.Println("✅ Services registered successfully")

	shutdownManager := bootstrap.NewShutdownManager(cfg.Server.ShutdownTimeout)

	go func() {
		shutdownManager.Wait()
		log.Println("🛑 Shutdown signal received, gracefully shutting down...")
		if err := shutdownManager.Shutdown(grpcServer, db, otelProvider, kafkaProducer); err != nil {
			log.Printf("❌ Shutdown error: %v", err)
			os.Exit(1)
		}
		log.Println("👋 Application stopped successfully")
		os.Exit(0)
	}()

	log.Println("✨ Application started successfully!")
	log.Printf("🎯 gRPC Server listening on port %d", cfg.Server.GRPCPort)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return grpcServer.Start()
}
