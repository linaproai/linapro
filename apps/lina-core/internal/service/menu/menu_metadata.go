// menu_metadata.go declares host-owned menu mount metadata used to validate
// plugin menu parent relationships and stable top-level catalog bindings. The
// constants here are stable contract keys consumed by plugin registration and
// must remain independent of localized display names.

package menu

// Stable host catalog menu keys.
const (
	// Dashboard is the stable key of the workspace catalog.
	Dashboard = "dashboard"
	// IAM is the stable key of the identity-and-access catalog.
	IAM = "iam"
	// Org is the stable key of the organization catalog.
	Org = "org"
	// Setting is the stable key of the system-settings catalog.
	Setting = "setting"
	// Content is the stable key of the content catalog.
	Content = "content"
	// Monitor is the stable key of the monitoring catalog.
	Monitor = "monitor"
	// Scheduler is the stable key of the scheduled-job catalog.
	Scheduler = "scheduler"
	// Extension is the stable key of the extension-governance catalog.
	Extension = "extension"
	// Platform is the stable key of the platform-administration catalog.
	Platform = "platform"
	// Developer is the stable key of the developer-support catalog.
	Developer = "developer"
)

var stableCatalogKeys = map[string]struct{}{
	Dashboard: {},
	IAM:       {},
	Org:       {},
	Setting:   {},
	Content:   {},
	Monitor:   {},
	Scheduler: {},
	Extension: {},
	Platform:  {},
	Developer: {},
}

// IsStableCatalogKey reports whether the given menu key belongs to one
// host-owned top-level catalog.
func IsStableCatalogKey(menuKey string) bool {
	_, ok := stableCatalogKeys[menuKey]
	return ok
}
