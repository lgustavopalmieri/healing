package infra

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/database/postgresql/auth"
)

func InitAuthDatabase(cfg *config.Config) (*sql.DB, error) {
	log.Printf("Connecting to Auth PostgreSQL (%s:%d)...", cfg.AuthDB.Host, cfg.AuthDB.Port)

	db, err := auth.NewConnection(auth.Config{
		Host:            cfg.AuthDB.Host,
		Port:            cfg.AuthDB.Port,
		User:            cfg.AuthDB.User,
		Password:        cfg.AuthDB.Password,
		Database:        cfg.AuthDB.Database,
		SSLMode:         cfg.AuthDB.SSLMode,
		MaxOpenConns:    cfg.AuthDB.MaxOpenConns,
		MaxIdleConns:    cfg.AuthDB.MaxIdleConns,
		ConnMaxLifetime: cfg.AuthDB.ConnMaxLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize auth database: %w", err)
	}

	db.SetConnMaxIdleTime(cfg.AuthDB.ConnMaxIdleTime)

	log.Println("Running auth database migrations...")
	if err := auth.RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run auth migrations: %w", err)
	}

	log.Printf("Auth database ready (db=%s, pool=%d/%d)",
		cfg.AuthDB.Database,
		cfg.AuthDB.MaxIdleConns,
		cfg.AuthDB.MaxOpenConns,
	)

	return db, nil
}
