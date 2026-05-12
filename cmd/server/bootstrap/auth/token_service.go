package auth

import (
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
)

func InitTokenService(cfg *config.Config) (*tokenissuer.Signer, *tokenissuer.Keyring, error) {
	privKey, err := tokenissuer.LoadPrivateKey(cfg.Auth.PrivateKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load auth private key: %w", err)
	}

	keyring := tokenissuer.NewKeyring(cfg.Auth.CurrentKeyID, privKey)

	signer := tokenissuer.NewSigner(keyring, tokenissuer.SignerConfig{
		Issuer:   cfg.Auth.Issuer,
		Audience: cfg.Auth.Audience,
	})

	return signer, keyring, nil
}
