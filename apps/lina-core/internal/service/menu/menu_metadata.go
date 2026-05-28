// menu_metadata.go declares host-owned menu mount metadata used to validate
// plugin menu parent relationships and stable top-level catalog bindings. The
// constants here are stable contract keys consumed by plugin registration and
// must remain independent of localized display names.

package menu

import "strings"

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

// AuthProvider is the stable key of the host-owned authentication-provider
// catalog. Unlike stable catalogs, it is not seeded unconditionally; the host
// materializes it on demand when a third-party login plugin mounts under it and
// removes it once the last such plugin is uninstalled.
const AuthProvider = "auth-provider"

// ManagedCatalog describes one host-owned top-level catalog that the host
// materializes on demand when a plugin mounts under it and removes once it has
// no remaining child menu. The host owns the catalog's display metadata; plugins
// only reference the key through their manifest parent_key.
type ManagedCatalog struct {
	// Key is the stable menu key of the catalog.
	Key string
	// Name is the source-language display name used as the i18n fallback.
	Name string
	// Icon is the catalog icon identifier.
	Icon string
	// Sort is the top-level display order.
	Sort int
	// Remark is the host-owned administrative note stored on the catalog row.
	Remark string
}

// managedCatalogs holds host-owned on-demand top-level catalogs keyed by menu key.
var managedCatalogs = map[string]ManagedCatalog{
	AuthProvider: {
		Key:    AuthProvider,
		Name:   "Authentication Providers",
		Icon:   "lucide:key-round",
		Sort:   11,
		Remark: "宿主托管目录：第三方授权登录管理",
	},
}

// LookupManagedCatalog returns the host-owned on-demand catalog definition for a
// menu key. The second result reports whether the key names a managed catalog.
func LookupManagedCatalog(menuKey string) (ManagedCatalog, bool) {
	catalog, ok := managedCatalogs[strings.TrimSpace(menuKey)]
	return catalog, ok
}

// IsManagedCatalogKey reports whether the given menu key names a host-owned
// on-demand catalog the host may materialize and remove based on plugin demand.
func IsManagedCatalogKey(menuKey string) bool {
	_, ok := managedCatalogs[strings.TrimSpace(menuKey)]
	return ok
}
