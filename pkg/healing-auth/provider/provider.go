package provider

import "errors"

type Provider string

const (
	Password  Provider = "password"
	Google    Provider = "google"
	Instagram Provider = "instagram"
	Biometric Provider = "biometric"
)

var ErrInvalidProvider = errors.New("invalid provider")

func Parse(s string) (Provider, error) {
	p := Provider(s)
	if !p.Valid() {
		return "", ErrInvalidProvider
	}
	return p, nil
}

func (p Provider) Valid() bool {
	switch p {
	case Password, Google, Instagram, Biometric:
		return true
	}
	return false
}

func (p Provider) String() string { return string(p) }

func (p Provider) RequiresPassword() bool { return p == Password }
func (p Provider) IsExternal() bool       { return p == Google || p == Instagram }
