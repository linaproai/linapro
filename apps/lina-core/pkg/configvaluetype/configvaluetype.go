// Package configvaluetype defines the closed set of sys_config value input
// types used by management APIs and admin form rendering. Runtime host
// readers continue to consume the string value field only.
package configvaluetype

// Code identifies one persisted parameter value input type.
type Code string

// Closed value-type codes.
const (
	// Text is a single-line text input.
	Text Code = "text"
	// Textarea is a multi-line plain text input.
	Textarea Code = "textarea"
	// Number is a numeric input serialized as a decimal string.
	Number Code = "number"
	// Boolean is a true/false switch serialized as "true" or "false".
	Boolean Code = "boolean"
	// Select is a single-select dropdown driven by options metadata.
	Select Code = "select"
	// Radio is a single-choice radio group driven by options metadata.
	Radio Code = "radio"
	// MultiSelect is a multi-choice control; selected values join with ";".
	MultiSelect Code = "multi_select"
	// Richtext is rich text content; UIs may degrade to an enhanced textarea.
	Richtext Code = "richtext"
)

// String returns the canonical persisted code value.
func (code Code) String() string {
	return string(code)
}

// IsSupported reports whether code is a published value type.
func (code Code) IsSupported() bool {
	switch code {
	case Text, Textarea, Number, Boolean, Select, Radio, MultiSelect, Richtext:
		return true
	default:
		return false
	}
}

// RequiresOptions reports whether the type must persist a non-empty options list.
func (code Code) RequiresOptions() bool {
	switch code {
	case Select, Radio, MultiSelect:
		return true
	default:
		return false
	}
}

// Normalize returns a supported code, defaulting empty or unknown values to Text.
func Normalize(raw string) Code {
	code := Code(raw)
	if code.IsSupported() {
		return code
	}
	return Text
}

// All returns every published value type in stable order for documentation and tests.
func All() []Code {
	return []Code{Text, Textarea, Number, Boolean, Select, Radio, MultiSelect, Richtext}
}
