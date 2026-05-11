package credential

import (
	"database/sql"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/password"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRow(scanner rowScanner) (*credential.Credential, error) {
	var c credential.Credential
	var (
		roleStr        string
		providerStr    string
		statusStr      string
		providerUserID sql.NullString
		passwordHash   sql.NullString
		lastUsedAt     sql.NullTime
	)

	if err := scanner.Scan(
		&c.ID,
		&c.SubjectID,
		&roleStr,
		&providerStr,
		&providerUserID,
		&passwordHash,
		&c.Email,
		&statusStr,
		&lastUsedAt,
		&c.CreatedAt,
		&c.UpdatedAt,
	); err != nil {
		return nil, err
	}

	parsedRole, err := role.Parse(roleStr)
	if err != nil {
		return nil, fmt.Errorf(FailedToFindCredentialErr, err)
	}
	parsedProvider, err := provider.Parse(providerStr)
	if err != nil {
		return nil, fmt.Errorf(FailedToFindCredentialErr, err)
	}

	c.Role = parsedRole
	c.Provider = parsedProvider
	c.Status = credential.Status(statusStr)
	if providerUserID.Valid {
		c.ProviderUserID = providerUserID.String
	}
	if passwordHash.Valid {
		c.PasswordHash = password.NewHashedPassword(passwordHash.String)
	}
	if lastUsedAt.Valid {
		c.LastUsedAt = &lastUsedAt.Time
	}
	return &c, nil
}
