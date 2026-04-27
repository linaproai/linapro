// This file defines plugin installation and enablement status constants and
// helpers for building composite release status strings.

package catalog

import (
	"strings"

	"lina-core/internal/model/entity"
)

// Status defines the typed plugin enablement enum used by status derivation logic.
type Status int

// InstalledStatus defines the typed plugin installation enum used by status derivation logic.
type InstalledStatus int

// Typed plugin status enums used by catalog-level state derivation helpers.
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
// plugin enablement enum used by catalog helpers.
func NormalizeStatus(value int) Status {
	if value == PluginStatusEnabled.Int() {
		return PluginStatusEnabled
	}
	return PluginStatusDisabled
}

// NormalizeInstalledStatus converts one raw database/entity integer into the
// typed plugin installation enum used by catalog helpers.
func NormalizeInstalledStatus(value int) InstalledStatus {
	if value == PluginInstalledYes.Int() {
		return PluginInstalledYes
	}
	return PluginInstalledNo
}

// Database projection constants shared by DAO/entity integration points that
// still persist raw integer fields.
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

// Stable string markers and message templates shared by registry, menu, and
// runtime projections. These values are identifiers or messages rather than
// enum-style state sets, so they remain ordinary string constants.
const (
	// MenuKeyPrefix is the common prefix for plugin-owned menu keys in sys_menu.menu_key.
	MenuKeyPrefix = "plugin:"
	// MenuRemarkPrefix is the legacy plugin marker prefix stored in sys_menu.remark.
	MenuRemarkPrefix = "plugin:"
	// DynamicRoutePermissionMenuKeySeparator marks synthetic route-permission menu keys.
	DynamicRoutePermissionMenuKeySeparator = ":perm:"
	// DynamicRoutePermissionMenuRemarkSuffix marks synthetic route-permission menu remarks.
	DynamicRoutePermissionMenuRemarkSuffix = ":dynamic-route-permission"
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
// and enablement flags using the canonical format "<installed>_<enabled>".
func BuildReleaseStatus(installed int, enabled int) ReleaseStatus {
	if NormalizeInstalledStatus(installed) != PluginInstalledYes {
		return ReleaseStatusUninstalled
	}
	if NormalizeStatus(enabled) == PluginStatusEnabled {
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
