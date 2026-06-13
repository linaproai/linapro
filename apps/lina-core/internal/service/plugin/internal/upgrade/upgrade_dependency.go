// This file builds dependency projections used by upgrade preview and
// validation without depending on the root plugin facade internals.

package upgrade

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/plugintypes"
)

// resolveInstallDependenciesForManifest evaluates dependency status using a
// candidate manifest override.
func (s *serviceImpl) resolveInstallDependenciesForManifest(
	ctx context.Context,
	manifest *catalog.Manifest,
) (*plugindep.InstallCheckResult, error) {
	snapshots, err := s.buildDependencySnapshots(ctx, manifest)
	if err != nil {
		return nil, err
	}
	return s.dependencyResolver.CheckInstall(plugindep.InstallCheckInput{
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
	return s.dependencyResolver.CheckReverse(plugindep.ReverseCheckInput{
		TargetID:         strings.TrimSpace(pluginID),
		CandidateVersion: strings.TrimSpace(candidateVersion),
		Plugins:          snapshots,
	}), nil
}

// buildDependencySnapshots combines discovered manifests with installed
// registry release snapshots using one manifest scan and one batched registry read.
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
			Dependencies: plugintypes.CloneDependencySpec(manifest.Dependencies),
		}
	}
	if candidate != nil && strings.TrimSpace(candidate.ID) != "" {
		snapshotByID[candidate.ID] = &plugindep.PluginSnapshot{
			ID:           strings.TrimSpace(candidate.ID),
			Name:         strings.TrimSpace(candidate.Name),
			Version:      strings.TrimSpace(candidate.Version),
			Manifest:     candidate,
			Dependencies: plugintypes.CloneDependencySpec(candidate.Dependencies),
		}
	}

	readCtx, err := s.storeSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	registries, err := s.storeSvc.ListAllRegistries(readCtx)
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
			snapshot.Installed = registry.Installed == plugintypes.InstalledYes
			continue
		}
		plugindep.ApplyRegistrySnapshot(readCtx, s.storeSvc, snapshot, registry)
	}

	out := make([]*plugindep.PluginSnapshot, 0, len(snapshotByID))
	for _, snapshot := range snapshotByID {
		out = append(out, snapshot)
	}
	return plugindep.ClonePluginSnapshots(out), nil
}

// frameworkVersion returns the current LinaPro framework version authority.
func (s *serviceImpl) frameworkVersion(ctx context.Context) string {
	if s == nil || s.metadataSvc == nil {
		return ""
	}
	metadata := s.metadataSvc.GetMetadata(ctx)
	if metadata == nil {
		return ""
	}
	return strings.TrimSpace(metadata.Framework.Version)
}
