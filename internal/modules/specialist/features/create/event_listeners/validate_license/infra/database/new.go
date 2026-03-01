package database

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/application"
)

func NewValidateLicenseRepository(db *sql.DB) application.ValidateLicenseRepositoryInterface {
	return &ValidateLicenseRepository{
		DB: db,
	}
}
