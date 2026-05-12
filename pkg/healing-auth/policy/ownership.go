package policy

import (
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func MustMatchSubject(c *claims.Claims, resourceOwnerID string, expectedRole role.Role) error {
	if c == nil || !c.Valid() {
		return autherrors.ErrUnauthenticated
	}
	if c.Role != expectedRole {
		return autherrors.ErrForbiddenWrongRole
	}
	if c.Subject != resourceOwnerID {
		return autherrors.ErrForbiddenNotOwner
	}
	return nil
}
