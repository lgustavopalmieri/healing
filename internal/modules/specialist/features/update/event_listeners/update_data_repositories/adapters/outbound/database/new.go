package database

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/event_listeners/update_data_repositories/listener"
)

func NewSourceRepository(db *sql.DB) listener.SourceRepository {
	return &SourceRepository{
		DB: db,
	}
}
