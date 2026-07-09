// This file defines plugin installation and enablement status value objects
// plus helpers for deriving persisted release and menu key state.

package plugintypes

import (
	"strings"

	"lina-core/pkg/statusflag"
)

// Status reuses the shared API enabled flag type for plugin enablement.
type Status = statusflag.Enabled

// InstalledStatus reuses the shared API installation flag type.
type InstalledStatus = statusflag.Installation

// NormalizeStatus converts one raw database/entity integer into the typed
// plugin enablement enum used by state derivation helpers.
func NormalizeStatus(value int) Status {
	if value == statusflag.EnabledValue.Int() {
		return statusflag.EnabledValue
	}
	return statusflag.Disabled
}

// NormalizeInstalledStatus converts one raw database/entity integer into the
// typed plugin installation enum used by state derivation helpers.
func NormalizeInstalledStatus(value int) InstalledStatus {
	if value == statusflag.Installed.Int() {
		return statusflag.Installed
	}
	return statusflag.Uninstalled
}

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
	if NormalizeInstalledStatus(installed) != statusflag.Installed {
		return ReleaseStatusUninstalled
	}
	if NormalizeStatus(enabled) == statusflag.EnabledValue {
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
