package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func GenerateRSAKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key, &key.PublicKey
}

type SignOptions struct {
	KID           string
	SigningMethod jwtlib.SigningMethod
	Claims        jwtlib.Claims
}

func SignTokenWithKID(t *testing.T, key *rsa.PrivateKey, opts SignOptions) string {
	t.Helper()
	method := opts.SigningMethod
	if method == nil {
		method = jwtlib.SigningMethodRS256
	}
	token := jwtlib.NewWithClaims(method, opts.Claims)
	if opts.KID != "" {
		token.Header["kid"] = opts.KID
	}
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func SignTokenHS256(t *testing.T, secret []byte, kid string, c jwtlib.Claims) string {
	t.Helper()
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, c)
	if kid != "" {
		token.Header["kid"] = kid
	}
	signed, err := token.SignedString(secret)
	require.NoError(t, err)
	return signed
}
