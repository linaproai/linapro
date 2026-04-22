// Package menutype defines the canonical menu-type codes shared by host menu
// governance and plugin menu manifests.
package menutype

import "strings"

// Code identifies one persisted menu category.
type Code string

// Canonical menu-type codes.
const (
	// Directory marks a sidebar directory node.
	Directory Code = "D"
	// Menu marks a clickable routed page node.
	Menu Code = "M"
	// Button marks a hidden permission/button node.
	Button Code = "B"
)

// String returns the canonical persisted code value.
func (code Code) String() string {
	return string(code)
}

// Normalize converts a raw persisted value into one canonical menu code.
func Normalize(value string) Code {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case Directory.String():
		return Directory
	case Menu.String(), "":
		return Menu
	case Button.String():
		return Button
	default:
		return ""
	}
}

// IsSupported reports whether the menu code is one of the published values.
func IsSupported(code Code) bool {
	return code == Directory || code == Menu || code == Button
}
