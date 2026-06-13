// This file builds side-effect-free runtime upgrade previews from unified
// upgrade component dependencies.

package upgrade

import (
	"context"
	"errors"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
)

const (
	// RuntimeUpgradeRiskHintUpgradeSQLRequiresReview warns that upgrade SQL
	// should be reviewed before the user confirms runtime side effects.
	RuntimeUpgradeRiskHintUpgradeSQLRequiresReview = "plugin.runtimeUpgrade.risk.upgradeSqlRequiresReview"
	// RuntimeUpgradeRiskHintMockSQLExcluded warns that mock SQL is never loaded by upgrade.
	RuntimeUpgradeRiskHintMockSQLExcluded = "plugin.runtimeUpgrade.risk.mockSqlExcluded"
	// RuntimeUpgradeRiskHintHostServiceAuthorizationChanged warns that hostServices changed.
	RuntimeUpgradeRiskHintHostServiceAuthorizationChanged = "plugin.runtimeUpgrade.risk.hostServiceAuthorizationChanged"
	// RuntimeUpgradeRiskHintDependencyBlockers warns that dependency checks found hard blockers.
	RuntimeUpgradeRiskHintDependencyBlockers = "plugin.runtimeUpgrade.risk.dependencyBlockers"
)

// PreviewRuntimeUpgrade returns a side-effect-free upgrade preview for one
// plugin currently marked as pending or failed runtime upgrade.
func (s *serviceImpl) PreviewRuntimeUpgrade(ctx context.Context, pluginID string) (*RuntimeUpgradePreview, error) {
	if err := s.cacheFreshener.EnsureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil, bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", normalizedPluginID))
	}

	targetManifest, err := s.loadDesiredManifestForPreview(normalizedPluginID)
	if err != nil {
		return nil, err
	}
	registry, err := s.storeSvc.GetRegistry(ctx, normalizedPluginID)
	if err != nil {
		return nil, err
	}
	if registry == nil {
		return nil, bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", normalizedPluginID))
	}

	projection, err := s.storeSvc.BuildRuntimeUpgradeState(ctx, registry, targetManifest)
	if err != nil {
		return nil, err
	}
	if !CanExecute(projection.State) {
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradePreviewUnavailable,
			bizerr.P("pluginId", normalizedPluginID),
			bizerr.P("runtimeState", projection.State.String()),
		)
	}

	fromSnapshot, err := s.loadEffectiveManifestSnapshot(ctx, registry)
	if err != nil {
		return nil, err
	}
	toSnapshot, err := s.buildTargetManifestSnapshot(targetManifest)
	if err != nil {
		return nil, err
	}
	dependencyCheck, err := s.buildRuntimeUpgradeDependencyCheck(ctx, targetManifest)
	if err != nil {
		return nil, err
	}
	hostServicesDiff, err := buildRuntimeUpgradeHostServicesDiff(fromSnapshot, toSnapshot)
	if err != nil {
		return nil, err
	}
	sqlSummary := BuildSQLSummary(toSnapshot)

	return &RuntimeUpgradePreview{
		PluginID:          normalizedPluginID,
		RuntimeState:      RuntimeUpgradeState(projection.State),
		EffectiveVersion:  projection.EffectiveVersion,
		DiscoveredVersion: projection.DiscoveredVersion,
		FromManifest:      fromSnapshot,
		ToManifest:        toSnapshot,
		DependencyCheck:   dependencyCheck,
		SQLSummary:        sqlSummary,
		HostServicesDiff:  hostServicesDiff,
		RiskHints: buildRuntimeUpgradeRiskHints(
			sqlSummary,
			hostServicesDiff,
			plugindep.ProjectionHasBlockers(dependencyCheck),
		),
	}, nil
}

// loadDesiredManifestForPreview finds the currently discovered manifest while
// preserving catalog scan errors as diagnostics instead of converting every
// failure into not-found.
func (s *serviceImpl) loadDesiredManifestForPreview(pluginID string) (*catalog.Manifest, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}
	for _, manifest := range manifests {
		if manifest != nil && strings.TrimSpace(manifest.ID) == pluginID {
			return manifest, nil
		}
	}
	return nil, bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", pluginID))
}

// buildRuntimeUpgradeDependencyCheck evaluates target dependencies and
// downstream compatibility against the target upgrade version without side effects.
func (s *serviceImpl) buildRuntimeUpgradeDependencyCheck(
	ctx context.Context,
	targetManifest *catalog.Manifest,
) (*DependencyCheckResult, error) {
	if targetManifest == nil {
		return &DependencyCheckResult{}, nil
	}
	installResult, err := s.resolveInstallDependenciesForManifest(ctx, targetManifest)
	if err != nil {
		return nil, err
	}
	reverseResult, err := s.resolveReverseDependencies(ctx, targetManifest.ID, targetManifest.Version)
	if err != nil {
		return nil, err
	}
	result := plugindep.ToCheckProjection(installResult)
	result.ReverseDependents = plugindep.ToReverseDependentProjections(reverseResult.Dependents)
	result.ReverseBlockers = plugindep.ToBlockerProjections(reverseResult.Blockers)
	return result, nil
}

