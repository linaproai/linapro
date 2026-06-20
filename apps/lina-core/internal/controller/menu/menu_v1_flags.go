// This file centralizes conversions between typed API flag contracts and the
// menu service's integer flag input model.

package menu

import (
	"lina-core/pkg/statusflag"
)

// enabledPtrToInt converts an optional enabled flag into the service integer
// pointer expected by menu filters and patch inputs.
func enabledPtrToInt(value *statusflag.Enabled) *int {
	if value == nil {
		return nil
	}
	converted := value.Int()
	return &converted
}

// visibilityPtrToInt converts an optional visibility flag into the service
// integer pointer expected by menu filters and patch inputs.
func visibilityPtrToInt(value *statusflag.Visibility) *int {
	if value == nil {
		return nil
	}
	converted := value.Int()
	return &converted
}

// yesNoPtrToInt converts an optional yes/no flag into the service integer
// pointer expected by menu patch inputs.
func yesNoPtrToInt(value *statusflag.YesNo) *int {
	if value == nil {
		return nil
	}
	converted := value.Int()
	return &converted
}

// statusflagEnabled converts a service integer flag into an API enabled flag.
func statusflagEnabled(value int) statusflag.Enabled {
	return statusflag.Enabled(value)
}

// statusflagVisibility converts a service integer flag into an API visibility flag.
func statusflagVisibility(value int) statusflag.Visibility {
	return statusflag.Visibility(value)
}

// statusflagYesNo converts a service integer flag into an API yes/no flag.
func statusflagYesNo(value int) statusflag.YesNo {
	return statusflag.YesNo(value)
}
