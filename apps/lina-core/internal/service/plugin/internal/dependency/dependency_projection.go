// This file converts pure dependency resolver results into management-facing
// projection DTOs and deterministic diagnostics used by the root plugin facade.

package dependency

import "strings"

// FrameworkProjection exposes framework compatibility for management clients.
type FrameworkProjection struct {
	// RequiredVersion is the framework semantic-version range declared by the plugin.
	RequiredVersion string
	// CurrentVersion is the current LinaPro framework version.
	CurrentVersion string
	// Status is the compatibility state returned by the resolver.
	Status string
}

// PluginProjection exposes one plugin dependency edge.
type PluginProjection struct {
	// OwnerID is the plugin that declares the dependency.
	OwnerID string
	// DependencyID is the depended-on plugin identifier.
	DependencyID string
	// DependencyName is the depended-on plugin display name when known.
	DependencyName string
	// RequiredVersion is the declared dependency version range.
	RequiredVersion string
	// CurrentVersion is the discovered or installed dependency version.
	CurrentVersion string
	// Installed reports whether the dependency is already installed.
	Installed bool
	// Discovered reports whether the dependency is discoverable.
	Discovered bool
	// Status is the dependency state returned by the resolver.
	Status string
	// Chain is the dependency chain leading to this edge.
	Chain []string
}

// BlockerProjection exposes one hard dependency failure.
type BlockerProjection struct {
	// Code identifies the dependency failure category.
	Code string
	// PluginID is the plugin whose lifecycle is blocked.
	PluginID string
	// DependencyID is the dependency plugin when applicable.
	DependencyID string
	// RequiredVersion is the declared version range when applicable.
	RequiredVersion string
	// CurrentVersion is the observed version when applicable.
	CurrentVersion string
	// Chain is the dependency chain associated with the blocker.
	Chain []string
	// Detail is a concise operator diagnostic.
	Detail string
}

// ReverseDependentProjection exposes one installed downstream hard dependency.
type ReverseDependentProjection struct {
	// PluginID is the downstream plugin ID.
	PluginID string
	// Name is the downstream plugin display name.
	Name string
	// Version is the downstream plugin version.
	Version string
	// RequiredVersion is the target version range declared by the downstream plugin.
	RequiredVersion string
	// OwnerHostServices summarizes owner-aware host service declarations for the target owner.
	OwnerHostServices []*OwnerHostServiceProjection
}

// OwnerHostServiceProjection exposes one owner-aware host service summary.
type OwnerHostServiceProjection struct {
	// Owner is the owner plugin ID for the plugin-owned host service.
	Owner string
	// Service is the logical host service identifier.
	Service string
	// Version is the owner capability protocol version.
	Version string
	// Methods lists declared host service methods.
	Methods []string
}

// CheckProjection is the management-facing dependency status snapshot.
type CheckProjection struct {
	// TargetID is the checked plugin.
	TargetID string
	// Framework contains the framework compatibility result.
	Framework FrameworkProjection
	// Dependencies contains direct and transitive dependency edge checks.
	Dependencies []*PluginProjection
	// Blockers lists install-side hard failures.
	Blockers []*BlockerProjection
	// Cycle contains the first detected dependency cycle.
	Cycle []string
	// ReverseDependents lists installed downstream hard dependencies.
	ReverseDependents []*ReverseDependentProjection
	// ReverseBlockers lists uninstall or downstream-version blockers.
	ReverseBlockers []*BlockerProjection
}

// ToCheckProjection converts resolver install output into a management DTO.
func ToCheckProjection(result *InstallCheckResult) *CheckProjection {
	if result == nil {
		return &CheckProjection{}
	}
	return &CheckProjection{
		TargetID: strings.TrimSpace(result.TargetID),
		Framework: FrameworkProjection{
			RequiredVersion: result.Framework.RequiredVersion,
			CurrentVersion:  result.Framework.CurrentVersion,
			Status:          string(result.Framework.Status),
		},
		Dependencies: ToPluginProjections(result.Dependencies),
		Blockers:     ToBlockerProjections(result.Blockers),
		Cycle:        cloneStringSlice(result.Cycle),
	}
}

// ToPluginProjections converts resolver dependency edges.
func ToPluginProjections(items []*PluginDependencyCheck) []*PluginProjection {
	out := make([]*PluginProjection, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &PluginProjection{
			OwnerID:         item.OwnerID,
			DependencyID:    item.DependencyID,
			DependencyName:  item.DependencyName,
			RequiredVersion: item.RequiredVersion,
			CurrentVersion:  item.CurrentVersion,
			Installed:       item.Installed,
			Discovered:      item.Discovered,
			Status:          string(item.Status),
			Chain:           cloneStringSlice(item.Chain),
		})
	}
	return out
}

// ToBlockerProjections converts resolver blockers.
func ToBlockerProjections(items []*Blocker) []*BlockerProjection {
	out := make([]*BlockerProjection, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &BlockerProjection{
			Code:            string(item.Code),
			PluginID:        item.PluginID,
			DependencyID:    item.DependencyID,
			RequiredVersion: item.RequiredVersion,
			CurrentVersion:  item.CurrentVersion,
			Chain:           cloneStringSlice(item.Chain),
			Detail:          item.Detail,
		})
	}
	return out
}

// ToReverseDependentProjections converts resolver reverse-dependency results.
func ToReverseDependentProjections(items []*ReverseDependent) []*ReverseDependentProjection {
	out := make([]*ReverseDependentProjection, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &ReverseDependentProjection{
			PluginID:          item.PluginID,
			Name:              item.Name,
			Version:           item.Version,
			RequiredVersion:   item.RequiredVersion,
			OwnerHostServices: toOwnerHostServiceProjections(item.OwnerHostServices),
		})
	}
	return out
}

