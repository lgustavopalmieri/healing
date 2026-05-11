package application

import (
	tokenpair "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/token_pair"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type SetPasswordDTO struct {
	Token      string
	Password   string
	DeviceInfo string
	IPAddress  string
	UserAgent  string
}

type SetPasswordResult struct {
	TokenPair *tokenpair.TokenPair
	SubjectID string
	Role      role.Role
}

type ValidatedSetPasswordToken struct {
	SubjectID string
	Role      role.Role
	JTI       string
}
