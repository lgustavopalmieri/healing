package admin

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusInactive:
		return true
	}
	return false
}

func (s Status) CanTransitionTo(t Status) bool {
	switch s {
	case StatusActive:
		return t == StatusInactive
	case StatusInactive:
		return t == StatusActive
	}
	return false
}
