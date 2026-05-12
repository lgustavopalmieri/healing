package application

import (
	tokenpair "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/token_pair"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type LoginDTO struct {
	Email        string
	Password     string
	ExpectedRole string
	DeviceInfo   string
	IPAddress    string
	UserAgent    string
}

type LoginResult struct {
	TokenPair *tokenpair.TokenPair
	SubjectID string
	Role      role.Role
}
