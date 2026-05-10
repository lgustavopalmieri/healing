package database

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/register-credential/event_listeners/create_specialist_credential/listener"
)

func NewCredentialDatabaseRepository(db *sql.DB) listener.CredentialRepository {
	return &CredentialDatabaseRepository{
		DB: db,
	}
}
