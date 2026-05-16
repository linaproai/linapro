// This file contains explicit source-plugin runtime upgrade execution
// orchestration.

package sourceupgrade

import (
	"context"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

// UpgradeSourcePlugin applies one explicit source-plugin upgrade from the
// current effective version to the newer discovered source version.
func (s *serviceImpl) UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourceUpgradeResult, error) {
	candidate, err := s.findSourceUpgradeCandidate(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if candidate == nil || candidate.manifest == nil || candidate.status == nil {
		return nil, bizerr.NewCode(CodePluginSourceUpgradeCandidateNotFound, bizerr.P("pluginId", pluginID))
	}

	result := &SourceUpgradeResult{
		PluginID:    candidate.status.PluginID,
		Name:        candidate.status.Name,
		FromVersion: candidate.status.EffectiveVersion,
		ToVersion:   candidate.status.DiscoveredVersion,
	}
	if candidate.status.Installed != catalog.InstalledYes {
		setSourceUpgradeResultMessage(
			ctx,
			s.i18nSvc,
			result,
			sourceUpgradeNotInstalledSkippedKey,
			"Source plugin is not installed. Upgrade skipped.",
			nil,
		)
		return result, nil
	}

	registry, err := s.catalogSvc.SyncManifest(ctx, candidate.manifest)
	if err != nil {
		return nil, err
	}
	candidate.registry = registry
	candidate.status, err = buildSourceUpgradeStatus(candidate.manifest, registry)
	if err != nil {
		return nil, err
	}
	result.FromVersion = candidate.status.EffectiveVersion
	result.ToVersion = candidate.status.DiscoveredVersion

	versionCompare, err := compareSourceUpgradeVersions(
		candidate.status.EffectiveVersion,
		candidate.status.DiscoveredVersion,
	)
	if err != nil {
		return nil, err
	}
	if versionCompare == 0 {
		setSourceUpgradeResultMessage(
			ctx,
			s.i18nSvc,
			result,
			sourceUpgradeAlreadyLatestKey,
			"The current source plugin is already up to date. No upgrade is required.",
			nil,
		)
		return result, nil
	}
	if versionCompare > 0 {
		return nil, bizerr.NewCode(
			CodePluginSourceUpgradeDowngradeUnsupported,
			bizerr.P("pluginId", candidate.status.PluginID),
			bizerr.P("effectiveVersion", candidate.status.EffectiveVersion),
			bizerr.P("discoveredVersion", candidate.status.DiscoveredVersion),
		)
	}

	targetRelease, err := s.catalogSvc.GetRelease(
		ctx,
		candidate.manifest.ID,
		candidate.manifest.Version,
	)
	if err != nil {
		return nil, err
	}
	if targetRelease == nil {
		return nil, bizerr.NewCode(
			CodePluginSourceUpgradeTargetReleaseNotFound,
			bizerr.P("pluginId", candidate.manifest.ID),
			bizerr.P("version", candidate.manifest.Version),
		)
	}
	if err = s.validateCandidateDependencies(ctx, candidate.manifest); err != nil {
		return nil, err
	}

	currentRelease, err := s.catalogSvc.GetRegistryRelease(ctx, candidate.registry)
	if err != nil {
		return nil, err
	}
	plan, err := s.buildSourceUpgradePlan(
		currentRelease,
		targetRelease,
		candidate.status.EffectiveVersion,
		candidate.status.DiscoveredVersion,
	)
	if err != nil {
		return nil, err
	}
	if err = s.executeBeforeSourceUpgrade(ctx, candidate.manifest, plan); err != nil {
		s.markSourcePluginUpgradeFailed(ctx, candidate.manifest, targetRelease, "before-upgrade", err)
		return nil, err
	}
	if err = s.executeSourceUpgradeCallback(ctx, candidate.manifest, plan); err != nil {
		s.markSourcePluginUpgradeFailed(ctx, candidate.manifest, targetRelease, "upgrade-callback", err)
		return nil, err
	}

	if err = s.lifecycleSvc.ExecuteManifestSQLFiles(
		ctx,
		candidate.manifest,
		catalog.MigrationDirectionUpgrade,
	); err != nil {
		s.markSourcePluginReleaseFailed(ctx, candidate.manifest, targetRelease)
		return nil, err
	}
	if err = s.integrationSvc.SyncPluginMenusAndPermissions(ctx, candidate.manifest); err != nil {
		s.markSourcePluginUpgradeFailed(ctx, candidate.manifest, targetRelease, "governance-menu", err)
		return nil, err
	}
	if err = s.integrationSvc.SyncPluginResourceReferences(ctx, candidate.manifest); err != nil {
		s.markSourcePluginUpgradeFailed(ctx, candidate.manifest, targetRelease, "governance-resource-ref", err)
		return nil, err
	}
	if err = s.applySourcePluginUpgradedRelease(ctx, candidate.registry, candidate.manifest, targetRelease); err != nil {
		s.markSourcePluginUpgradeFailed(ctx, candidate.manifest, targetRelease, "release-switch", err)
		return nil, err
	}

	updatedRegistry, err := s.catalogSvc.GetRegistry(ctx, candidate.manifest.ID)
	if err != nil {
		s.markSourcePluginUpgradeFailed(ctx, candidate.manifest, targetRelease, "registry-refresh", err)
		return nil, err
	}
	if updatedRegistry == nil {
		err = bizerr.NewCode(
			CodePluginSourceUpgradeRegistryAfterUpgradeNotFound,
			bizerr.P("pluginId", candidate.manifest.ID),
		)
		s.markSourcePluginUpgradeFailed(ctx, candidate.manifest, targetRelease, "registry-refresh", err)
		return nil, err
	}

	if currentRelease != nil && currentRelease.Id > 0 && currentRelease.Id != targetRelease.Id {
		if err = s.catalogSvc.UpdateReleaseState(
			ctx,
			currentRelease.Id,
			catalog.ReleaseStatusInstalled,
			"",
		); err != nil {
			s.markSourcePluginUpgradeFailed(ctx, candidate.manifest, targetRelease, "previous-release-state", err)
			return nil, err
		}
	}
	if err = s.catalogSvc.UpdateReleaseState(
		ctx,
		targetRelease.Id,
		catalog.BuildReleaseStatus(updatedRegistry.Installed, updatedRegistry.Status),
		s.catalogSvc.BuildPackagePath(candidate.manifest),
	); err != nil {
		s.markSourcePluginUpgradeFailed(ctx, candidate.manifest, targetRelease, "target-release-state", err)
		return nil, err
	}
	if err = s.runtimeSvc.SyncPluginNodeState(
		ctx,
		updatedRegistry.PluginId,
		updatedRegistry.Version,
		updatedRegistry.Installed,
		updatedRegistry.Status,
		"Source plugin runtime upgrade completed.",
	); err != nil {
		s.markSourcePluginUpgradeFailed(ctx, candidate.manifest, targetRelease, "node-state", err)
		return nil, err
	}
	if err = s.executeAfterSourceUpgrade(ctx, candidate.manifest, plan); err != nil {
		logger.Warningf(ctx, "source plugin after-upgrade callback failed plugin=%s err=%v", candidate.manifest.ID, err)
	}
	if err = s.integrationSvc.DispatchPluginHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginUpgraded,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: candidate.manifest.ID,
			Name:     candidate.manifest.Name,
			Version:  candidate.manifest.Version,
			Status:   &updatedRegistry.Status,
		}),
	); err != nil {
		logger.Warningf(ctx, "source plugin upgraded hook dispatch failed plugin=%s err=%v", candidate.manifest.ID, err)
	}

	result.Executed = true
	setSourceUpgradeResultMessage(
		ctx,
		s.i18nSvc,
		result,
		sourceUpgradeSuccessKey,
		"Source plugin upgraded from {fromVersion} to {toVersion}.",
		map[string]any{
			"fromVersion": candidate.status.EffectiveVersion,
			"toVersion":   candidate.status.DiscoveredVersion,
		},
	)
	return result, nil
}

// validateCandidateDependencies delegates dependency checks through an explicit
// optional seam so sourceupgrade stays decoupled from the root plugin facade.
func (s *serviceImpl) validateCandidateDependencies(ctx context.Context, manifest *catalog.Manifest) error {
	if s.dependencyValidator == nil {
		return nil
	}
	return s.dependencyValidator.ValidateSourcePluginUpgradeCandidate(ctx, manifest)
}
