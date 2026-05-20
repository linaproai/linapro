// Package statusflag defines small integer flag contracts shared by API DTOs.
// It deliberately avoids a single generic Status type because enabled,
// visibility, installation, read-state, and yes/no flags carry different
// business meanings even when they use the same persisted 0/1 values.
package statusflag

// Enabled identifies a common enabled/disabled flag.
type Enabled int

// Visibility identifies whether a UI-facing item should be shown.
type Visibility int

// Installation identifies whether an extension is installed.
type Installation int

// ReadState identifies whether an inbox item has been read.
type ReadState int

// YesNo identifies a generic persisted yes/no flag.
type YesNo int

// Common enabled-state values.
const (
	// Disabled marks an item as disabled.
	Disabled Enabled = 0
	// EnabledValue marks an item as enabled.
	EnabledValue Enabled = 1
)

// Common visibility values.
const (
	// Hidden marks an item as hidden.
	Hidden Visibility = 0
	// Visible marks an item as visible.
	Visible Visibility = 1
)

// Common installation values.
const (
	// Uninstalled marks an extension as not installed.
	Uninstalled Installation = 0
	// Installed marks an extension as installed.
	Installed Installation = 1
)

// Common read-state values.
const (
	// Unread marks an inbox item as unread.
	Unread ReadState = 0
	// Read marks an inbox item as read.
	Read ReadState = 1
)

// Common yes/no values.
const (
	// No marks a negative flag.
	No YesNo = 0
	// Yes marks an affirmative flag.
	Yes YesNo = 1
)

// Int returns the serialized integer value for an enabled flag.
func (value Enabled) Int() int {
	return int(value)
}

// IsSupported reports whether value is a published enabled flag.
func (value Enabled) IsSupported() bool {
	return value == Disabled || value == EnabledValue
}

// Int returns the serialized integer value for a visibility flag.
func (value Visibility) Int() int {
	return int(value)
}

// IsSupported reports whether value is a published visibility flag.
func (value Visibility) IsSupported() bool {
	return value == Hidden || value == Visible
}

// Int returns the serialized integer value for an installation flag.
func (value Installation) Int() int {
	return int(value)
}

// IsSupported reports whether value is a published installation flag.
func (value Installation) IsSupported() bool {
	return value == Uninstalled || value == Installed
}

// Int returns the serialized integer value for a read-state flag.
func (value ReadState) Int() int {
	return int(value)
}

// IsSupported reports whether value is a published read-state flag.
func (value ReadState) IsSupported() bool {
	return value == Unread || value == Read
}

// Int returns the serialized integer value for a yes/no flag.
func (value YesNo) Int() int {
	return int(value)
}

// IsSupported reports whether value is a published yes/no flag.
func (value YesNo) IsSupported() bool {
	return value == No || value == Yes
}
