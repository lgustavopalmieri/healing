package application

import "time"

type LogoutDTO struct {
	RefreshToken   string
	AccessTokenJTI string
	AccessTokenExp time.Time
	SubjectID      string
	Role           string
	IPAddress      string
	UserAgent      string
}
