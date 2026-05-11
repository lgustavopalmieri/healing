package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/session"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/sqlnull"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type SessionDatabaseRepository struct {
	DB *sql.DB
}

func NewSessionDatabaseRepository(db *sql.DB) *SessionDatabaseRepository {
	return &SessionDatabaseRepository{DB: db}
}

func (r *SessionDatabaseRepository) Save(ctx context.Context, s *session.Session) error {
	query := `
		INSERT INTO sessions (
			id, subject_id, role, refresh_token_hash,
			device_info, ip_address, user_agent,
			expires_at, revoked_at, last_used_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.DB.ExecContext(ctx, query,
		s.ID,
		s.SubjectID,
		s.Role.String(),
		s.RefreshTokenHash,
		sqlnull.String(s.DeviceInfo),
		sqlnull.String(s.IPAddress),
		sqlnull.String(s.UserAgent),
		s.ExpiresAt,
		sqlnull.Time(s.RevokedAt),
		sqlnull.Time(s.LastUsedAt),
		s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf(FailedToSaveSessionErr, err)
	}
	return nil
}

func (r *SessionDatabaseRepository) FindByRefreshTokenHash(ctx context.Context, hash string) (*session.Session, error) {
	query := `
		SELECT id, subject_id, role, refresh_token_hash,
		       device_info, ip_address, user_agent,
		       expires_at, revoked_at, last_used_at, created_at
		FROM sessions
		WHERE refresh_token_hash = $1
		LIMIT 1`

	var (
		s          session.Session
		roleStr    string
		deviceInfo sql.NullString
		ipAddress  sql.NullString
		userAgent  sql.NullString
		revokedAt  sql.NullTime
		lastUsedAt sql.NullTime
	)

	err := r.DB.QueryRowContext(ctx, query, hash).Scan(
		&s.ID,
		&s.SubjectID,
		&roleStr,
		&s.RefreshTokenHash,
		&deviceInfo,
		&ipAddress,
		&userAgent,
		&s.ExpiresAt,
		&revokedAt,
		&lastUsedAt,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf(FailedToFindSessionErr, err)
	}

	parsedRole, err := role.Parse(roleStr)
	if err != nil {
		return nil, fmt.Errorf(FailedToFindSessionErr, err)
	}
	s.Role = parsedRole
	if deviceInfo.Valid {
		s.DeviceInfo = deviceInfo.String
	}
	if ipAddress.Valid {
		s.IPAddress = ipAddress.String
	}
	if userAgent.Valid {
		s.UserAgent = userAgent.String
	}
	if revokedAt.Valid {
		s.RevokedAt = &revokedAt.Time
	}
	if lastUsedAt.Valid {
		s.LastUsedAt = &lastUsedAt.Time
	}
	return &s, nil
}

func (r *SessionDatabaseRepository) Revoke(ctx context.Context, sessionID string) error {
	query := `UPDATE sessions SET revoked_at = $1 WHERE id = $2 AND revoked_at IS NULL`

	now := time.Now().UTC()
	result, err := r.DB.ExecContext(ctx, query, now, sessionID)
	if err != nil {
		return fmt.Errorf(FailedToRevokeSessionErr, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(FailedToRevokeSessionErr, err)
	}
	if rows == 0 {
		return fmt.Errorf(FailedToRevokeSessionErr, sql.ErrNoRows)
	}
	return nil
}

func (r *SessionDatabaseRepository) RevokeAllForSubject(ctx context.Context, subjectID string, r2 role.Role) (int64, error) {
	query := `UPDATE sessions SET revoked_at = $1 WHERE subject_id = $2 AND role = $3 AND revoked_at IS NULL`

	now := time.Now().UTC()
	result, err := r.DB.ExecContext(ctx, query, now, subjectID, r2.String())
	if err != nil {
		return 0, fmt.Errorf(FailedToRevokeSessionErr, err)
	}
	return result.RowsAffected()
}
