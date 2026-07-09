// This file centralizes plugin-controller flag conversions between public API
// DTO contracts and service-layer integer fields. Keep these helpers local to
// the controller so service internals do not import API-only packages.

package plugin

import "lina-core/pkg/statusflag"

// enabledPtrToInt converts an optional public enabled flag into a service
// filter pointer while preserving nil as an omitted filter.
func enabledPtrToInt(value *statusflag.Enabled) *int {
	if value == nil {
		return nil
	}
	converted := value.Int()
	return &converted
}

// installationPtrToInt converts an optional public installation flag into a
// service filter pointer while preserving nil as an omitted filter.
func installationPtrToInt(value *statusflag.Installation) *int {
	if value == nil {
		return nil
	}
	converted := value.Int()
	return &converted
}

// yesNoPtrToBool converts an optional public yes/no flag into the effective
// boolean policy, using fallback when callers omit the flag.
func yesNoPtrToBool(value *statusflag.YesNo, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value == statusflag.Yes
}

// boolToYesNo converts a boolean flag into the shared public yes/no flag.
func boolToYesNo(value bool) statusflag.YesNo {
	if value {
		return statusflag.Yes
	}
	return statusflag.No
}
