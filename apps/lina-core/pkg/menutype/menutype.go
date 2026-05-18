// Package menutype defines the canonical menu-type codes shared by host menu
// governance and plugin menu manifests.
package menutype

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
