package password

import "golang.org/x/crypto/bcrypt"

type Password struct {
	Raw string
}

func NewPassword(raw string, cfg ValidationConfig) (Password, error) {
	if err := validate(raw, cfg); err != nil {
		return Password{}, err
	}
	return Password{Raw: raw}, nil
}

func (p Password) Hash(cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(p.Raw), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func Matches(rawAttempt, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(rawAttempt)) == nil
}
