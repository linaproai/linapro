// This file connects the side-effect-free dependency resolver to plugin
// lifecycle orchestration, API projections, and upgrade validation.

package plugin

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/pkg/bizerr"
)

type (
	// DependencyFrameworkCheck exposes framework compatibility for management clients.
	DependencyFrameworkCheck struct {
		// RequiredVersion is the framework semantic-version range declared by the plugin.
		RequiredVersion string
		// CurrentVersion is the current LinaPro framework version.
		CurrentVersion string
		// Status is the compatibility state returned by the resolver.
		Status string
	}

	// DependencyPluginCheck exposes one plugin dependency edge.
	DependencyPluginCheck struct {
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
		// Required reports whether the dependency blocks lifecycle.
		Required bool
		// InstallMode is the declared dependency install mode.
		InstallMode string
		// Installed reports whether the dependency is already installed.
		Installed bool
		// Discovered reports whether the dependency is discoverable.
		Discovered bool
		// Status is the dependency state returned by the resolver.
		Status string
		// Chain is the dependency chain leading to this edge.
		Chain []string
	}

	// DependencyAutoInstallItem exposes one dependency install planned before the target.
	DependencyAutoInstallItem struct {
		// PluginID is the dependency plugin that will be installed automatically.
		PluginID string
		// Name is the dependency display name when available.
		Name string
		// Version is the dependency version to install.
		Version string
		// RequiredBy is the direct parent plugin declaring the dependency.
		RequiredBy string
		// Chain is the dependency chain leading to this plan item.
		Chain []string
	}

	// DependencyBlocker exposes one hard dependency failure.
	DependencyBlocker struct {
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

	// DependencyReverseDependent exposes one installed downstream hard dependency.
	DependencyReverseDependent struct {
		// PluginID is the downstream plugin ID.
		PluginID string
		// Name is the downstream plugin display name.
		Name string
		// Version is the downstream plugin version.
		Version string
		// RequiredVersion is the target version range declared by the downstream plugin.
		RequiredVersion string
	}

	// DependencyCheckResult is the management-facing dependency status snapshot.
	DependencyCheckResult struct {
		// TargetID is the checked plugin.
		TargetID string
		// Framework contains the framework compatibility result.
		Framework DependencyFrameworkCheck
		// Dependencies contains direct and transitive dependency edge checks.
		Dependencies []*DependencyPluginCheck
		// AutoInstallPlan lists missing hard dependencies that can be installed automatically.
		AutoInstallPlan []*DependencyAutoInstallItem
		// AutoInstalled lists dependencies already installed by the current request.
		AutoInstalled []*DependencyAutoInstallItem
		// ManualInstallRequired lists hard dependencies that need manual installation first.
		ManualInstallRequired []*DependencyPluginCheck
		// SoftUnsatisfied lists optional dependencies that are unavailable or incompatible.
		SoftUnsatisfied []*DependencyPluginCheck
		// Blockers lists install-side hard failures.
		Blockers []*DependencyBlocker
		// Cycle contains the first detected dependency cycle.
		Cycle []string
		// ReverseDependents lists installed downstream hard dependencies.
		ReverseDependents []*DependencyReverseDependent
		// ReverseBlockers lists uninstall or downstream-version blockers.
		ReverseBlockers []*DependencyBlocker
	}

	// dependencyInstallContext records automatic install state for one request.
	dependencyInstallContext struct {
		// active marks target IDs already being installed in this request.
		active map[string]bool
		// skipAutoPlan disables recursive auto-install planning for plan items.
		skipAutoPlan bool
	}
)

// dependencyInstallContextKey stores request-local dependency orchestration state.
type dependencyInstallContextKey struct{}

// CheckPluginDependencies evaluates install and uninstall dependency status for one plugin.
func (s *serviceImpl) CheckPluginDependencies(ctx context.Context, pluginID string) (*DependencyCheckResult, error) {
	installResult, err := s.resolveInstallDependencies(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	reverseResult, err := s.resolveReverseDependencies(ctx, pluginID, "")
	if err != nil {
		return nil, err
	}
	result := toDependencyCheckResult(installResult)
	result.ReverseDependents = toDependencyReverseDependents(reverseResult.Dependents)
	result.ReverseBlockers = toDependencyBlockers(reverseResult.Blockers)
	return result, nil
}

// prepareInstallDependencies verifies a target and installs automatic dependencies first.
func (s *serviceImpl) prepareInstallDependencies(
	ctx context.Context,
	pluginID string,
	options InstallOptions,
) (*DependencyCheckResult, context.Context, error) {
	depCtx := dependencyContextFrom(ctx)
	normalizedID := strings.TrimSpace(pluginID)
	if normalizedID == "" {
		return nil, ctx, nil
	}
	if depCtx.active[normalizedID] {
		return nil, ctx, nil
	}

	depCtx.active[normalizedID] = true
	defer delete(depCtx.active, normalizedID)

	nextCtx := context.WithValue(ctx, dependencyInstallContextKey{}, depCtx)
	check, err := s.resolveInstallDependencies(nextCtx, normalizedID)
	if err != nil {
		return nil, nextCtx, err
	}
	result := toDependencyCheckResult(check)
	if hasDependencyBlockers(check.Blockers) {
		return result, nextCtx, s.buildDependencyBlockedError(normalizedID, check.Blockers)
	}
	if s.dependencyTargetAlreadyInstalled(ctx, normalizedID) && len(check.AutoInstallPlan) > 0 {
		blockers := blockersFromAutoInstallPlan(normalizedID, check.AutoInstallPlan)
		result.Blockers = append(result.Blockers, toDependencyBlockers(blockers)...)
		return result, nextCtx, s.buildDependencyBlockedError(normalizedID, blockers)
	}
	if depCtx.skipAutoPlan {
		return result, nextCtx, nil
	}
	installed := make([]*DependencyAutoInstallItem, 0, len(check.AutoInstallPlan))
	for _, item := range check.AutoInstallPlan {
		if item == nil || strings.TrimSpace(item.PluginID) == "" {
			continue
		}
		installCtx := dependencyContextFrom(nextCtx)
		installCtx.skipAutoPlan = true
		itemCtx := context.WithValue(nextCtx, dependencyInstallContextKey{}, installCtx)
		if _, err = s.Install(itemCtx, item.PluginID, dependencyInstallOptions(options)); err != nil {
			return result, nextCtx, s.buildDependencyAutoInstallFailedError(normalizedID, item.PluginID, installed, err)
		}
		installed = append(installed, toDependencyAutoInstallItem(item))
	}
	result.AutoInstalled = installed
	return result, nextCtx, nil
}

// ensureNoReverseDependencies blocks uninstall when installed downstream plugins depend on target.
func (s *serviceImpl) ensureNoReverseDependencies(ctx context.Context, pluginID string) error {
	result, err := s.resolveReverseDependencies(ctx, pluginID, "")
	if err != nil {
		return err
	}
	if !hasDependencyBlockers(result.Blockers) {
		return nil
	}
	return s.buildReverseDependencyBlockedError(pluginID, result)
}

// ValidateSourcePluginUpgradeCandidate validates a source upgrade target before side effects.
func (s *serviceImpl) ValidateSourcePluginUpgradeCandidate(ctx context.Context, manifest *catalog.Manifest) error {
	return s.validateUpgradeCandidateDependencies(ctx, manifest)
}

// ValidateDynamicPluginCandidate validates a dynamic release candidate before side effects.
func (s *serviceImpl) ValidateDynamicPluginCandidate(ctx context.Context, manifest *catalog.Manifest) error {
	return s.validateUpgradeCandidateDependencies(ctx, manifest)
}

// validateUpgradeCandidateDependencies checks candidate dependencies and downstream version safety.
func (s *serviceImpl) validateUpgradeCandidateDependencies(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return nil
	}
	installResult, err := s.resolveInstallDependenciesForManifest(ctx, manifest)
	if err != nil {
		return err
	}
	if hasDependencyBlockers(installResult.Blockers) {
		return s.buildDependencyBlockedError(manifest.ID, installResult.Blockers)
	}
	if len(installResult.AutoInstallPlan) > 0 {
		blockers := blockersFromAutoInstallPlan(manifest.ID, installResult.AutoInstallPlan)
		return s.buildDependencyBlockedError(manifest.ID, blockers)
	}

	reverseResult, err := s.resolveReverseDependencies(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if hasDependencyBlockers(reverseResult.Blockers) {
		return s.buildReverseDependencyBlockedError(manifest.ID, reverseResult)
	}
	return nil
}

// resolveInstallDependencies evaluates dependency status for one discovered target.
func (s *serviceImpl) resolveInstallDependencies(ctx context.Context, pluginID string) (*plugindep.InstallCheckResult, error) {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return nil, err
	}
	return s.resolveInstallDependenciesForManifest(ctx, manifest)
}

// resolveInstallDependenciesForManifest evaluates dependency status using a candidate manifest override.
func (s *serviceImpl) resolveInstallDependenciesForManifest(
	ctx context.Context,
	manifest *catalog.Manifest,
) (*plugindep.InstallCheckResult, error) {
	snapshots, err := s.buildDependencySnapshots(ctx, manifest)
	if err != nil {
		return nil, err
	}
	resolver := plugindep.New()
	return resolver.CheckInstall(plugindep.InstallCheckInput{
		TargetID:         strings.TrimSpace(manifest.ID),
		FrameworkVersion: s.frameworkVersion(ctx),
		Plugins:          snapshots,
	}), nil
}

// resolveReverseDependencies evaluates installed downstream dependencies for one target.
func (s *serviceImpl) resolveReverseDependencies(
	ctx context.Context,
	pluginID string,
	candidateVersion string,
) (*plugindep.ReverseCheckResult, error) {
	snapshots, err := s.buildDependencySnapshots(ctx, nil)
	if err != nil {
		return nil, err
	}
	resolver := plugindep.New()
	return resolver.CheckReverse(plugindep.ReverseCheckInput{
		TargetID:         strings.TrimSpace(pluginID),
		CandidateVersion: strings.TrimSpace(candidateVersion),
		Plugins:          snapshots,
	}), nil
}

// buildDependencySnapshots combines discovered manifests with installed registry release snapshots.
func (s *serviceImpl) buildDependencySnapshots(
	ctx context.Context,
	candidate *catalog.Manifest,
) ([]*plugindep.PluginSnapshot, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}
	snapshotByID := make(map[string]*plugindep.PluginSnapshot, len(manifests)+1)
	for _, manifest := range manifests {
		if manifest == nil || strings.TrimSpace(manifest.ID) == "" {
			continue
		}
		snapshotByID[manifest.ID] = &plugindep.PluginSnapshot{
			ID:           strings.TrimSpace(manifest.ID),
			Name:         strings.TrimSpace(manifest.Name),
			Version:      strings.TrimSpace(manifest.Version),
			Manifest:     manifest,
			Dependencies: catalog.CloneDependencySpec(manifest.Dependencies),
		}
	}
	if candidate != nil && strings.TrimSpace(candidate.ID) != "" {
		snapshotByID[candidate.ID] = &plugindep.PluginSnapshot{
			ID:           strings.TrimSpace(candidate.ID),
			Name:         strings.TrimSpace(candidate.Name),
			Version:      strings.TrimSpace(candidate.Version),
			Manifest:     candidate,
			Dependencies: catalog.CloneDependencySpec(candidate.Dependencies),
		}
	}

	registries, err := s.catalogSvc.ListAllRegistries(ctx)
	if err != nil {
		return nil, err
	}
	candidateID := ""
	if candidate != nil {
		candidateID = strings.TrimSpace(candidate.ID)
	}
	for _, registry := range registries {
		if registry == nil || strings.TrimSpace(registry.PluginId) == "" {
			continue
		}
		snapshot := snapshotByID[registry.PluginId]
		if snapshot == nil {
			snapshot = &plugindep.PluginSnapshot{ID: strings.TrimSpace(registry.PluginId)}
			snapshotByID[registry.PluginId] = snapshot
		}
		if strings.TrimSpace(registry.PluginId) == candidateID {
			snapshot.Installed = registry.Installed == catalog.InstalledYes
			continue
		}
		applyRegistryDependencySnapshot(ctx, s.catalogSvc, snapshot, registry)
	}

	out := make([]*plugindep.PluginSnapshot, 0, len(snapshotByID))
	for _, snapshot := range snapshotByID {
		out = append(out, snapshot)
	}
	return out, nil
}

