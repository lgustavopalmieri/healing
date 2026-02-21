package domain

type SpecialistStatus string

const (
	StatusPending     SpecialistStatus = "pending"
	StatusActive      SpecialistStatus = "active"
	StatusUnavailable SpecialistStatus = "unavailable"
	StatusDeleted     SpecialistStatus = "deleted"
	StatusBanned      SpecialistStatus = "banned"
)

func (s SpecialistStatus) IsSearchable() bool {
	return s == StatusActive
}
