package audit

type EventType string

const (
	EventLoginSuccess           EventType = "login_success"
	EventLoginFailure           EventType = "login_failure"
	EventLogout                 EventType = "logout"
	EventPasswordSet            EventType = "password_set"
	EventPasswordReset          EventType = "password_reset"
	EventPasswordChanged        EventType = "password_changed"
	EventSessionRevoked         EventType = "session_revoked"
	EventRevokeAllSessions      EventType = "revoke_all_sessions"
	EventCredentialLocked       EventType = "credential_locked"
	EventAdminAccessResource    EventType = "admin_access_resource"
	EventAccessDenied           EventType = "access_denied"
	EventPasswordResetRequested EventType = "password_reset_requested"
)

type Category string

const (
	CategoryAuthentication Category = "authentication"
	CategoryPassword       Category = "password"
	CategoryAdmin          Category = "admin"
	CategorySecurity       Category = "security"
)

func (e EventType) Category() Category {
	switch e {
	case EventLoginSuccess, EventLoginFailure, EventLogout:
		return CategoryAuthentication
	case EventPasswordSet, EventPasswordReset, EventPasswordChanged, EventPasswordResetRequested:
		return CategoryPassword
	case EventAdminAccessResource:
		return CategoryAdmin
	}
	return CategorySecurity
}