// applyRegistryDependencySnapshot prefers installed release snapshots for effective dependency metadata.
func applyRegistryDependencySnapshot(
	ctx context.Context,
	catalogSvc catalog.Service,
	snapshot *plugindep.PluginSnapshot,
	registry *entity.SysPlugin,
) {
	if snapshot == nil || registry == nil {
		return
	}
	if strings.TrimSpace(registry.Name) != "" {
		snapshot.Name = strings.TrimSpace(registry.Name)
	}
	if strings.TrimSpace(registry.Version) != "" {
		snapshot.Version = strings.TrimSpace(registry.Version)
	}
	snapshot.Installed = registry.Installed == catalog.InstalledYes
	if !snapshot.Installed {
		return
	}
	release, err := catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil || release == nil {
		snapshot.DependencySnapshotUnknown = true
		return
	}
	releaseSnapshot, err := catalogSvc.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil || releaseSnapshot == nil {
		snapshot.DependencySnapshotUnknown = true
		return
	}
	if strings.TrimSpace(releaseSnapshot.Name) != "" {
		snapshot.Name = strings.TrimSpace(releaseSnapshot.Name)
	}
	if strings.TrimSpace(releaseSnapshot.Version) != "" {
		snapshot.Version = strings.TrimSpace(releaseSnapshot.Version)
	}
	snapshot.Dependencies = catalog.CloneDependencySpec(releaseSnapshot.Dependencies)
}

