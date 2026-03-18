package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/stats"
)

const (
	maxConnectionIdle     = 5 * time.Minute
	maxConnectionAge      = 10 * time.Minute
	maxConnectionAgeGrace = 5 * time.Second
	keepAliveTime         = 2 * time.Minute
	keepAliveTimeout      = 20 * time.Second
)

type GRPCServer struct {
	Server   *grpc.Server
	Listener net.Listener
	Port     int
}

type Config struct {
	Port               int
	MaxConnections     int
	ConnectionTimeout  time.Duration
	Interceptors       []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor
	StatsHandler       stats.Handler
}

func NewGRPCServer(cfg Config) (*GRPCServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.Port, err)
	}

	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(uint32(cfg.MaxConnections)),
		grpc.ConnectionTimeout(cfg.ConnectionTimeout),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     maxConnectionIdle,
			MaxConnectionAge:      maxConnectionAge,
			MaxConnectionAgeGrace: maxConnectionAgeGrace,
			Time:                  keepAliveTime,
			Timeout:               keepAliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             1 * time.Minute,
			PermitWithoutStream: true,
		}),
	}

	if len(cfg.Interceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(cfg.Interceptors...))
	}

	if len(cfg.StreamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(cfg.StreamInterceptors...))
	}

	if cfg.StatsHandler != nil {
		opts = append(opts, grpc.StatsHandler(cfg.StatsHandler))
	}

	server := grpc.NewServer(opts...)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(server)

	log.Printf("Initializing gRPC server (Port: %d)...", cfg.Port)

	return &GRPCServer{
		Server:   server,
		Listener: listener,
		Port:     cfg.Port,
	}, nil
}

func (s *GRPCServer) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.Server.RegisterService(desc, impl)
}

func (s *GRPCServer) Start() error {
	log.Printf("Starting gRPC server on port %d...", s.Port)
	if err := s.Server.Serve(s.Listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		s.Server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.Server.Stop()
		return fmt.Errorf("forced shutdown due to timeout")
	}
}

func (s *GRPCServer) GetServer() *grpc.Server {
	return s.Server
}
