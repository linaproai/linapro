// This file defines plugin enum aliases and normalization helpers used by
// manifest validation, lifecycle policy checks, and governance projections.

package plugintypes

import (
	"strings"

	pluginv1 "lina-core/api/plugin/v1"
)

// PluginType reuses the API plugin type enum for normalization helpers.
type PluginType = pluginv1.PluginType

// PluginDistribution reuses the API plugin distribution enum.
type PluginDistribution = pluginv1.PluginDistribution

// ScopeNature reuses the API plugin scope nature enum.
type ScopeNature = pluginv1.ScopeNature

// InstallMode reuses the API plugin install mode enum.
type InstallMode = pluginv1.InstallMode

// NormalizeType converts a raw type string to the canonical PluginType.
func NormalizeType(value string) PluginType {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case pluginv1.PluginTypeSource.String():
		return pluginv1.PluginTypeSource
	case pluginv1.PluginTypeDynamic.String():
		return pluginv1.PluginTypeDynamic
	default:
		return pluginv1.PluginType(strings.TrimSpace(strings.ToLower(value)))
	}
}

// IsSupportedType reports whether the given type string is a recognized plugin type.
func IsSupportedType(value string) bool {
	pluginType := NormalizeType(value)
	return pluginType == pluginv1.PluginTypeSource || pluginType == pluginv1.PluginTypeDynamic
}

// NormalizeDistribution converts a raw manifest value to the canonical distribution value.
func NormalizeDistribution(value string) PluginDistribution {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "":
		return pluginv1.PluginDistributionMarketplace
	case pluginv1.PluginDistributionMarketplace.String():
		return pluginv1.PluginDistributionMarketplace
	case pluginv1.PluginDistributionBuiltin.String():
		return pluginv1.PluginDistributionBuiltin
	default:
		return pluginv1.PluginDistribution(strings.TrimSpace(strings.ToLower(value)))
	}
}

// IsSupportedDistribution reports whether the distribution value is recognized.
func IsSupportedDistribution(value string) bool {
	distribution := NormalizeDistribution(value)
	return distribution == pluginv1.PluginDistributionMarketplace || distribution == pluginv1.PluginDistributionBuiltin
}

// NormalizeScopeNature converts one raw manifest value to the canonical enum.
func NormalizeScopeNature(value string) ScopeNature {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case pluginv1.ScopeNaturePlatformOnly.String():
		return pluginv1.ScopeNaturePlatformOnly
	default:
		return pluginv1.ScopeNatureTenantAware
	}
}

// NormalizeInstallMode converts one raw manifest value to the canonical enum.
func NormalizeInstallMode(value string) InstallMode {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case pluginv1.InstallModeTenantScoped.String():
		return pluginv1.InstallModeTenantScoped
	default:
		return pluginv1.InstallModeGlobal
	}
}

// IsSupportedScopeNature reports whether the raw value is one supported scope nature.
func IsSupportedScopeNature(value string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	return trimmed == pluginv1.ScopeNaturePlatformOnly.String() || trimmed == pluginv1.ScopeNatureTenantAware.String()
}

// IsSupportedInstallMode reports whether the raw value is one supported install mode.
func IsSupportedInstallMode(value string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	return trimmed == pluginv1.InstallModeGlobal.String() || trimmed == pluginv1.InstallModeTenantScoped.String()
}
