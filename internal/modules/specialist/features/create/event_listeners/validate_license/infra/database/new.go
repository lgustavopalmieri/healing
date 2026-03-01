package database

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/listener"
)

func NewValidateLicenseRepository(db *sql.DB) listener.ValidateLicenseRepositoryInterface {
	return &ValidateLicenseRepository{
		DB: db,
	}
}
