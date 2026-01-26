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

	"github.com/lgustavopalmieri/healing-specialist/internal/platform/kafka"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/opentelemetry"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
)

type ShutdownManager struct {
	timeout time.Duration
	signals chan os.Signal
	done    chan struct{}
}

func NewShutdownManager(timeout time.Duration) *ShutdownManager {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	return &ShutdownManager{
		timeout: timeout,
		signals: signals,
		done:    make(chan struct{}),
	}
}

func (sm *ShutdownManager) Wait() {
	<-sm.signals
	log.Println("Shutdown signal received, starting graceful shutdown...")
}

func (sm *ShutdownManager) Shutdown(
	grpcServer *server.GRPCServer,
	db *sql.DB,
	otelProvider *opentelemetry.GrafanaProvider,
	kafkaProducer *kafka.KafkaProducer,
) error {
	log.Println("🛑 Shutdown signal received, gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), sm.timeout)
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Shutting down gRPC server...")
		if err := grpcServer.Shutdown(ctx); err != nil {
			errChan <- fmt.Errorf("grpc server shutdown: %w", err)
			return
		}
		log.Println("gRPC server stopped")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Closing database connections...")
		if err := db.Close(); err != nil {
			errChan <- fmt.Errorf("database close: %w", err)
			return
		}
		log.Println("Database connections closed")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Shutting down observability provider...")
		if err := otelProvider.Shutdown(ctx); err != nil {
			errChan <- fmt.Errorf("observability shutdown: %w", err)
			return
		}
		log.Println("Observability provider stopped")
	}()

	if kafkaProducer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("Closing Kafka producer...")
			kafkaProducer.Close()
			log.Println("Kafka producer closed")
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
		log.Println("👋 Graceful shutdown completed successfully")
		return nil

	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout exceeded")
	}
}