// loadEffectiveManifestSnapshot reads the manifest snapshot tied to the
// database-effective registry release.
func (s *serviceImpl) loadEffectiveManifestSnapshot(
	ctx context.Context,
	registry *store.PluginRecord,
) (*store.ManifestSnapshot, error) {
	release, err := s.storeSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return nil, err
	}
	if release == nil {
		pluginID := ""
		version := ""
		if registry != nil {
			pluginID = registry.PluginId
			version = registry.Version
		}
		return nil, bizerr.NewCode(
			CodePluginReleaseNotFound,
			bizerr.P("pluginId", pluginID),
			bizerr.P("version", version),
		)
	}
	snapshot, err := s.storeSvc.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		return nil, bizerr.WrapCode(
			err,
			CodePluginRuntimeUpgradeSnapshotInvalid,
			bizerr.P("pluginId", registry.PluginId),
			bizerr.P("version", registry.Version),
		)
	}
	if snapshot == nil {
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradeSnapshotMissing,
			bizerr.P("pluginId", registry.PluginId),
			bizerr.P("version", registry.Version),
		)
	}
	return snapshot, nil
}

// buildTargetManifestSnapshot converts the discovered target manifest into the
// same review snapshot model persisted for releases.
func (s *serviceImpl) buildTargetManifestSnapshot(
	manifest *catalog.Manifest,
) (*store.ManifestSnapshot, error) {
	snapshotYAML, err := s.storeSvc.BuildManifestSnapshot(manifest)
	if err != nil {
		return nil, bizerr.WrapCode(
			err,
			CodePluginRuntimeUpgradeSnapshotInvalid,
			bizerr.P("pluginId", manifestPluginID(manifest)),
			bizerr.P("version", manifestVersion(manifest)),
		)
	}
	snapshot, err := s.storeSvc.ParseManifestSnapshot(snapshotYAML)
	if err != nil {
		return nil, bizerr.WrapCode(
			err,
			CodePluginRuntimeUpgradeSnapshotInvalid,
			bizerr.P("pluginId", manifestPluginID(manifest)),
			bizerr.P("version", manifestVersion(manifest)),
		)
	}
	if snapshot == nil {
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradeSnapshotMissing,
			bizerr.P("pluginId", manifest.ID),
			bizerr.P("version", manifest.Version),
		)
	}
	return snapshot, nil
}

// manifestPluginID safely extracts the plugin ID for diagnostics.
func manifestPluginID(manifest *catalog.Manifest) string {
	if manifest == nil {
		return ""
	}
	return manifest.ID
}

// manifestVersion safely extracts the plugin version for diagnostics.
func manifestVersion(manifest *catalog.Manifest) string {
	if manifest == nil {
		return ""
	}
	return manifest.Version
}

// buildRuntimeUpgradeHostServicesDiff compares effective and target requested
// hostServices at service level.
func buildRuntimeUpgradeHostServicesDiff(
	fromSnapshot *store.ManifestSnapshot,
	toSnapshot *store.ManifestSnapshot,
) (RuntimeUpgradeHostServicesDiff, error) {
	diff, err := BuildHostServicesDiff(fromSnapshot, toSnapshot)
	if err != nil {
		var snapshotErr *SnapshotInvalidError
		if errors.As(err, &snapshotErr) {
			return RuntimeUpgradeHostServicesDiff{}, bizerr.WrapCode(
				err,
				CodePluginRuntimeUpgradeSnapshotInvalid,
				bizerr.P("pluginId", snapshotErr.PluginID),
				bizerr.P("version", snapshotErr.Version),
			)
		}
		return RuntimeUpgradeHostServicesDiff{}, err
	}
	return diff, nil
}

// buildRuntimeUpgradeRiskHints returns stable i18n keys for operator risk hints.
func buildRuntimeUpgradeRiskHints(
	sqlSummary RuntimeUpgradeSQLSummary,
	hostServicesDiff RuntimeUpgradeHostServicesDiff,
	dependencyBlocked bool,
) []string {
	return BuildRiskHints(sqlSummary, hostServicesDiff, dependencyBlocked, RiskHintKeys{
		UpgradeSQLRequiresReview:        RuntimeUpgradeRiskHintUpgradeSQLRequiresReview,
		MockSQLExcluded:                 RuntimeUpgradeRiskHintMockSQLExcluded,
		HostServiceAuthorizationChanged: RuntimeUpgradeRiskHintHostServiceAuthorizationChanged,
		DependencyBlockers:              RuntimeUpgradeRiskHintDependencyBlockers,
	})
}
