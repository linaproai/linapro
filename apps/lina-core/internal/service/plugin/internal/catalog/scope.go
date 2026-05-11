// This file defines plugin tenant-scope governance enums and normalization
// helpers used by plugin manifest validation and lifecycle policy checks.

package catalog

import "strings"

// ScopeNature defines how a plugin participates in tenant governance.
type ScopeNature string

// InstallMode defines how a tenant-aware plugin is enabled across tenants.
type InstallMode string

const (
	// ScopeNaturePlatformOnly marks a plugin as visible only to platform administrators.
	ScopeNaturePlatformOnly ScopeNature = "platform_only"
	// ScopeNatureTenantAware marks a plugin as available inside tenant contexts.
	ScopeNatureTenantAware ScopeNature = "tenant_aware"
)

const (
	// InstallModeGlobal enables one plugin globally after platform installation.
	InstallModeGlobal InstallMode = "global"
	// InstallModeTenantScoped lets each tenant enable the plugin independently.
	InstallModeTenantScoped InstallMode = "tenant_scoped"
)

// String returns the canonical scope-nature value.
func (value ScopeNature) String() string {
	return string(value)
}

// String returns the canonical install-mode value.
func (value InstallMode) String() string {
	return string(value)
}

// NormalizeScopeNature converts one raw manifest value to the canonical enum.
func NormalizeScopeNature(value string) ScopeNature {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ScopeNaturePlatformOnly.String():
		return ScopeNaturePlatformOnly
	default:
		return ScopeNatureTenantAware
	}
}

// NormalizeInstallMode converts one raw manifest value to the canonical enum.
func NormalizeInstallMode(value string) InstallMode {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case InstallModeTenantScoped.String():
		return InstallModeTenantScoped
	default:
		return InstallModeGlobal
	}
}

// IsSupportedScopeNature reports whether the raw value is one supported scope nature.
func IsSupportedScopeNature(value string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	return trimmed == ScopeNaturePlatformOnly.String() || trimmed == ScopeNatureTenantAware.String()
}

// IsSupportedInstallMode reports whether the raw value is one supported install mode.
func IsSupportedInstallMode(value string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	return trimmed == InstallModeGlobal.String() || trimmed == InstallModeTenantScoped.String()
}

