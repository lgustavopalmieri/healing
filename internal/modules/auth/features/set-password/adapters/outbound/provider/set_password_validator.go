package provider

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/application"
	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
	sdktoken "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token"
)

const setPasswordPurpose = "set-password"

type SetPasswordTokenValidatorConfig struct {
	Keyring  *tokenissuer.Keyring
	Issuer   string
	Audience string
}

type SetPasswordTokenValidator struct {
	inner *sdktoken.SpecialPurposeJWTValidator
}

func NewSetPasswordTokenValidator(cfg SetPasswordTokenValidatorConfig) *SetPasswordTokenValidator {
	return &SetPasswordTokenValidator{
		inner: sdktoken.NewSpecialPurposeJWTValidator(sdktoken.SpecialPurposeJWTValidatorConfig{
			PublicKeys: cfg.Keyring.PublicKeys,
			Issuer:     cfg.Issuer,
			Audience:   cfg.Audience,
		}),
	}
}

func (v *SetPasswordTokenValidator) Validate(ctx context.Context, rawToken string) (*application.ValidatedSetPasswordToken, error) {
	c, err := v.inner.Validate(ctx, rawToken, setPasswordPurpose)
	if err != nil {
		return nil, err
	}
	return &application.ValidatedSetPasswordToken{
		SubjectID: c.Subject,
		Role:      c.Role,
		JTI:       c.TokenID,
	}, nil
}
