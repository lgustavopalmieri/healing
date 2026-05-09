package tokenissuer

import "crypto/rsa"

type Keyring struct {
	CurrentKID  string
	PrivateKeys map[string]*rsa.PrivateKey
	PublicKeys  map[string]*rsa.PublicKey
}

func NewKeyring(currentKID string, privateKey *rsa.PrivateKey) *Keyring {
	return &Keyring{
		CurrentKID: currentKID,
		PrivateKeys: map[string]*rsa.PrivateKey{
			currentKID: privateKey,
		},
		PublicKeys: map[string]*rsa.PublicKey{
			currentKID: &privateKey.PublicKey,
		},
	}
}