// frameworkVersion returns the current LinaPro framework version authority.
func (s *serviceImpl) frameworkVersion(ctx context.Context) string {
	if s == nil || s.configSvc == nil {
		return ""
	}
	metadata := s.configSvc.GetMetadata(ctx)
	if metadata == nil {
		return ""
	}
	return strings.TrimSpace(metadata.Framework.Version)
}

// dependencyContextFrom returns one request-local dependency orchestration context.
func dependencyContextFrom(ctx context.Context) *dependencyInstallContext {
	if ctx != nil {
		if value, ok := ctx.Value(dependencyInstallContextKey{}).(*dependencyInstallContext); ok && value != nil {
			if value.active == nil {
				value.active = make(map[string]bool)
			}
			return value
		}
	}
	return &dependencyInstallContext{active: make(map[string]bool)}
}

// dependencyInstallOptions limits operator-only install decorations to the original target plugin.
func dependencyInstallOptions(options InstallOptions) InstallOptions {
	return InstallOptions{}
}

// dependencyTargetAlreadyInstalled reports whether the target is already installed.
func (s *serviceImpl) dependencyTargetAlreadyInstalled(ctx context.Context, pluginID string) bool {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil || registry == nil {
		return false
	}
	return registry.Installed == catalog.InstalledYes
}

