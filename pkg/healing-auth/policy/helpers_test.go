package policy_test

import (
	"time"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func validClaimsFactory(overrides ...func(*claims.Claims)) *claims.Claims {
	c := &claims.Claims{
		Subject:  "subject-abc",
		Role:     role.Specialist,
		Email:    "specialist@healing.com",
		Provider: provider.Password,
		TokenID:  "token-id-xyz",
		IssuedAt: time.Now().Add(-5 * time.Minute),
		ExpireAt: time.Now().Add(55 * time.Minute),
		Issuer:   "healing-specialist",
		Audience: "healing-platform",
	}
	for _, o := range overrides {
		o(c)
	}
	return c
}
