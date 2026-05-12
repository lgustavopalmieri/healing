package admin

type SubRole string

const (
	SubRoleAdmin     SubRole = "admin"
	SubRoleSupport   SubRole = "support"
	SubRoleModerator SubRole = "moderator"
)

func (s SubRole) Valid() bool {
	switch s {
	case SubRoleAdmin, SubRoleSupport, SubRoleModerator:
		return true
	}
	return false
}

func (s SubRole) CanReadSpecialists() bool {
	return s == SubRoleAdmin || s == SubRoleSupport || s == SubRoleModerator
}

func (s SubRole) CanReadPatients() bool {
	return s == SubRoleAdmin || s == SubRoleSupport
}

func (s SubRole) CanBlockAccounts() bool {
	return s == SubRoleAdmin
}

func (s SubRole) CanViewAuditLogs() bool {
	return s == SubRoleAdmin || s == SubRoleSupport
}

func (s SubRole) CanModerateReviews() bool {
	return s == SubRoleAdmin || s == SubRoleModerator
}
