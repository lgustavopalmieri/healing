package bootstrap

import (
	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
	sdktoken "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token"
)

func InitAuthTokenValidator(cfg *config.Config, keyring *tokenissuer.Keyring) *sdktoken.JWTValidator {
	return sdktoken.NewJWTValidator(sdktoken.JWTValidatorConfig{
		PublicKeys: keyring.PublicKeys,
		Issuer:     cfg.Auth.Issuer,
		Audience:   cfg.Auth.Audience,
	})
}
