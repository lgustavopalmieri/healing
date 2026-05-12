package main

import (
	"context"
	"log"
	"os"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/bootstrap/auth"
	"github.com/lgustavopalmieri/healing-specialist/cmd/server/bootstrap/infra"
	"github.com/lgustavopalmieri/healing-specialist/cmd/server/bootstrap/specialist"
	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	_ "github.com/lgustavopalmieri/healing-specialist/docs"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/telemetry"
	authgrpc "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/grpc"
	authhttp "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/http"
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

	otelProvider, err := infra.InitOtel(ctx, cfg)
	if err != nil {
		return err
	}

	db, err := infra.InitDatabase(cfg)
	if err != nil {
		return err
	}

	osFactory, err := infra.InitOpenSearch(cfg)
	if err != nil {
		return err
	}

	sqsResources, err := infra.InitSQS(ctx, cfg)
	if err != nil {
		return err
	}

	snsResources, err := infra.InitSNS(ctx, cfg, sqsResources)
	if err != nil {
		return err
	}

	authDB, err := infra.InitAuthDatabase(cfg)
	if err != nil {
		return err
	}

	redisClient, err := infra.InitRedis(ctx, cfg)
	if err != nil {
		return err
	}

	signer, keyring, err := auth.InitTokenService(cfg)
	if err != nil {
		return err
	}

	authMiddleware := auth.InitMiddleware(keyring, redisClient, cfg.Auth.Issuer, cfg.Auth.Audience)

	emailSender := infra.InitEmailSender(cfg)

	serviceName := cfg.Otel.ServiceName
	grpcMetrics := telemetry.NewGRPCMetrics(serviceName)

	grpcServer, err := server.NewGRPCServer(server.Config{
		Port:              cfg.Server.GRPCPort,
		MaxConnections:    cfg.Server.MaxConnections,
		ConnectionTimeout: cfg.Server.ConnectionTimeout,
		Interceptors: []grpc.UnaryServerInterceptor{
			grpcMetrics.UnaryServerInterceptor(),
			authgrpc.UnaryInterceptor(authMiddleware.ValidateTokenUseCase, authMiddleware.Enforcer, authMiddleware.RoutePolicy),
		},
	})
	if err != nil {
		return err
	}

	httpMetrics := telemetry.NewHTTPMetrics(serviceName)

	httpServer := server.NewHTTPServer(server.HTTPConfig{
		Port: cfg.Server.HTTPPort,
	})
	httpServer.Engine.Use(httpMetrics.Middleware())
	httpServer.Engine.Use(authhttp.Middleware(authMiddleware.ValidateTokenUseCase, authMiddleware.Enforcer, authMiddleware.RoutePolicy))

	specialistLogger := telemetry.NewSlogLogger("healing-specialist")
	specialist.RegisterGRPCServices(grpcServer, specialist.ServiceDependencies{
		DB:             db,
		OSFactory:      osFactory,
		EventPublisher: snsResources.Producer,
		Logger:         specialistLogger,
	})
	specialist.RegisterHTTPServices(httpServer, specialist.ServiceDependencies{
		DB:             db,
		OSFactory:      osFactory,
		EventPublisher: snsResources.Producer,
		Logger:         specialistLogger,
	})

	authLogger := telemetry.NewSlogLogger("healing-auth")
	auth.RegisterHTTPServices(httpServer, auth.HTTPDependencies{
		AuthDB:         authDB,
		RedisClient:    redisClient,
		Signer:         signer,
		Keyring:        keyring,
		EventPublisher: snsResources.Producer,
		Logger:         authLogger,
		Config:         cfg,
	})

	specialist.InitSQSConsumers(ctx, specialist.SQSConsumerDependencies{
		DB:             db,
		OSFactory:      osFactory,
		EventPublisher: snsResources.Producer,
		EmailSender:    emailSender,
		SQS:            sqsResources,
		Config:         cfg,
	})

	auth.InitSQSConsumers(ctx, auth.SQSConsumerDependencies{
		AuthDB:         authDB,
		RedisClient:    redisClient,
		Signer:         signer,
		EventPublisher: snsResources.Producer,
		EmailSender:    emailSender,
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

	shutdownManager := infra.NewShutdownManager(cfg.Server.ShutdownTimeout)
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

	if err := shutdownManager.Shutdown(infra.ShutdownResources{
		GRPCServer:  grpcServer,
		DB:          db,
		AuthDB:      authDB,
		RedisClient: redisClient,
	}); err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	return nil
}
