package tokenissuer_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
)

func TestNewKeyring(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	const kid = "healing-2026-05"

	t.Run("happy path - cria keyring com CurrentKID, PrivateKeys e PublicKeys populados", func(t *testing.T) {
		kr := tokenissuer.NewKeyring(kid, privateKey)

		require.NotNil(t, kr)
		assert.Equal(t, kid, kr.CurrentKID)

		require.Contains(t, kr.PrivateKeys, kid)
		assert.Same(t, privateKey, kr.PrivateKeys[kid])

		require.Contains(t, kr.PublicKeys, kid)
		assert.NotNil(t, kr.PublicKeys[kid])
	})

	t.Run("happy path - chave publica no keyring e derivada da privada", func(t *testing.T) {
		kr := tokenissuer.NewKeyring(kid, privateKey)

		pub := kr.PublicKeys[kid]
		require.NotNil(t, pub)
		assert.Equal(t, privateKey.PublicKey.N, pub.N)
		assert.Equal(t, privateKey.PublicKey.E, pub.E)
	})
}
