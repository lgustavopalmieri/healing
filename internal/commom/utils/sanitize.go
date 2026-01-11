package utils

import "strings"

func SanitizeStringArray(keywords []string) []string {
	if len(keywords) == 0 {
		return []string{}
	}

	seen := make(map[string]bool)
	var normalized []string

	for _, keyword := range keywords {
		cleaned := strings.ToLower(strings.TrimSpace(keyword))
		if cleaned != "" && !seen[cleaned] {
			seen[cleaned] = true
			normalized = append(normalized, cleaned)
		}
	}

	return normalized
}
