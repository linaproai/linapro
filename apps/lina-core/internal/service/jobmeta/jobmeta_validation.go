// This file contains scheduled-job enum validation and i18n key normalization
// helpers used by management and registry code.

package jobmeta

import (
	"strings"
)

// messageKeyPath converts backend anchors into dotted i18n key fragments.
func messageKeyPath(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return ""
	}

	var builder strings.Builder
	previousDot := false
	for _, current := range trimmed {
		switch {
		case current >= 'a' && current <= 'z',
			current >= '0' && current <= '9',
			current == '-':
			builder.WriteRune(current)
			previousDot = false
		default:
			if !previousDot {
				builder.WriteRune('.')
				previousDot = true
			}
		}
	}
	return strings.Trim(builder.String(), ".")
}
