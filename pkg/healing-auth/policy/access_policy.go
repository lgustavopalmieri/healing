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
	if len(roles) == 0 {
		roles = []role.Role{role.Specialist, role.Patient, role.Admin}
	}
	return AccessPolicy{
		AllowedRoles:     append([]role.Role(nil), roles...),
		RequireOwnership: false,
		AllowPublic:      false,
	}
}

func AnyAuthenticated() AccessPolicy {
	return AuthenticatedAccess()
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
