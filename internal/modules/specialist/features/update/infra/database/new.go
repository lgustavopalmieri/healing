package database

import (
	"database/sql"

	eventlisteners "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/infra/event_listeners"
)

func NewSpecialistFindByIDRepository(db *sql.DB) eventlisteners.FindByIDRepositoryInterface {
	return &SpecialistFindByIDRepository{
		db: db,
	}
}
