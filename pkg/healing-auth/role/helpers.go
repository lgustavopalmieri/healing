package role

import autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"

func MustBeIn(actual Role, allowed []Role) error {
	if len(allowed) == 0 {
		return nil
	}
	for _, r := range allowed {
		if r == actual {
			return nil
		}
	}
	return autherrors.ErrForbiddenWrongRole
}
