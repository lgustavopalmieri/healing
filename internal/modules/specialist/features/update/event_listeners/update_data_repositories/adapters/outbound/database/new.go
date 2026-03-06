package database

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/event_listeners/update_data_repositories/command"
)

func NewSourceRepository(db *sql.DB) command.SourceRepository {
	return &SourceRepository{
		DB: db,
	}
}
