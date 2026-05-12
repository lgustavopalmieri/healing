package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/audit"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/sqlnull"
)

type AuditDatabaseRepository struct {
	DB *sql.DB
}

func NewAuditDatabaseRepository(db *sql.DB) *AuditDatabaseRepository {
	return &AuditDatabaseRepository{DB: db}
}

func (r *AuditDatabaseRepository) Save(ctx context.Context, log *audit.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			id, subject_id, role, event_type,
			ip_address, user_agent, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	var metadataJSON *[]byte
	if len(log.Metadata) > 0 {
		raw, err := json.Marshal(log.Metadata)
		if err != nil {
			return fmt.Errorf(FailedToSaveAuditLogErr, err)
		}
		metadataJSON = &raw
	}

	_, err := r.DB.ExecContext(ctx, query,
		log.ID,
		sqlnull.String(log.SubjectID),
		sqlnull.String(log.Role.String()),
		string(log.EventType),
		sqlnull.String(log.IPAddress),
		sqlnull.String(log.UserAgent),
		metadataJSON,
		log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf(FailedToSaveAuditLogErr, err)
	}
	return nil
}
