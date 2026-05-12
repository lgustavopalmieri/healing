package password

import "unicode"

type ValidationConfig struct {
	MinLength int
}

func validate(raw string, cfg ValidationConfig) error {
	if len(raw) < cfg.MinLength {
		return ErrTooShort
	}

	var hasLetter, hasDigit bool
	for _, r := range raw {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return ErrMissingRequiredChars
	}
	return nil
}
