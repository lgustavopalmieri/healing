package listener

import "errors"

var (
	ErrUnmarshalPayload   = errors.New("failed to unmarshal specialist updated event payload")
	ErrFindSpecialistByID = errors.New("failed to find specialist by ID")
	ErrUpdateProjection   = errors.New("failed to update read projection")
)
