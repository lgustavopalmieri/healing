package tokenissuer_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
)

func writePrivateKeyPEM(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "priv.pem")

	data := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: data}

	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, pem.Encode(f, block))

	return path
}

func writePublicKeyPEM(t *testing.T, pub *rsa.PublicKey) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "pub.pem")

	data, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: data}

	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, pem.Encode(f, block))

	return path
}

func writeRawFile(t *testing.T, content string, name string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestLoadPrivateKey(t *testing.T) {
	validKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	validPath := writePrivateKeyPEM(t, validKey)
	publicOnlyPath := writePublicKeyPEM(t, &validKey.PublicKey)
	garbagePath := writeRawFile(t, "not a pem file", "garbage.pem")

	tests := []struct {
		name        string
		path        string
		expectError bool
		errContains string
		validate    func(t *testing.T, key *rsa.PrivateKey)
	}{
		{
			name: "happy path - carrega chave privada PEM valida",
			path: validPath,
			validate: func(t *testing.T, key *rsa.PrivateKey) {
				require.NotNil(t, key)
				assert.Equal(t, validKey.N, key.N)
			},
		},
		{
			name:        "failure - arquivo inexistente retorna erro com 'read private key'",
			path:        "/nonexistent/path/priv.pem",
			expectError: true,
			errContains: "read private key",
		},
		{
			name:        "failure - arquivo com conteudo nao-PEM retorna erro com 'parse private key'",
			path:        garbagePath,
			expectError: true,
			errContains: "parse private key",
		},
		{
			name:        "failure - arquivo contendo apenas chave publica retorna erro de parse",
			path:        publicOnlyPath,
			expectError: true,
			errContains: "parse private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := tokenissuer.LoadPrivateKey(tt.path)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, key)
				return
			}
			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, key)
			}
		})
	}
}

func TestLoadPublicKey(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	validPath := writePublicKeyPEM(t, &privKey.PublicKey)
	garbagePath := writeRawFile(t, "not a pem file", "garbage.pem")

	tests := []struct {
		name        string
		path        string
		expectError bool
		errContains string
		validate    func(t *testing.T, key *rsa.PublicKey)
	}{
		{
			name: "happy path - carrega chave publica PEM valida",
			path: validPath,
			validate: func(t *testing.T, key *rsa.PublicKey) {
				require.NotNil(t, key)
				assert.Equal(t, privKey.PublicKey.N, key.N)
			},
		},
		{
			name:        "failure - arquivo inexistente retorna erro com 'read public key'",
			path:        "/nonexistent/path/pub.pem",
			expectError: true,
			errContains: "read public key",
		},
		{
			name:        "failure - arquivo com conteudo nao-PEM retorna erro com 'parse public key'",
			path:        garbagePath,
			expectError: true,
			errContains: "parse public key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := tokenissuer.LoadPublicKey(tt.path)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, key)
				return
			}
			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, key)
			}
		})
	}
}
