package credential

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/session"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/sqlnull"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

const selectColumns = `
	id, subject_id, role, provider, provider_user_id,
	password_hash, email, status, last_used_at, created_at, updated_at`

type CredentialDatabaseRepository struct {
	DB *sql.DB
}

func NewCredentialDatabaseRepository(db *sql.DB) *CredentialDatabaseRepository {
	return &CredentialDatabaseRepository{DB: db}
}

func (r *CredentialDatabaseRepository) Save(ctx context.Context, c *credential.Credential) error {
	query := `
		INSERT INTO credentials (
			id, subject_id, role, provider, provider_user_id,
			password_hash, email, status, last_used_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.DB.ExecContext(
		ctx,
		query,
		c.ID,
		c.SubjectID,
		c.Role.String(),
		c.Provider.String(),
		sqlnull.String(c.ProviderUserID),
		sqlnull.String(c.PasswordHash.String()),
		c.Email,
		string(c.Status),
		sqlnull.Time(c.LastUsedAt),
		c.CreatedAt,
		c.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code.Name() == "unique_violation" || strings.Contains(pqErr.Message, "duplicate key") {
				return errors.New(CredentialAlreadyExistsErr)
			}
		}
		return fmt.Errorf(FailedToSaveCredentialErr, err)
	}
	return nil
}

func (r *CredentialDatabaseRepository) FindByEmailProviderRole(
	ctx context.Context,
	email string,
	p provider.Provider,
	rl role.Role,
) (*credential.Credential, error) {
	query := `SELECT ` + selectColumns + `
		FROM credentials
		WHERE email = $1 AND provider = $2 AND role = $3 AND status != 'deleted'
		LIMIT 1`

	cred, err := scanRow(r.DB.QueryRowContext(ctx, query, email, p.String(), rl.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf(FailedToFindCredentialErr, err)
	}
	return cred, nil
}

func (r *CredentialDatabaseRepository) FindBySubjectAndRole(
	ctx context.Context,
	subjectID string,
	rl role.Role,
) (*credential.Credential, error) {
	query := `SELECT ` + selectColumns + `
		FROM credentials
		WHERE subject_id = $1 AND role = $2 AND status != 'deleted'
		LIMIT 1`

	cred, err := scanRow(r.DB.QueryRowContext(ctx, query, subjectID, rl.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf(FailedToFindCredentialErr, err)
	}
	return cred, nil
}

func (r *CredentialDatabaseRepository) Update(ctx context.Context, c *credential.Credential) error {
	return updateCredential(ctx, r.DB, c)
}

func (r *CredentialDatabaseRepository) UpdateWithSessionInTransaction(
	ctx context.Context,
	cred *credential.Credential,
	sess *session.Session,
) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := updateCredential(ctx, tx, cred); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("%w (rollback error: %v)", err, rollbackErr)
		}
		return err
	}

	if err := insertSession(ctx, tx, sess); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("%w (rollback error: %v)", err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func updateCredential(ctx context.Context, exec execer, c *credential.Credential) error {
	query := `
		UPDATE credentials
		SET password_hash = $1,
		    status = $2,
		    updated_at = $3,
		    last_used_at = $4
		WHERE id = $5`

	result, err := exec.ExecContext(ctx, query,
		sqlnull.String(c.PasswordHash.String()),
		string(c.Status),
		c.UpdatedAt,
		sqlnull.Time(c.LastUsedAt),
		c.ID,
	)
	if err != nil {
		return fmt.Errorf(FailedToUpdateCredentialErr, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(FailedToUpdateCredentialErr, err)
	}
	if rows == 0 {
		return fmt.Errorf(FailedToUpdateCredentialErr, sql.ErrNoRows)
	}
	return nil
}

func insertSession(ctx context.Context, exec execer, s *session.Session) error {
	query := `
		INSERT INTO sessions (
			id, subject_id, role, refresh_token_hash,
			device_info, ip_address, user_agent,
			expires_at, revoked_at, last_used_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := exec.ExecContext(ctx, query,
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
