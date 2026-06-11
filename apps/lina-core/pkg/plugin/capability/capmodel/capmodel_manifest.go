// This file defines the neutral typed manifest snapshot primitive shared by
// source-plugin upgrade callbacks and dynamic-plugin lifecycle requests. The
// type is transport-agnostic; bridge wire naming and versioning are owned by
// pluginbridge/contract through type aliases.

package capmodel

// ManifestSnapshot is the typed manifest snapshot published to plugin
// lifecycle callbacks.
type ManifestSnapshot struct {
	// ID is the plugin identifier recorded in the manifest snapshot.
	ID string `json:"id"`
	// Name is the plugin display name recorded in the manifest snapshot.
	Name string `json:"name"`
	// Version is the plugin version recorded in the manifest snapshot.
	Version string `json:"version"`
	// Type is the plugin type recorded in the manifest snapshot.
	Type string `json:"type"`
	// ScopeNature is the plugin tenant-scope nature recorded in the manifest snapshot.
	ScopeNature string `json:"scopeNature"`
	// SupportsMultiTenant reports whether the plugin declares linapro-tenant-core support.
	SupportsMultiTenant bool `json:"supportsMultiTenant"`
	// DefaultInstallMode is the plugin default installation mode.
	DefaultInstallMode string `json:"defaultInstallMode"`
	// Description is the plugin description recorded in the manifest snapshot.
	Description string `json:"description"`
	// InstallSQLCount is the number of install SQL assets recorded in the snapshot.
	InstallSQLCount int `json:"installSqlCount"`
	// UninstallSQLCount is the number of uninstall SQL assets recorded in the snapshot.
	UninstallSQLCount int `json:"uninstallSqlCount"`
	// MockSQLCount is the number of mock SQL assets recorded in the snapshot.
	MockSQLCount int `json:"mockSqlCount"`
	// MenuCount is the number of menu definitions recorded in the snapshot.
	MenuCount int `json:"menuCount"`
	// BackendHookCount is the number of backend hook registrations recorded in the snapshot.
	BackendHookCount int `json:"backendHookCount"`
	// ResourceSpecCount is the number of resource specs recorded in the snapshot.
	ResourceSpecCount int `json:"resourceSpecCount"`
	// HostServiceAuthRequired reports whether host-service authorization is required.
	HostServiceAuthRequired bool `json:"hostServiceAuthRequired"`
}
