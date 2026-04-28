// Package contract defines the stable source-plugin upgrade governance
// contracts shared by host services and development tooling.
package contract

import "context"

const (
	// SourcePluginEnabledNo marks a source plugin as disabled in the upgrade snapshot.
	SourcePluginEnabledNo = 0
	// SourcePluginEnabledYes marks a source plugin as enabled in the upgrade snapshot.
	SourcePluginEnabledYes = 1
	// SourcePluginInstalledNo marks a source plugin as not installed in the upgrade snapshot.
	SourcePluginInstalledNo = 0
	// SourcePluginInstalledYes marks a source plugin as installed in the upgrade snapshot.
	SourcePluginInstalledYes = 1
)

// SourcePluginStatus describes one source plugin's effective version,
// discovered source version, and pending-upgrade state.
type SourcePluginStatus struct {
	// PluginID is the immutable plugin identifier.
	PluginID string
	// Name is the human-readable plugin display name.
	Name string
	// EffectiveVersion is the current effective version stored in sys_plugin.
	EffectiveVersion string
	// DiscoveredVersion is the version currently discovered from plugin.yaml.
	DiscoveredVersion string
	// Installed reports whether the plugin is already installed.
	Installed int
	// Enabled reports whether the plugin is currently enabled.
	Enabled int
	// NeedsUpgrade reports whether an installed plugin discovered a newer source version.
	NeedsUpgrade bool
	// DowngradeDetected reports whether the discovered source version is lower
	// than the current effective version, which is unsupported in this iteration.
	DowngradeDetected bool
}

// SourcePluginUpgradeResult describes the outcome of one explicit source-plugin upgrade request.
type SourcePluginUpgradeResult struct {
	// PluginID is the immutable plugin identifier.
	PluginID string
	// Name is the human-readable plugin display name.
	Name string
	// FromVersion is the effective version before the request ran.
	FromVersion string
	// ToVersion is the discovered version targeted by the request.
	ToVersion string
	// Executed reports whether upgrade work actually ran.
	Executed bool
	// Message explains the no-op or successful outcome in the effective locale.
	Message string
	// MessageKey is the runtime i18n key used to render Message.
	MessageKey string
	// MessageParams stores runtime i18n named parameters for MessageKey.
	MessageParams map[string]any
}

// Service defines source-plugin upgrade operations published to development tools.
type Service interface {
	// ListSourcePluginStatuses returns the current effective/discovered source-plugin version pairs.
	ListSourcePluginStatuses(ctx context.Context) ([]*SourcePluginStatus, error)
	// UpgradeSourcePlugin applies one explicit source-plugin upgrade.
	UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourcePluginUpgradeResult, error)
	// ValidateSourcePluginUpgradeReadiness fails fast when startup would hit pending source-plugin upgrades.
	ValidateSourcePluginUpgradeReadiness(ctx context.Context) error
}
