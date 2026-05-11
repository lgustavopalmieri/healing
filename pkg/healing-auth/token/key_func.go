package token

import (
	"crypto/rsa"

	jwtlib "github.com/golang-jwt/jwt/v5"

	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
)

func rsaKeyFunc(publicKeys map[string]*rsa.PublicKey) jwtlib.Keyfunc {
	return func(t *jwtlib.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodRSA); !ok {
			return nil, autherrors.ErrInvalidToken
		}
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, autherrors.ErrInvalidToken
		}
		key, ok := publicKeys[kid]
		if !ok {
			return nil, autherrors.ErrInvalidToken
		}
		return key, nil
	}
}

func containsAudience(aud jwtlib.ClaimStrings, expected string) bool {
	for _, a := range aud {
		if a == expected {
			return true
		}
	}
	return false
}

func firstAudience(aud jwtlib.ClaimStrings) string {
	if len(aud) > 0 {
		return aud[0]
	}
	return ""
}
