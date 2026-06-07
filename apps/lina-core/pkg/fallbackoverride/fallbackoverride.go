// Package fallbackoverride defines the stable operation modes used when API
// rows inherit from platform defaults and may expose a fallback override action.
// The package is intentionally small so configuration and dictionary APIs can
// share the response contract without coupling to either module's service
// implementation or to plugin tenant capability packages.
package fallbackoverride

// Mode identifies the override operation available for one fallback row.
type Mode string

// Fallback override operation modes.
const (
	// None means no override action is available for the row.
	None Mode = "none"
	// CreateTenantOverride means the current tenant may create its own row that
	// overrides the inherited platform default.
	CreateTenantOverride Mode = "createTenantOverride"
)
