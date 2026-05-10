package listener

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

//go:generate mockgen -source=interface.go -destination=mocks/mocks.go -package=mocks
type CredentialRepository interface {
	FindByEmailProviderRole(ctx context.Context, email string, p provider.Provider, r role.Role) (*credential.Credential, error)
	Save(ctx context.Context, c *credential.Credential) error
}

type SetPasswordTokenGenerator interface {
	Generate(ctx context.Context, subjectID string) (tokenString string, jti string, err error)
}
