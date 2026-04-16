// This file defines plugin installation and enablement status constants and
// helpers for building composite release status strings.

package catalog

import (
	"strings"

	"lina-core/internal/model/entity"
)

const (
	// StatusDisabled marks a plugin as disabled (enabled=0 in DB).
	StatusDisabled = 0
	// StatusEnabled marks a plugin as enabled (enabled=1 in DB).
	StatusEnabled = 1
	// InstalledNo marks a plugin as not installed (installed=0 in DB).
	InstalledNo = 0
	// InstalledYes marks a plugin as installed (installed=1 in DB).
	InstalledYes = 1

	// MenuKeyPrefix is the common prefix for plugin-owned menu keys in sys_menu.menu_key.
	MenuKeyPrefix = "plugin:"
	// MenuRemarkPrefix is the legacy plugin marker prefix stored in sys_menu.remark.
	MenuRemarkPrefix = "plugin:"
	// DynamicRoutePermissionMenuKeySeparator marks synthetic route-permission menu keys.
	DynamicRoutePermissionMenuKeySeparator = ":perm:"
	// DynamicRoutePermissionMenuRemarkSuffix marks synthetic route-permission menu remarks.
	DynamicRoutePermissionMenuRemarkSuffix = ":dynamic-route-permission"
	// DynamicRoutePermissionMenuNamePrefix prefixes hidden route-permission menu names.
	DynamicRoutePermissionMenuNamePrefix = "动态路由权限:"
	// PluginStatusKeyPrefix is the stable status record key exposed to runtime consumers.
	PluginStatusKeyPrefix = "sys_plugin.status:"
	// PluginNodeStateMessageManifestSynchronized records a manifest-sync node-state update.
	PluginNodeStateMessageManifestSynchronized = "Source plugin manifest synchronized into host registry."
	// PluginNodeStateMessageStatusUpdated records a management-triggered status update.
	PluginNodeStateMessageStatusUpdated = "Plugin status updated from management API."
)

// BuildReleaseStatus builds the composite release status string from installation
// and enablement flags using the canonical format "<installed>_<enabled>".
func BuildReleaseStatus(installed int, enabled int) ReleaseStatus {
	if installed != InstalledYes {
		return ReleaseStatusUninstalled
	}
	if enabled == StatusEnabled {
		return ReleaseStatusActive
	}
	return ReleaseStatusInstalled
}

// ParsePluginIDFromMenu extracts the owning plugin ID from a menu row's key or remark field.
// It checks the menu key first and falls back to the remark field.
func ParsePluginIDFromMenu(menu *entity.SysMenu) string {
	if menu == nil {
		return ""
	}
	if pluginID := parsePluginIDFromMenuTagged(menu.MenuKey, MenuKeyPrefix); pluginID != "" {
		return pluginID
	}
	return parsePluginIDFromMenuTagged(menu.Remark, MenuRemarkPrefix)
}

// parsePluginIDFromMenuTagged extracts the plugin ID segment from a tagged value such as
// "plugin:<id>:rest" or "plugin:<id> rest" by trimming the prefix and stopping at ":" or " ".
func parsePluginIDFromMenuTagged(value string, prefix string) string {
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
