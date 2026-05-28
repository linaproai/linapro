// This file connects the side-effect-free dependency resolver to plugin
// lifecycle orchestration, API projections, and upgrade validation.

package plugin

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/management"
	"lina-core/pkg/bizerr"
)

type (
	// dependencyInstallContext records automatic install state for one request.
	dependencyInstallContext struct {
		// active marks target IDs already being installed in this request.
		active map[string]bool
	}

	// dependencySnapshotCache stores request-local dependency snapshots for
	// repeated read-only dependency checks during one plugin list projection.
	dependencySnapshotCache struct {
		snapshots []*plugindep.PluginSnapshot
	}
)

// dependencyInstallContextKey stores request-local dependency orchestration state.
type dependencyInstallContextKey struct{}

// dependencySnapshotCacheContextKey stores request-local dependency snapshots.
type dependencySnapshotCacheContextKey struct{}

// WithDependencySnapshotCache returns a child context that can reuse dependency
// snapshots across repeated read-only dependency checks in one request.
func (s *serviceImpl) WithDependencySnapshotCache(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if dependencySnapshotCacheFromContext(ctx) != nil {
		return ctx
	}
	return context.WithValue(ctx, dependencySnapshotCacheContextKey{}, &dependencySnapshotCache{})
}

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
	result := plugindep.ToCheckProjection(installResult)
	result.ReverseDependents = plugindep.ToReverseDependentProjections(reverseResult.Dependents)
	result.ReverseBlockers = plugindep.ToBlockerProjections(reverseResult.Blockers)
	return result, nil
}

// prepareInstallDependencies verifies a target before lifecycle side effects.
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
	result := plugindep.ToCheckProjection(check)
	if plugindep.HasBlockers(check.Blockers) {
		return result, nextCtx, s.buildDependencyBlockedError(normalizedID, check.Blockers)
	}
	return result, nextCtx, nil
}

// ensureNoReverseDependencies blocks uninstall when installed downstream plugins depend on target.
func (s *serviceImpl) ensureNoReverseDependencies(ctx context.Context, pluginID string) error {
	result, err := s.resolveReverseDependencies(ctx, pluginID, "")
	if err != nil {
		return err
	}
	if !plugindep.HasBlockers(result.Blockers) {
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
	if plugindep.HasBlockers(installResult.Blockers) {
		return s.buildDependencyBlockedError(manifest.ID, installResult.Blockers)
	}

	if !s.dependencyTargetAlreadyInstalled(ctx, manifest.ID) {
		return nil
	}
	reverseResult, err := s.resolveReverseDependencies(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if plugindep.HasBlockers(reverseResult.Blockers) {
		return s.buildReverseDependencyBlockedError(manifest.ID, reverseResult)
	}
	return nil
}

// resolveInstallDependencies evaluates dependency status for one discovered target.
func (s *serviceImpl) resolveInstallDependencies(ctx context.Context, pluginID string) (*plugindep.InstallCheckResult, error) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if manifest := management.ManifestByIDFromContext(ctx, normalizedPluginID); manifest != nil {
		return s.resolveInstallDependenciesForManifest(ctx, manifest)
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(normalizedPluginID)
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
	if candidate == nil {
		if cache := dependencySnapshotCacheFromContext(ctx); cache != nil && cache.snapshots != nil {
			return plugindep.ClonePluginSnapshots(cache.snapshots), nil
		}
	}
	manifests := management.ManifestSnapshotFromContext(ctx)
	if manifests == nil {
		var err error
		manifests, err = s.catalogSvc.ScanManifests()
		if err != nil {
			return nil, err
		}
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

	readCtx, err := s.catalogSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	registries, err := s.catalogSvc.ListAllRegistries(readCtx)
	if err != nil {
		return nil, err
	}
	candidateID := ""
	if candidate != nil {
		candidateID = strings.TrimSpace(candidate.ID)
	}
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		registryPluginID := strings.TrimSpace(registry.PluginId)
		if registryPluginID == "" {
			continue
		}
		snapshot := snapshotByID[registryPluginID]
		if snapshot == nil {
			if registry.ReleaseId <= 0 {
				continue
			}
			snapshot = &plugindep.PluginSnapshot{ID: registryPluginID}
			snapshotByID[registryPluginID] = snapshot
		}
		if registryPluginID == candidateID {
			snapshot.Installed = registry.Installed == catalog.InstalledYes
			continue
		}
		plugindep.ApplyRegistrySnapshot(readCtx, s.catalogSvc, snapshot, registry)
	}

	out := make([]*plugindep.PluginSnapshot, 0, len(snapshotByID))
	for _, snapshot := range snapshotByID {
		out = append(out, snapshot)
	}
	if candidate == nil {
		if cache := dependencySnapshotCacheFromContext(ctx); cache != nil {
			cache.snapshots = plugindep.ClonePluginSnapshots(out)
		}
	}
	return out, nil
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

// dependencySnapshotCacheFromContext returns the request-local dependency
// snapshot cache, if the current read path enabled one.
func dependencySnapshotCacheFromContext(ctx context.Context) *dependencySnapshotCache {
	if ctx == nil {
		return nil
	}
	value, ok := ctx.Value(dependencySnapshotCacheContextKey{}).(*dependencySnapshotCache)
	if !ok {
		return nil
	}
	return value
}

// dependencyTargetAlreadyInstalled reports whether the target is already installed.
func (s *serviceImpl) dependencyTargetAlreadyInstalled(ctx context.Context, pluginID string) bool {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil || registry == nil {
		return false
	}
	return registry.Installed == catalog.InstalledYes
}

// buildDependencyBlockedError converts resolver blockers into one structured business error.
func (s *serviceImpl) buildDependencyBlockedError(pluginID string, blockers []*plugindep.Blocker) error {
	dependencyID, requiredVersion, currentVersion := plugindep.FirstBlockerFields(blockers)
	return bizerr.NewCode(
		CodePluginDependencyBlocked,
		bizerr.P("pluginId", strings.TrimSpace(pluginID)),
		bizerr.P("dependencyId", dependencyID),
		bizerr.P("requiredVersion", requiredVersion),
		bizerr.P("currentVersion", currentVersion),
		bizerr.P("chain", plugindep.FirstBlockerChain(blockers)),
		bizerr.P("blockers", plugindep.FormatBlockers(blockers)),
	)
}

// buildReverseDependencyBlockedError converts reverse dependency blockers into one structured error.
func (s *serviceImpl) buildReverseDependencyBlockedError(
	pluginID string,
	result *plugindep.ReverseCheckResult,
) error {
	dependents := plugindep.ToReverseDependentProjections(result.Dependents)
	dependencyID, requiredVersion, currentVersion := plugindep.FirstBlockerFields(result.Blockers)
	return bizerr.NewCode(
		CodePluginReverseDependencyBlocked,
		bizerr.P("pluginId", strings.TrimSpace(pluginID)),
		bizerr.P("dependencyId", dependencyID),
		bizerr.P("requiredVersion", requiredVersion),
		bizerr.P("currentVersion", currentVersion),
		bizerr.P("dependents", strings.Join(plugindep.ReverseDependentIDs(dependents), ",")),
		bizerr.P("blockers", plugindep.FormatBlockers(result.Blockers)),
	)
}
