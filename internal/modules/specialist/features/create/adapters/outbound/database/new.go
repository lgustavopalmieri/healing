package database

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
)

func NewSpecialistCreateRepository(db *sql.DB) application.SpecialistCreateRepositoryInterface {
	return &SpecialistCreateRepository{
		db: db,
	}
}
