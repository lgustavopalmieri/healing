package role

import "errors"

type Role string

const (
	Anonymous  Role = "anonymous"
	Specialist Role = "specialist"
	Patient    Role = "patient"
	Admin      Role = "admin"
)

var ErrInvalidRole = errors.New("invalid role")

func Parse(s string) (Role, error) {
	r := Role(s)
	if !r.Valid() {
		return "", ErrInvalidRole
	}
	return r, nil
}

func (r Role) Valid() bool {
	switch r {
	case Anonymous, Specialist, Patient, Admin:
		return true
	}
	return false
}

func (r Role) String() string {
	return string(r)
}

func (r Role) IsAnonymous() bool  { return r == Anonymous }
func (r Role) IsSpecialist() bool { return r == Specialist }
func (r Role) IsPatient() bool    { return r == Patient }
func (r Role) IsAdmin() bool      { return r == Admin }
