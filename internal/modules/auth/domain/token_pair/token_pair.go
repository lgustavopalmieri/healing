package tokenpair

import "time"

type TokenPair struct {
	AccessToken      string
	AccessJTI        string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}
