// menu_metadata.go declares host-owned menu mount metadata used to validate
// plugin menu parent relationships and stable top-level catalog bindings.
package menu

import "lina-core/pkg/orgcap"

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

// Official source-plugin identifiers bound to stable host menu catalogs.
const (
	// OrgCenter provides department and post management.
	OrgCenter = orgcap.ProviderPluginID
	// ContentNotice provides notice management.
	ContentNotice = "content-notice"
	// Media provides media strategy, binding, and stream alias management.
	Media = "media"
	// MonitorOnline provides online-user query and force-logout governance.
	MonitorOnline = "monitor-online"
	// MonitorServer provides server-monitor collection and query features.
	MonitorServer = "monitor-server"
	// MonitorOperLog provides operation-log persistence and management.
	MonitorOperLog = "monitor-operlog"
	// MonitorLoginLog provides login-log persistence and management.
	MonitorLoginLog = "monitor-loginlog"
	// MultiTenant provides tenant subject, membership, and lifecycle governance.
	MultiTenant = "multi-tenant"
)

var stableParentKeys = map[string][]string{
	OrgCenter:       {Org},
	ContentNotice:   {Content},
	Media:           {Content},
	MonitorOnline:   {Monitor},
	MonitorServer:   {Monitor},
	MonitorOperLog:  {Monitor},
	MonitorLoginLog: {Monitor},
	MultiTenant:     {Platform},
}

// ExpectedStableParentKey returns the primary host top-level parent key for
// one official source plugin. The second return value reports whether the
// plugin ID belongs to the published first-party plugin set.
func ExpectedStableParentKey(pluginID string) (string, bool) {
	parentKeys, ok := ExpectedStableParentKeys(pluginID)
	if !ok || len(parentKeys) == 0 {
		return "", ok
	}
	return parentKeys[0], true
}

// ExpectedStableParentKeys returns all allowed host top-level parent keys for
// one official source plugin. The returned slice is isolated from package state.
func ExpectedStableParentKeys(pluginID string) ([]string, bool) {
	parentKeys, ok := stableParentKeys[pluginID]
	if !ok || len(parentKeys) == 0 {
		return nil, ok
	}
	return append([]string(nil), parentKeys...), true
}

// IsExpectedStableParentKey reports whether parentKey is one allowed stable
// host catalog for the specified first-party source plugin.
func IsExpectedStableParentKey(pluginID string, parentKey string) (bool, bool) {
	parentKeys, ok := stableParentKeys[pluginID]
	if !ok {
		return false, false
	}
	for _, expected := range parentKeys {
		if expected == parentKey {
			return true, true
		}
	}
	return false, true
}