// blockersFromAutoInstallPlan converts upgrade-time auto plans into hard blockers.
func blockersFromAutoInstallPlan(pluginID string, plan []*plugindep.AutoInstallPlanItem) []*plugindep.Blocker {
	blockers := make([]*plugindep.Blocker, 0, len(plan))
	for _, item := range plan {
		if item == nil {
			continue
		}
		blockers = append(blockers, &plugindep.Blocker{
			Code:           plugindep.BlockerDependencyManualInstallRequired,
			PluginID:       strings.TrimSpace(pluginID),
			DependencyID:   strings.TrimSpace(item.PluginID),
			CurrentVersion: strings.TrimSpace(item.Version),
			Chain:          cloneStringSlice(item.Chain),
			Detail:         "dependency must be installed before upgrading or refreshing the target plugin",
		})
	}
	return blockers
}

// hasDependencyBlockers reports whether resolver blockers contain any hard failure.
func hasDependencyBlockers(blockers []*plugindep.Blocker) bool {
	return len(blockers) > 0
}

// buildDependencyBlockedError converts resolver blockers into one structured business error.
func (s *serviceImpl) buildDependencyBlockedError(pluginID string, blockers []*plugindep.Blocker) error {
	dependencyID, requiredVersion, currentVersion := firstDependencyBlockerFields(blockers)
	return bizerr.NewCode(
		CodePluginDependencyBlocked,
		bizerr.P("pluginId", strings.TrimSpace(pluginID)),
		bizerr.P("dependencyId", dependencyID),
		bizerr.P("requiredVersion", requiredVersion),
		bizerr.P("currentVersion", currentVersion),
		bizerr.P("chain", firstDependencyBlockerChain(blockers)),
		bizerr.P("blockers", formatDependencyBlockers(blockers)),
	)
}

