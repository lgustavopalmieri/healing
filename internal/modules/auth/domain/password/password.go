package password

import "golang.org/x/crypto/bcrypt"

type Password struct {
	raw string
}

func NewPassword(raw string, cfg ValidationConfig) (Password, error) {
	if err := validate(raw, cfg); err != nil {
		return Password{}, err
	}
	return Password{raw: raw}, nil
}

func (p Password) Hash(cost int) (HashedPassword, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(p.raw), cost)
	if err != nil {
		return HashedPassword{}, err
	}
	return HashedPassword{value: string(hash)}, nil
}

func (p Password) Matches(hashed HashedPassword) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed.value), []byte(p.raw)) == nil
}

type HashedPassword struct {
	value string
}

func NewHashedPassword(v string) HashedPassword {
	return HashedPassword{value: v}
}

func (h HashedPassword) String() string {
	return h.value
}

func (h HashedPassword) IsEmpty() bool {
	return h.value == ""
}
