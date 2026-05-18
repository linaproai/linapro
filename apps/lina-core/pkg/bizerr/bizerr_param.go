// This file defines runtime message parameters used for template rendering.

package bizerr

import "strings"

// Param carries one named value used when rendering a runtime message template.
type Param struct {
	Name  string // Name is the placeholder name without braces.
	Value any    // Value is the placeholder value.
}

// P constructs one named runtime-message parameter.
func P(name string, value any) Param {
	return Param{Name: strings.TrimSpace(name), Value: value}
}