// buildDependencyAutoInstallFailedError returns a structured error for partial auto-install failure.
func (s *serviceImpl) buildDependencyAutoInstallFailedError(
	pluginID string,
	failedPluginID string,
	installed []*DependencyAutoInstallItem,
	cause error,
) error {
	return bizerr.WrapCode(
		cause,
		CodePluginDependencyAutoInstallFailed,
		bizerr.P("pluginId", strings.TrimSpace(pluginID)),
		bizerr.P("failedPluginId", strings.TrimSpace(failedPluginID)),
		bizerr.P("installedDependencies", strings.Join(dependencyAutoInstalledIDs(installed), ",")),
	)
}

// buildReverseDependencyBlockedError converts reverse dependency blockers into one structured error.
func (s *serviceImpl) buildReverseDependencyBlockedError(
	pluginID string,
	result *plugindep.ReverseCheckResult,
) error {
	dependents := toDependencyReverseDependents(result.Dependents)
	dependencyID, requiredVersion, currentVersion := firstDependencyBlockerFields(result.Blockers)
	return bizerr.NewCode(
		CodePluginReverseDependencyBlocked,
		bizerr.P("pluginId", strings.TrimSpace(pluginID)),
		bizerr.P("dependencyId", dependencyID),
		bizerr.P("requiredVersion", requiredVersion),
		bizerr.P("currentVersion", currentVersion),
		bizerr.P("dependents", strings.Join(reverseDependentIDs(dependents), ",")),
		bizerr.P("blockers", formatDependencyBlockers(result.Blockers)),
	)
}

// toDependencyCheckResult converts resolver install output into a service DTO.
func toDependencyCheckResult(result *plugindep.InstallCheckResult) *DependencyCheckResult {
	if result == nil {
		return &DependencyCheckResult{}
	}
	return &DependencyCheckResult{
		TargetID: strings.TrimSpace(result.TargetID),
		Framework: DependencyFrameworkCheck{
			RequiredVersion: result.Framework.RequiredVersion,
			CurrentVersion:  result.Framework.CurrentVersion,
			Status:          string(result.Framework.Status),
		},
		Dependencies:          toDependencyPluginChecks(result.Dependencies),
		AutoInstallPlan:       toDependencyAutoInstallItems(result.AutoInstallPlan),
		ManualInstallRequired: toDependencyPluginChecks(result.ManualInstallRequired),
		SoftUnsatisfied:       toDependencyPluginChecks(result.SoftUnsatisfied),
		Blockers:              toDependencyBlockers(result.Blockers),
		Cycle:                 cloneStringSlice(result.Cycle),
	}
}

