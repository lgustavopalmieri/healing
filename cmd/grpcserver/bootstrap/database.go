package bootstrap

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/cmd/grpcserver/config"
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
	db, err := postgresql.NewConnection(postgresql.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Database,
		SSLMode:         defaultSSLMode,
		MaxOpenConns:    defaultMaxOpenConns,
		MaxIdleConns:    defaultMaxIdleConns,
		ConnMaxLifetime: defaultConnMaxLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	db.SetConnMaxIdleTime(defaultConnMaxIdleTime)

	return db, nil
}
