// Package dependency resolves plugin dependency constraints without performing
// lifecycle side effects.
package dependency

import "lina-core/internal/service/plugin/internal/catalog"

// BlockerCode identifies one dependency check failure category.
type BlockerCode string

// DependencyStatus identifies one plugin dependency edge state.
type DependencyStatus string

// FrameworkStatus identifies the framework-version compatibility state.
type FrameworkStatus string

// Blocker codes returned by the dependency resolver.
const (
	// BlockerFrameworkVersionUnsatisfied reports a host framework-version mismatch.
	BlockerFrameworkVersionUnsatisfied BlockerCode = "framework_version_unsatisfied"
	// BlockerDependencyMissing reports a missing hard dependency.
	BlockerDependencyMissing BlockerCode = "dependency_missing"
	// BlockerDependencyVersionUnsatisfied reports a plugin dependency-version mismatch.
	BlockerDependencyVersionUnsatisfied BlockerCode = "dependency_version_unsatisfied"
	// BlockerDependencyManualInstallRequired reports a missing hard dependency that must be installed manually.
	BlockerDependencyManualInstallRequired BlockerCode = "dependency_manual_install_required"
	// BlockerDependencyCycle reports a hard-dependency cycle.
	BlockerDependencyCycle BlockerCode = "dependency_cycle"
	// BlockerDependencySnapshotUnknown reports an installed plugin whose dependency snapshot cannot be trusted.
	BlockerDependencySnapshotUnknown BlockerCode = "dependency_snapshot_unknown"
	// BlockerReverseDependency reports an installed downstream hard dependency.
	BlockerReverseDependency BlockerCode = "reverse_dependency"
	// BlockerReverseDependencyVersion reports an installed downstream hard dependency that rejects a candidate version.
	BlockerReverseDependencyVersion BlockerCode = "reverse_dependency_version"
)

// Dependency edge states returned by the resolver.
const (
	// DependencyStatusSatisfied reports that a dependency is installed and version-compatible.
	DependencyStatusSatisfied DependencyStatus = "satisfied"
	// DependencyStatusAutoInstallPlanned reports that a missing hard dependency can be automatically installed.
	DependencyStatusAutoInstallPlanned DependencyStatus = "auto_install_planned"
	// DependencyStatusManualInstallRequired reports that a missing hard dependency needs manual install first.
	DependencyStatusManualInstallRequired DependencyStatus = "manual_install_required"
	// DependencyStatusMissing reports that a dependency plugin cannot be found.
	DependencyStatusMissing DependencyStatus = "missing"
	// DependencyStatusVersionUnsatisfied reports that a dependency version is outside the requested range.
	DependencyStatusVersionUnsatisfied DependencyStatus = "version_unsatisfied"
	// DependencyStatusSoftUnsatisfied reports that a soft dependency is currently unavailable or incompatible.
	DependencyStatusSoftUnsatisfied DependencyStatus = "soft_unsatisfied"
)

// Framework compatibility states returned by the resolver.
const (
	// FrameworkStatusNotDeclared reports that no framework dependency was declared.
	FrameworkStatusNotDeclared FrameworkStatus = "not_declared"
	// FrameworkStatusSatisfied reports that the current framework version matches the range.
	FrameworkStatusSatisfied FrameworkStatus = "satisfied"
	// FrameworkStatusUnsatisfied reports that the current framework version does not match the range.
	FrameworkStatusUnsatisfied FrameworkStatus = "unsatisfied"
)

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
	Dependencies *catalog.DependencySpec
	// DependencySnapshotUnknown conservatively blocks reverse checks when true.
	DependencySnapshotUnknown bool
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
}

// InstallCheckResult is the side-effect-free dependency decision for one target.
type InstallCheckResult struct {
	// TargetID is the plugin being checked.
	TargetID string
	// Framework contains the target plugin framework compatibility result.
	Framework FrameworkCheck
	// Dependencies contains direct and transitive dependency edge checks.
	Dependencies []*PluginDependencyCheck
	// AutoInstallPlan lists missing hard dependencies that can be installed automatically.
	AutoInstallPlan []*AutoInstallPlanItem
	// ManualInstallRequired lists missing hard dependencies requiring manual install.
	ManualInstallRequired []*PluginDependencyCheck
	// SoftUnsatisfied lists optional dependencies that are missing or incompatible.
	SoftUnsatisfied []*PluginDependencyCheck
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
	// Required reports whether this dependency blocks lifecycle when unsatisfied.
	Required bool
	// InstallMode is the declared dependency install strategy.
	InstallMode catalog.DependencyInstallMode
	// Installed reports whether the dependency plugin is already installed.
	Installed bool
	// Discovered reports whether the dependency plugin manifest is available.
	Discovered bool
	// Status is the dependency edge state.
	Status DependencyStatus
	// Chain is the dependency chain leading to this edge.
	Chain []string
}

// AutoInstallPlanItem describes one planned automatic dependency installation.
type AutoInstallPlanItem struct {
	// PluginID is the dependency plugin to install.
	PluginID string
	// Name is the dependency display name when available.
	Name string
	// Version is the version that will be installed.
	Version string
	// RequiredBy is the direct parent plugin that requested this dependency.
	RequiredBy string
	// Chain is the dependency chain leading to this plan item.
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
