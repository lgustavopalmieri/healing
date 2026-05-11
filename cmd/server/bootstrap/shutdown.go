package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
)

type ShutdownManager struct {
	Timeout time.Duration
	Signals chan os.Signal
	Done    chan struct{}
}

func NewShutdownManager(timeout time.Duration) *ShutdownManager {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	return &ShutdownManager{
		Timeout: timeout,
		Signals: signals,
		Done:    make(chan struct{}),
	}
}

func (sm *ShutdownManager) Wait() {
	<-sm.Signals
	log.Println("Shutdown signal received, starting graceful shutdown...")
}

type ShutdownResources struct {
	GRPCServer  *server.GRPCServer
	DB          *sql.DB
	AuthDB      *sql.DB
	RedisClient *redis.Client
}

func (sm *ShutdownManager) Shutdown(resources ShutdownResources) error {
	log.Println("Shutdown signal received, gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), sm.Timeout)
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	if resources.GRPCServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("Shutting down gRPC server...")
			if err := resources.GRPCServer.Shutdown(ctx); err != nil {
				errChan <- fmt.Errorf("grpc server shutdown: %w", err)
				return
			}
			log.Println("gRPC server stopped")
		}()
	}

	if resources.DB != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("Closing specialist database connections...")
			if err := resources.DB.Close(); err != nil {
				errChan <- fmt.Errorf("specialist database close: %w", err)
				return
			}
			log.Println("Specialist database connections closed")
		}()
	}

	if resources.AuthDB != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("Closing auth database connections...")
			if err := resources.AuthDB.Close(); err != nil {
				errChan <- fmt.Errorf("auth database close: %w", err)
				return
			}
			log.Println("Auth database connections closed")
		}()
	}

	if resources.RedisClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("Closing redis client...")
			if err := resources.RedisClient.Close(); err != nil {
				errChan <- fmt.Errorf("redis close: %w", err)
				return
			}
			log.Println("Redis client closed")
		}()
	}

	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		close(errChan)
		var errs []error
		for err := range errChan {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return fmt.Errorf("shutdown errors: %v", errs)
		}
		log.Println("Graceful shutdown completed successfully")
		return nil

	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout exceeded")
	}
}
