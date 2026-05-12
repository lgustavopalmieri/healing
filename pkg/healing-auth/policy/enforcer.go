package policy

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type Enforcer interface {
	Enforce(ctx context.Context, p AccessPolicy, c *claims.Claims, resourceOwnerID string) error
	EnforceRoleOnly(ctx context.Context, p AccessPolicy, c *claims.Claims) error
}

type LocalEnforcer struct{}

func NewLocalEnforcer() *LocalEnforcer { return &LocalEnforcer{} }

func (e *LocalEnforcer) Enforce(ctx context.Context, p AccessPolicy, c *claims.Claims, resourceOwnerID string) error {
	if p.AllowPublic {
		return nil
	}
	if len(p.AllowedRoles) == 0 {
		return autherrors.ErrForbidden
	}
	if c == nil || !c.Valid() {
		return autherrors.ErrUnauthenticated
	}
	if err := role.MustBeIn(c.Role, p.AllowedRoles); err != nil {
		return err
	}
	if p.RequireOwnership {
		if c.Subject != resourceOwnerID {
			return autherrors.ErrForbiddenNotOwner
		}
	}
	return nil
}

func (e *LocalEnforcer) EnforceRoleOnly(ctx context.Context, p AccessPolicy, c *claims.Claims) error {
	if p.AllowPublic {
		return nil
	}
	if len(p.AllowedRoles) == 0 {
		return autherrors.ErrForbidden
	}
	if c == nil || !c.Valid() {
		return autherrors.ErrUnauthenticated
	}
	return role.MustBeIn(c.Role, p.AllowedRoles)
}
