package domain

type SpecialistStatus string

const (
	StatusPending           SpecialistStatus = "pending"
	StatusAuthorizedLicense SpecialistStatus = "authorized_license"
	StatusActive            SpecialistStatus = "active"
	StatusUnavailable       SpecialistStatus = "unavailable"
	StatusDeleted           SpecialistStatus = "deleted"
	StatusBanned            SpecialistStatus = "banned"
)

func (s SpecialistStatus) IsSearchable() bool {
	return s == StatusActive || s == StatusAuthorizedLicense
}

func SearchableStatuses() []string {
	return []string{string(StatusActive), string(StatusAuthorizedLicense)}
}
