package update

import "strings"

func validateID(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidID
	}
	return nil
}
