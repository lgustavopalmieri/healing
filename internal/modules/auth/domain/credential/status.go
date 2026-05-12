package credential

type Status string

const (
	StatusPending Status = "pending"
	StatusActive  Status = "active"
	StatusLocked  Status = "locked"
	StatusDeleted Status = "deleted"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusActive, StatusLocked, StatusDeleted:
		return true
	}
	return false
}

func (s Status) CanTransitionTo(t Status) bool {
	switch s {
	case StatusPending:
		return t == StatusActive || t == StatusDeleted
	case StatusActive:
		return t == StatusLocked || t == StatusDeleted
	case StatusLocked:
		return t == StatusActive || t == StatusDeleted
	}
	return false
}
