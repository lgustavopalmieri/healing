package database

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application/listener"
)

func NewSpecialistFindByIDRepository(db *sql.DB) listener.SpecialistFindByIDRepositoryInterface {
	return &SpecialistFindByIDRepository{
		db: db,
	}
}
