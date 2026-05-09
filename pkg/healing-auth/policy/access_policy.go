package policy

import "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"

type AccessPolicy struct {
	AllowedRoles     []role.Role
	RequireOwnership bool
	AllowPublic      bool
}

func PublicAccess() AccessPolicy {
	return AccessPolicy{AllowPublic: true}
}

func AuthenticatedAccess(roles ...role.Role) AccessPolicy {
	return AccessPolicy{
		AllowedRoles:     roles,
		RequireOwnership: false,
		AllowPublic:      false,
	}
}

func OwnedAccess(r role.Role) AccessPolicy {
	return AccessPolicy{
		AllowedRoles:     []role.Role{r},
		RequireOwnership: true,
		AllowPublic:      false,
	}
}

func AdminReadOnly() AccessPolicy {
	return AccessPolicy{
		AllowedRoles:     []role.Role{role.Admin},
		RequireOwnership: false,
		AllowPublic:      false,
	}
}
