// This file contains explicit source-plugin runtime upgrade orchestration
// owned by the unified upgrade component.

package upgrade

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
)

// SourceUpgradeResult describes the outcome of one explicit source-plugin upgrade request.
type SourceUpgradeResult struct {
	// PluginID is the immutable plugin identifier.
	PluginID string
	// Name is the human-readable plugin display name.
	Name string
	// FromVersion is the effective version before the request ran.
	FromVersion string
	// ToVersion is the discovered version targeted by the request.
	ToVersion string
	// Executed reports whether upgrade work actually ran.
	Executed bool
	// Message explains the no-op or successful outcome in the effective locale.
	Message string
	// MessageKey is the runtime i18n key used to render Message.
	MessageKey string
	// MessageParams stores runtime i18n named parameters for MessageKey.
	MessageParams map[string]any
}

// UpgradeSourcePlugin applies one explicit source-plugin runtime upgrade and
// publishes source-plugin scoped cache changes.
func (s *serviceImpl) UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourceUpgradeResult, error) {
	result, err := s.ExecuteSourcePluginUpgrade(ctx, pluginID)
	if err != nil {
		if markErr := s.cachePublisher.PublishPluginChange(
			ctx,
			pluginID,
			plugintypes.TypeSource.String(),
			"source_plugin_upgrade_failed",
		); markErr != nil {
			logger.Warningf(
				ctx,
				"mark runtime cache changed after source upgrade failure failed plugin=%s err=%v",
				pluginID,
				markErr,
			)
		}
		return nil, err
	}
	if result != nil && result.Executed {
		if err = s.cachePublisher.SyncEnabledSnapshotAndPublishRuntimeChange(
			ctx,
			pluginID,
			"source_plugin_upgraded",
		); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// ExecuteSourcePluginUpgrade applies one source-plugin upgrade without running
// the public facade governance guard or publishing cache changes.
func (s *serviceImpl) ExecuteSourcePluginUpgrade(ctx context.Context, pluginID string) (*SourceUpgradeResult, error) {
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
	if candidate.status.Installed != plugintypes.InstalledYes {
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

	registry, err := s.storeSvc.SyncManifest(ctx, candidate.manifest)
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

	targetRelease, err := s.storeSvc.GetRelease(
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
	if err = s.validateUpgradeCandidateDependencies(ctx, candidate.manifest); err != nil {
		return nil, err
	}

	currentRelease, err := s.storeSvc.GetRegistryRelease(ctx, candidate.registry)
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

	if err = s.migrationSvc.ExecuteManifestSQLFiles(
		ctx,
		candidate.manifest,
		plugintypes.MigrationDirectionUpgrade,
	); err != nil {
		s.markSourcePluginUpgradeFailed(ctx, candidate.manifest, targetRelease, "upgrade-step-source-sql", err)
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

	updatedRegistry, err := s.storeSvc.GetRegistry(ctx, candidate.manifest.ID)
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
		if err = s.storeSvc.UpdateReleaseState(
			ctx,
			currentRelease.Id,
			plugintypes.ReleaseStatusInstalled,
			"",
		); err != nil {
			s.markSourcePluginUpgradeFailed(ctx, candidate.manifest, targetRelease, "previous-release-state", err)
			return nil, err
		}
	}
	if err = s.storeSvc.UpdateReleaseState(
		ctx,
		targetRelease.Id,
		plugintypes.BuildReleaseStatus(updatedRegistry.Installed, updatedRegistry.Status),
		s.storeSvc.BuildPackagePath(candidate.manifest),
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
		dependencyID, requiredVersion, currentVersion := plugindep.FirstBlockerFields(installResult.Blockers)
		return bizerr.NewCode(
			CodePluginDependencyBlocked,
			bizerr.P("pluginId", manifest.ID),
			bizerr.P("dependencyId", dependencyID),
			bizerr.P("requiredVersion", requiredVersion),
			bizerr.P("currentVersion", currentVersion),
			bizerr.P("chain", plugindep.FirstBlockerChain(installResult.Blockers)),
			bizerr.P("blockers", plugindep.FormatBlockers(installResult.Blockers)),
		)
	}

	if !s.dependencyTargetAlreadyInstalled(ctx, manifest.ID) {
		return nil
	}
	reverseResult, err := s.resolveReverseDependencies(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if plugindep.HasBlockers(reverseResult.Blockers) {
		dependents := plugindep.ToReverseDependentProjections(reverseResult.Dependents)
		dependencyID, requiredVersion, currentVersion := plugindep.FirstBlockerFields(reverseResult.Blockers)
		return bizerr.NewCode(
			CodePluginReverseDependencyBlocked,
			bizerr.P("pluginId", manifest.ID),
			bizerr.P("dependencyId", dependencyID),
			bizerr.P("requiredVersion", requiredVersion),
			bizerr.P("currentVersion", currentVersion),
			bizerr.P("dependents", strings.Join(plugindep.ReverseDependentIDs(dependents), ",")),
			bizerr.P("blockers", plugindep.FormatBlockers(reverseResult.Blockers)),
		)
	}
	return nil
}

// dependencyTargetAlreadyInstalled reports whether the target is already installed.
func (s *serviceImpl) dependencyTargetAlreadyInstalled(ctx context.Context, pluginID string) bool {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil || registry == nil {
		return false
	}
	return registry.Installed == plugintypes.InstalledYes
}

// applySourcePluginUpgradedRelease promotes the discovered release to the
// current effective registry version without changing installed/enabled flags.
func (s *serviceImpl) applySourcePluginUpgradedRelease(
	ctx context.Context,
	registry *store.PluginRecord,
	manifest *catalog.Manifest,
	release *store.ReleaseRecord,
) error {
	if registry == nil {
		return bizerr.NewCode(CodePluginSourceUpgradeRegistryRequired)
	}
	if manifest == nil {
		return bizerr.NewCode(CodePluginSourceUpgradeManifestRequired)
	}
	if release == nil {
		return bizerr.NewCode(CodePluginSourceUpgradeTargetReleaseRequired)
	}

	_, err := s.storeSvc.PromoteSourceRelease(ctx, registry, manifest, release)
	return err
}
