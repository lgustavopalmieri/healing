package bootstrap

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/database/postgresql"
)

func InitDatabase(cfg *config.Config) (*sql.DB, error) {
	log.Printf("Connecting to PostgreSQL (%s:%d)...", cfg.Database.Host, cfg.Database.Port)

	db, err := postgresql.NewConnection(postgresql.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Database,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Running database migrations...")
	if err := postgresql.RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Printf("Database connected (db=%s, pool=%d/%d)",
		cfg.Database.Database,
		cfg.Database.MaxIdleConns,
		cfg.Database.MaxOpenConns,
	)

	return db, nil
}
