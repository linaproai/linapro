// This file implements message key derivation and fallback template rendering.

package bizerr

import (
	"fmt"
	"strings"
)

// MessageKey derives the default runtime i18n key from a stable error code.
func MessageKey(errorCode string) string {
	segments := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(errorCode)), func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == ' '
	})
	if len(segments) == 0 {
		return ""
	}
	return "error." + strings.Join(segments, ".")
}

// Format renders a message template by replacing `{name}` placeholders with
// values from params.
func Format(template string, params map[string]any) string {
	rendered := template
	for key, value := range params {
		placeholder := "{" + key + "}"
		rendered = strings.ReplaceAll(rendered, placeholder, fmt.Sprint(value))
	}
	return rendered
}
