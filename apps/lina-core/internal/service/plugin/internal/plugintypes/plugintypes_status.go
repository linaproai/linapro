// This file defines plugin installation and enablement status value objects
// plus helpers for deriving persisted release and menu key state.

package plugintypes

import "strings"

// Status defines the typed plugin enablement enum used by state derivation logic.
type Status int

// InstalledStatus defines the typed plugin installation enum used by state derivation logic.
type InstalledStatus int

const (
	// PluginStatusDisabled means the plugin is currently disabled.
	PluginStatusDisabled Status = 0
	// PluginStatusEnabled means the plugin is currently enabled.
	PluginStatusEnabled Status = 1

	// PluginInstalledNo means the plugin is currently not installed.
	PluginInstalledNo InstalledStatus = 0
	// PluginInstalledYes means the plugin is currently installed.
	PluginInstalledYes InstalledStatus = 1
)

// Int returns the database-compatible integer code for one plugin enablement status.
func (value Status) Int() int {
	return int(value)
}

// Int returns the database-compatible integer code for one plugin installation status.
func (value InstalledStatus) Int() int {
	return int(value)
}

// NormalizeStatus converts one raw database/entity integer into the typed
// plugin enablement enum used by state derivation helpers.
func NormalizeStatus(value int) Status {
	if value == PluginStatusEnabled.Int() {
		return PluginStatusEnabled
	}
	return PluginStatusDisabled
}

// NormalizeInstalledStatus converts one raw database/entity integer into the
// typed plugin installation enum used by state derivation helpers.
func NormalizeInstalledStatus(value int) InstalledStatus {
	if value == PluginInstalledYes.Int() {
		return PluginInstalledYes
	}
	return PluginInstalledNo
}

const (
	// StatusDisabled marks a plugin as disabled (enabled=0 in DB).
	StatusDisabled = 0
	// StatusEnabled marks a plugin as enabled (enabled=1 in DB).
	StatusEnabled = 1
	// InstalledNo marks a plugin as not installed (installed=0 in DB).
	InstalledNo = 0
	// InstalledYes marks a plugin as installed (installed=1 in DB).
	InstalledYes = 1
)

const (
	// MenuKeyPrefix is the common prefix for plugin-owned menu keys in sys_menu.menu_key.
	MenuKeyPrefix = "plugin:"
	// DynamicRoutePermissionMenuKeySeparator marks synthetic route-permission menu keys.
	DynamicRoutePermissionMenuKeySeparator = ":perm:"
	// DynamicRoutePermissionMenuNamePrefix prefixes hidden route-permission menu names.
	DynamicRoutePermissionMenuNamePrefix = "Dynamic Route Permission:"
	// PluginStatusKeyPrefix is the stable status record key exposed to runtime consumers.
	PluginStatusKeyPrefix = "sys_plugin.status:"
	// PluginNodeStateMessageManifestSynchronized records a manifest-sync node-state update.
	PluginNodeStateMessageManifestSynchronized = "Source plugin manifest synchronized into host registry."
	// PluginNodeStateMessageStatusUpdated records a management-triggered status update.
	PluginNodeStateMessageStatusUpdated = "Plugin status updated from management API."
)

// BuildReleaseStatus builds the composite release status string from installation
// and enablement flags.
func BuildReleaseStatus(installed int, enabled int) ReleaseStatus {
	if NormalizeInstalledStatus(installed) != PluginInstalledYes {
		return ReleaseStatusUninstalled
	}
	if NormalizeStatus(enabled) == PluginStatusEnabled {
		return ReleaseStatusActive
	}
	return ReleaseStatusInstalled
}

// ParsePluginIDFromMenuKey extracts the owning plugin ID from a menu key.
func ParsePluginIDFromMenuKey(menuKey string) string {
	return parsePluginIDFromTaggedValue(menuKey, MenuKeyPrefix)
}

// parsePluginIDFromTaggedValue extracts the plugin ID segment from a tagged
// value such as "plugin:<id>:rest" or "plugin:<id> rest".
func parsePluginIDFromTaggedValue(value string, prefix string) string {
	tagged := strings.TrimSpace(value)
	if !strings.HasPrefix(tagged, prefix) {
		return ""
	}
	suffix := tagged[len(prefix):]
	end := len(suffix)
	for _, sep := range []string{":", " "} {
		if idx := strings.Index(suffix, sep); idx >= 0 && idx < end {
			end = idx
		}
	}
	return suffix[:end]
}
