// This file defines plugin distribution governance values used to distinguish
// ordinary marketplace plugins from source plugins built into the host project.

package plugintypes

import "strings"

// PluginDistribution defines how the host governs plugin distribution lifecycle.
type PluginDistribution string

const (
	// DistributionMarketplace identifies a normal plugin managed through the plugin management entry.
	DistributionMarketplace PluginDistribution = "marketplace"
	// DistributionBuiltin identifies a source plugin that is part of the project distribution.
	DistributionBuiltin PluginDistribution = "builtin"
)

// String returns the canonical distribution value.
func (value PluginDistribution) String() string { return string(value) }

// NormalizeDistribution converts a raw manifest value to the canonical distribution value.
func NormalizeDistribution(value string) PluginDistribution {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "":
		return DistributionMarketplace
	case DistributionMarketplace.String():
		return DistributionMarketplace
	case DistributionBuiltin.String():
		return DistributionBuiltin
	default:
		return PluginDistribution(strings.TrimSpace(strings.ToLower(value)))
	}
}

// IsSupportedDistribution reports whether the distribution value is recognized.
func IsSupportedDistribution(value string) bool {
	distribution := NormalizeDistribution(value)
	return distribution == DistributionMarketplace || distribution == DistributionBuiltin
}