// toDependencyPluginChecks converts resolver dependency edges.
func toDependencyPluginChecks(items []*plugindep.PluginDependencyCheck) []*DependencyPluginCheck {
	out := make([]*DependencyPluginCheck, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &DependencyPluginCheck{
			OwnerID:         item.OwnerID,
			DependencyID:    item.DependencyID,
			DependencyName:  item.DependencyName,
			RequiredVersion: item.RequiredVersion,
			CurrentVersion:  item.CurrentVersion,
			Required:        item.Required,
			InstallMode:     item.InstallMode.String(),
			Installed:       item.Installed,
			Discovered:      item.Discovered,
			Status:          string(item.Status),
			Chain:           cloneStringSlice(item.Chain),
		})
	}
	return out
}

// toDependencyAutoInstallItems converts resolver auto-install plan items.
func toDependencyAutoInstallItems(items []*plugindep.AutoInstallPlanItem) []*DependencyAutoInstallItem {
	out := make([]*DependencyAutoInstallItem, 0, len(items))
	for _, item := range items {
		converted := toDependencyAutoInstallItem(item)
		if converted != nil {
			out = append(out, converted)
		}
	}
	return out
}

// toDependencyAutoInstallItem converts one resolver auto-install plan item.
func toDependencyAutoInstallItem(item *plugindep.AutoInstallPlanItem) *DependencyAutoInstallItem {
	if item == nil {
		return nil
	}
	return &DependencyAutoInstallItem{
		PluginID:   item.PluginID,
		Name:       item.Name,
		Version:    item.Version,
		RequiredBy: item.RequiredBy,
		Chain:      cloneStringSlice(item.Chain),
	}
}

// toDependencyBlockers converts resolver blockers.
func toDependencyBlockers(items []*plugindep.Blocker) []*DependencyBlocker {
	out := make([]*DependencyBlocker, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &DependencyBlocker{
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

// toDependencyReverseDependents converts resolver reverse-dependency results.
func toDependencyReverseDependents(items []*plugindep.ReverseDependent) []*DependencyReverseDependent {
	out := make([]*DependencyReverseDependent, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &DependencyReverseDependent{
			PluginID:        item.PluginID,
			Name:            item.Name,
			Version:         item.Version,
			RequiredVersion: item.RequiredVersion,
		})
	}
	return out
}

// formatDependencyBlockers renders a compact deterministic blocker summary for fallback messages.
func formatDependencyBlockers(blockers []*plugindep.Blocker) string {
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

// firstDependencyBlockerFields returns the first dependency/version tuple for error params.
func firstDependencyBlockerFields(blockers []*plugindep.Blocker) (string, string, string) {
	for _, blocker := range blockers {
		if blocker == nil {
			continue
		}
		return blocker.DependencyID, blocker.RequiredVersion, blocker.CurrentVersion
	}
	return "", "", ""
}

// firstDependencyBlockerChain returns the first blocker chain for structured errors.
func firstDependencyBlockerChain(blockers []*plugindep.Blocker) string {
	for _, blocker := range blockers {
		if blocker == nil {
			continue
		}
		return strings.Join(blocker.Chain, ">")
	}
	return ""
}

// dependencyAutoInstalledIDs extracts dependency IDs from installed plan items.
func dependencyAutoInstalledIDs(items []*DependencyAutoInstallItem) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil || strings.TrimSpace(item.PluginID) == "" {
			continue
		}
		out = append(out, strings.TrimSpace(item.PluginID))
	}
	return out
}

// reverseDependentIDs extracts downstream plugin IDs.
func reverseDependentIDs(items []*DependencyReverseDependent) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil || strings.TrimSpace(item.PluginID) == "" {
			continue
		}
		out = append(out, strings.TrimSpace(item.PluginID))
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