// CloneCheckProjection deep-copies a dependency projection attached to cached list rows.
func CloneCheckProjection(in *CheckProjection) *CheckProjection {
	if in == nil {
		return nil
	}
	out := *in
	out.Dependencies = clonePluginProjections(in.Dependencies)
	out.Blockers = cloneBlockerProjections(in.Blockers)
	out.Cycle = cloneStringSlice(in.Cycle)
	out.ReverseDependents = cloneReverseDependentProjections(in.ReverseDependents)
	out.ReverseBlockers = cloneBlockerProjections(in.ReverseBlockers)
	return &out
}

// HasBlockers reports whether resolver blockers contain any hard failure.
func HasBlockers(blockers []*Blocker) bool {
	return len(blockers) > 0
}

// ProjectionHasBlockers reports whether a management projection contains any hard failure.
func ProjectionHasBlockers(result *CheckProjection) bool {
	return result != nil && (len(result.Blockers) > 0 || len(result.ReverseBlockers) > 0)
}

// FormatBlockers renders a compact deterministic blocker summary for fallback messages.
func FormatBlockers(blockers []*Blocker) string {
	if len(blockers) == 0 {
		return ""
	}
	parts := make([]string, 0, len(blockers))
	for _, blocker := range blockers {
		if blocker == nil {
			continue
		}
		parts = append(parts, strings.Join([]string{
			string(blocker.Code),
			strings.TrimSpace(blocker.PluginID),
			strings.TrimSpace(blocker.DependencyID),
			strings.TrimSpace(blocker.RequiredVersion),
			strings.TrimSpace(blocker.CurrentVersion),
			strings.Join(blocker.Chain, ">"),
		}, "|"))
	}
	return strings.Join(parts, ";")
}

// FirstBlockerFields returns the first dependency/version tuple for error params.
func FirstBlockerFields(blockers []*Blocker) (string, string, string) {
	for _, blocker := range blockers {
		if blocker == nil {
			continue
		}
		return blocker.DependencyID, blocker.RequiredVersion, blocker.CurrentVersion
	}
	return "", "", ""
}

// FirstBlockerChain returns the first blocker chain for structured errors.
func FirstBlockerChain(blockers []*Blocker) string {
	for _, blocker := range blockers {
		if blocker == nil {
			continue
		}
		return strings.Join(blocker.Chain, ">")
	}
	return ""
}

// ReverseDependentIDs extracts downstream plugin IDs.
func ReverseDependentIDs(items []*ReverseDependentProjection) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil || strings.TrimSpace(item.PluginID) == "" {
			continue
		}
		out = append(out, strings.TrimSpace(item.PluginID))
	}
	return out
}

// FormatReverseDependentOwnerHostServices renders owner service method summaries
// for stable error parameters.
func FormatReverseDependentOwnerHostServices(items []*ReverseDependentProjection) string {
	if len(items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		for _, service := range item.OwnerHostServices {
			if service == nil {
				continue
			}
			parts = append(parts, strings.Join([]string{
				strings.TrimSpace(item.PluginID),
				strings.TrimSpace(service.Owner),
				strings.TrimSpace(service.Service),
				strings.TrimSpace(service.Version),
				strings.Join(cloneSortedStrings(service.Methods), ","),
			}, "|"))
		}
	}
	return strings.Join(parts, ";")
}

func toOwnerHostServiceProjections(items []*OwnerHostServiceSummary) []*OwnerHostServiceProjection {
	if len(items) == 0 {
		return nil
	}
	out := make([]*OwnerHostServiceProjection, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &OwnerHostServiceProjection{
			Owner:   item.Owner,
			Service: item.Service,
			Version: item.Version,
			Methods: cloneSortedStrings(item.Methods),
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// clonePluginProjections deep-copies dependency edge checks.
func clonePluginProjections(in []*PluginProjection) []*PluginProjection {
	if len(in) == 0 {
		return nil
	}
	out := make([]*PluginProjection, 0, len(in))
	for _, item := range in {
		if item == nil {
			continue
		}
		clone := *item
		clone.Chain = cloneStringSlice(item.Chain)
		out = append(out, &clone)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// cloneBlockerProjections deep-copies dependency blocker diagnostics.
func cloneBlockerProjections(in []*BlockerProjection) []*BlockerProjection {
	if len(in) == 0 {
		return nil
	}
	out := make([]*BlockerProjection, 0, len(in))
	for _, item := range in {
		if item == nil {
			continue
		}
		clone := *item
		clone.Chain = cloneStringSlice(item.Chain)
		out = append(out, &clone)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// cloneReverseDependentProjections deep-copies reverse dependency summaries.
func cloneReverseDependentProjections(in []*ReverseDependentProjection) []*ReverseDependentProjection {
	if len(in) == 0 {
		return nil
	}
	out := make([]*ReverseDependentProjection, 0, len(in))
	for _, item := range in {
		if item == nil {
			continue
		}
		clone := *item
		clone.OwnerHostServices = cloneOwnerHostServiceProjections(item.OwnerHostServices)
		out = append(out, &clone)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneOwnerHostServiceProjections(in []*OwnerHostServiceProjection) []*OwnerHostServiceProjection {
	if len(in) == 0 {
		return nil
	}
	out := make([]*OwnerHostServiceProjection, 0, len(in))
	for _, item := range in {
		if item == nil {
			continue
		}
		clone := *item
		clone.Methods = cloneSortedStrings(item.Methods)
		out = append(out, &clone)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// cloneStringSlice returns a copy of string values for DTO exposure.
func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}
