// Package dependency resolves plugin dependency constraints without performing
// lifecycle side effects. It owns side-effect-free install, upgrade, uninstall,
// and management projection checks for the host plugin service.
package dependency

import (
	pluginv1 "lina-core/api/plugin/v1"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
)

// Resolver evaluates plugin dependency declarations against discovered and
// installed plugin snapshots.
type Resolver struct{}

// New creates a dependency resolver.
func New() *Resolver {
	return &Resolver{}
}

// BlockerCode reuses the plugin API dependency blocker enum.
type BlockerCode = pluginv1.BlockerCode

// DependencyStatus reuses the plugin API dependency edge enum.
type DependencyStatus = pluginv1.DependencyStatus

// FrameworkStatus reuses the plugin API framework compatibility enum.
type FrameworkStatus = pluginv1.FrameworkStatus

// PluginSnapshot describes one discovered or installed plugin state supplied to
// the pure dependency resolver.
type PluginSnapshot struct {
	// ID is the stable plugin identifier.
	ID string
	// Name is the display name used in dependency results.
	Name string
	// Version is the effective or discovered plugin version.
	Version string
	// Installed reports whether the plugin is installed in host governance.
	Installed bool
	// Manifest is the latest discovered manifest, if available.
	Manifest *catalog.Manifest
	// Dependencies is the dependency snapshot that should be used for this plugin.
	Dependencies *plugintypes.DependencySpec
	// DependencySnapshotUnknown conservatively blocks reverse checks when true.
	DependencySnapshotUnknown bool
	// OwnerHostServices lists owner-aware host service declarations from the
	// effective release snapshot or current discovered manifest.
	OwnerHostServices []*OwnerHostServiceSummary
}

// InstallCheckInput defines all state required to evaluate an install request.
type InstallCheckInput struct {
	// TargetID is the plugin being installed or upgraded.
	TargetID string
	// FrameworkVersion is the current LinaPro framework version.
	FrameworkVersion string
	// Plugins contains discovered and installed plugin snapshots.
	Plugins []*PluginSnapshot
}

// ReverseCheckInput defines all state required to evaluate uninstall or upgrade
// reverse-dependency protection.
type ReverseCheckInput struct {
	// TargetID is the plugin being uninstalled or upgraded.
	TargetID string
	// CandidateVersion is the target version after upgrade. Empty means uninstall.
	CandidateVersion string
	// Plugins contains installed plugin dependency snapshots.
	Plugins []*PluginSnapshot
	// ReverseIndex optionally supplies a prebuilt reverse-dependency index for Plugins.
	ReverseIndex *ReverseDependencyIndex
}

// InstallCheckResult is the side-effect-free dependency decision for one target.
type InstallCheckResult struct {
	// TargetID is the plugin being checked.
	TargetID string
	// Framework contains the target plugin framework compatibility result.
	Framework FrameworkCheck
	// Dependencies contains direct and transitive dependency edge checks.
	Dependencies []*PluginDependencyCheck
	// Blockers lists hard failures that must be resolved before lifecycle side effects.
	Blockers []*Blocker
	// Cycle contains the first detected hard-dependency cycle, if any.
	Cycle []string
}

// ReverseCheckResult is the side-effect-free reverse-dependency decision.
type ReverseCheckResult struct {
	// TargetID is the plugin being checked.
	TargetID string
	// CandidateVersion is the target version after upgrade. Empty means uninstall.
	CandidateVersion string
	// Dependents lists installed plugins depending on the target.
	Dependents []*ReverseDependent
	// Blockers lists hard failures that must be resolved before lifecycle side effects.
	Blockers []*Blocker
}

// FrameworkCheck describes one framework-version compatibility result.
type FrameworkCheck struct {
	// RequiredVersion is the declared semantic-version range.
	RequiredVersion string
	// CurrentVersion is the current LinaPro framework version.
	CurrentVersion string
	// Status is the compatibility state.
	Status FrameworkStatus
}

// PluginDependencyCheck describes one plugin-to-plugin dependency edge.
type PluginDependencyCheck struct {
	// OwnerID is the plugin declaring the dependency.
	OwnerID string
	// DependencyID is the depended-on plugin ID.
	DependencyID string
	// DependencyName is the display name when the dependency is known.
	DependencyName string
	// RequiredVersion is the declared dependency semantic-version range.
	RequiredVersion string
	// CurrentVersion is the installed or discovered dependency version.
	CurrentVersion string
	// Installed reports whether the dependency plugin is already installed.
	Installed bool
	// Discovered reports whether the dependency plugin manifest is available.
	Discovered bool
	// Status is the dependency edge state.
	Status DependencyStatus
	// Chain is the dependency chain leading to this edge.
	Chain []string
}

// ReverseDependent describes one installed downstream plugin depending on a target.
type ReverseDependent struct {
	// PluginID is the downstream plugin ID.
	PluginID string
	// Name is the downstream plugin display name.
	Name string
	// Version is the downstream plugin version.
	Version string
	// RequiredVersion is the target version range declared by the downstream plugin.
	RequiredVersion string
	// OwnerHostServices summarizes owner-aware host services that target this dependency.
	OwnerHostServices []*OwnerHostServiceSummary
}

// Blocker describes one hard dependency failure.
type Blocker struct {
	// Code identifies the failure category.
	Code BlockerCode
	// PluginID is the plugin whose lifecycle is blocked.
	PluginID string
	// DependencyID is the dependency plugin when applicable.
	DependencyID string
	// RequiredVersion is the declared version range when applicable.
	RequiredVersion string
	// CurrentVersion is the observed version when applicable.
	CurrentVersion string
	// Chain is the dependency chain associated with this blocker.
	Chain []string
	// Detail provides a concise developer/operator diagnostic.
	Detail string
}

// ReverseDependencyIndex stores installed downstream dependencies by target ID.
// It is built once from the current snapshot set so each reverse check can look
// up the requested target without scanning all plugin dependency declarations.
type ReverseDependencyIndex struct {
	entriesByTarget map[string][]*reverseDependencyEntry
	wildcardUnknown []*reverseDependencyEntry
}

// OwnerHostServiceSummary describes one owner-aware host service declaration
// relevant to reverse dependency diagnostics.
type OwnerHostServiceSummary struct {
	// Owner is the owner plugin ID for the plugin-owned host service.
	Owner string
	// Service is the logical host service identifier.
	Service string
	// Version is the owner capability protocol version.
	Version string
	// Methods lists declared host service methods.
	Methods []string
}
