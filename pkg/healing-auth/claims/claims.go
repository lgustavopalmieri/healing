package claims

import (
	"time"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type Claims struct {
	Subject  string
	Role     role.Role
	Email    string
	Provider provider.Provider
	TokenID  string
	IssuedAt time.Time
	ExpireAt time.Time
	Issuer   string
	Audience string
}

func (c *Claims) IsExpired(now time.Time) bool {
	return now.After(c.ExpireAt)
}

func (c *Claims) Valid() bool {
	return c != nil && c.Subject != "" && c.Role.Valid() && c.Provider.Valid() && c.TokenID != ""
}
