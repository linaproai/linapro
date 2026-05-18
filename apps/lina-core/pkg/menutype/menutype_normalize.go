// This file contains menu-type normalization and support checks used by host
// menu governance and plugin manifests. It keeps parsing behavior separate from
// the package entrypoint while preserving the published helper functions.

package menutype

import "strings"

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
