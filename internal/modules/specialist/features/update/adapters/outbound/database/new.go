package database

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
)

func NewSpecialistUpdateRepository(db *sql.DB) application.SpecialistUpdateRepositoryInterface {
	return &SpecialistUpdateRepository{
		db: db,
	}
}
