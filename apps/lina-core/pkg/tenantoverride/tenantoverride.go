// Package tenantoverride defines the stable operation modes used when tenant
// scoped records inherit from platform defaults. The package is intentionally
// small so configuration and dictionary APIs can share the contract without
// coupling to either module's service implementation.
package tenantoverride

import "strings"

// Mode identifies the override operation available for one fallback row.
type Mode string

// Tenant override operation modes.
const (
	// None means no override action is available for the row.
	None Mode = "none"
	// CreateTenantOverride means the current tenant may create its own row that
	// overrides the inherited platform default.
	CreateTenantOverride Mode = "createTenantOverride"
)

// String returns the canonical serialized value.
func (mode Mode) String() string {
	return string(mode)
}

// IsSupported reports whether mode is one of the published override modes.
func (mode Mode) IsSupported() bool {
	return mode == None || mode == CreateTenantOverride
}

// Normalize converts raw caller or storage input into one canonical mode.
func Normalize(value string) Mode {
	switch strings.TrimSpace(value) {
	case CreateTenantOverride.String():
		return CreateTenantOverride
	case None.String(), "":
		return None
	default:
		return ""
	}
}

// CanCreateTenantOverride reports whether mode allows creating a tenant-owned
// override row for an inherited platform default.
func CanCreateTenantOverride(mode Mode) bool {
	return mode == CreateTenantOverride
}
