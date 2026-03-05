package bootstrap

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/database/postgresql"
)

const (
	defaultSSLMode         = "require"
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 5 * time.Minute
	defaultConnMaxIdleTime = 10 * time.Minute
)

func InitDatabase(cfg *config.Config) (*sql.DB, error) {
	log.Printf("🗄️  Connecting to PostgreSQL database (%s:%d)...", cfg.Database.Host, cfg.Database.Port)

	sslMode := cfg.Database.SSLMode
	if sslMode == "" {
		if cfg.Observability.Environment == "development" || cfg.Observability.Environment == "test" {
			sslMode = "disable"
		} else {
			sslMode = "require"
		}
	}

	db, err := postgresql.NewConnection(postgresql.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Database,
		SSLMode:         sslMode,
		MaxOpenConns:    defaultMaxOpenConns,
		MaxIdleConns:    defaultMaxIdleConns,
		ConnMaxLifetime: defaultConnMaxLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	db.SetConnMaxIdleTime(defaultConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("🔄 Running database migrations...")
	if err := postgresql.RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Println("✅ Migrations completed successfully")

	log.Printf("✅ Database connected successfully (Database: %s)", cfg.Database.Database)

	return db, nil
}
