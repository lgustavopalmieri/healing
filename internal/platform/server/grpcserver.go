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
)

const (
	maxConnectionIdle     = 5 * time.Minute
	maxConnectionAge      = 10 * time.Minute
	maxConnectionAgeGrace = 5 * time.Second
	keepAliveTime         = 2 * time.Minute
	keepAliveTimeout      = 20 * time.Second
)

type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	port     int
}

type Config struct {
	Port              int
	MaxConnections    int
	ConnectionTimeout time.Duration
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

	server := grpc.NewServer(opts...)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(server)

	return &GRPCServer{
		server:   server,
		listener: listener,
		port:     cfg.Port,
	}, nil
}

func (s *GRPCServer) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(desc, impl)
}

func (s *GRPCServer) Start() error {
	log.Printf("Starting gRPC server on port %d...", s.port)
	if err := s.server.Serve(s.listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return fmt.Errorf("forced shutdown due to timeout")
	}
}

func (s *GRPCServer) GetServer() *grpc.Server {
	return s.server
}
