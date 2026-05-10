package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type CredentialDatabaseRepository struct {
	DB *sql.DB
}

func (r *CredentialDatabaseRepository) FindByEmailProviderRole(
	ctx context.Context,
	email string,
	p provider.Provider,
	rl role.Role,
) (*credential.Credential, error) {
	query := `
		SELECT id, subject_id, role, provider, provider_user_id,
		       password_hash, email, status, last_used_at, created_at, updated_at
		FROM credentials
		WHERE email = $1 AND provider = $2 AND role = $3 AND status != 'deleted'
		LIMIT 1`

	var (
		cred           credential.Credential
		roleStr        string
		providerStr    string
		statusStr      string
		providerUserID sql.NullString
		passwordHash   sql.NullString
		lastUsedAt     sql.NullTime
	)

	err := r.DB.QueryRowContext(ctx, query, email, p.String(), rl.String()).Scan(
		&cred.ID,
		&cred.SubjectID,
		&roleStr,
		&providerStr,
		&providerUserID,
		&passwordHash,
		&cred.Email,
		&statusStr,
		&lastUsedAt,
		&cred.CreatedAt,
		&cred.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf(FailedToFindCredentialErr, err)
	}

	parsedRole, err := role.Parse(roleStr)
	if err != nil {
		return nil, fmt.Errorf(FailedToFindCredentialErr, err)
	}
	parsedProvider, err := provider.Parse(providerStr)
	if err != nil {
		return nil, fmt.Errorf(FailedToFindCredentialErr, err)
	}

	cred.Role = parsedRole
	cred.Provider = parsedProvider
	cred.Status = credential.Status(statusStr)
	if providerUserID.Valid {
		cred.ProviderUserID = providerUserID.String
	}
	if passwordHash.Valid {
		cred.PasswordHash = passwordHash.String
	}
	if lastUsedAt.Valid {
		cred.LastUsedAt = &lastUsedAt.Time
	}

	return &cred, nil
}

func (r *CredentialDatabaseRepository) Save(ctx context.Context, c *credential.Credential) error {
	query := `
		INSERT INTO credentials (
			id, subject_id, role, provider, provider_user_id,
			password_hash, email, status, last_used_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	var providerUserID sql.NullString
	if c.ProviderUserID != "" {
		providerUserID = sql.NullString{String: c.ProviderUserID, Valid: true}
	}

	var passwordHash sql.NullString
	if c.PasswordHash != "" {
		passwordHash = sql.NullString{String: c.PasswordHash, Valid: true}
	}

	var lastUsedAt sql.NullTime
	if c.LastUsedAt != nil {
		lastUsedAt = sql.NullTime{Time: *c.LastUsedAt, Valid: true}
	}

	_, err := r.DB.ExecContext(
		ctx,
		query,
		c.ID,
		c.SubjectID,
		c.Role.String(),
		c.Provider.String(),
		providerUserID,
		passwordHash,
		c.Email,
		string(c.Status),
		lastUsedAt,
		c.CreatedAt,
		c.UpdatedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code.Name() == "unique_violation" || strings.Contains(pqErr.Message, "duplicate key") {
				return errors.New(CredentialAlreadyExistsErr)
			}
		}
		return fmt.Errorf(FailedToSaveCredentialErr, err)
	}
	return nil
}
