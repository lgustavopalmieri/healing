package authutil

import "time"

func RemainingTTL(expiresAt time.Time) time.Duration {
	remaining := time.Until(expiresAt)
	if remaining <= 0 {
		return 0
	}
	return remaining
}
