// This file computes runtime-upgrade state from discovered manifest metadata
// and effective registry/release rows without mutating governance tables.

package store

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
)

const (
	// RuntimeUpgradeFailureCodeReleaseFailed is the stable diagnostic code for a failed target release.
	RuntimeUpgradeFailureCodeReleaseFailed = "plugin_upgrade_release_failed"
	// RuntimeUpgradeFailureMessageKeyReleaseFailed is the i18n key for failed target releases.
	RuntimeUpgradeFailureMessageKeyReleaseFailed = "plugin.runtimeUpgrade.failure.releaseFailed"
	// RuntimeUpgradeFailureCodeMigrationFailed is the stable diagnostic code for failed upgrade migration phases.
	RuntimeUpgradeFailureCodeMigrationFailed = "plugin_upgrade_migration_failed"
	// RuntimeUpgradeFailureMessageKeyMigrationFailed is the i18n key for failed upgrade migration phases.
	RuntimeUpgradeFailureMessageKeyMigrationFailed = "plugin.runtimeUpgrade.failure.migrationFailed"
)

// BuildRuntimeUpgradeState loads release snapshots and returns one runtime
// upgrade state projection for the supplied registry/discovered manifest pair.
func (s *serviceImpl) BuildRuntimeUpgradeState(
	ctx context.Context,
	registry *PluginRecord,
	manifest *catalog.Manifest,
) (plugintypes.RuntimeUpgradeProjection, error) {
	var (
		effectiveSnapshot *ManifestSnapshot
		targetRelease     *ReleaseRecord
		targetSnapshot    *ManifestSnapshot
		err               error
	)

	effectiveRelease, err := s.GetRegistryRelease(ctx, registry)
	if err != nil {
		return plugintypes.RuntimeUpgradeProjection{}, err
	}
	if effectiveRelease != nil {
		effectiveSnapshot, err = s.ParseManifestSnapshot(effectiveRelease.ManifestSnapshot)
		if err != nil {
			return plugintypes.RuntimeUpgradeProjection{}, err
		}
	}
	if manifest != nil {
		targetRelease, err = s.GetRelease(ctx, manifest.ID, manifest.Version)
		if err != nil {
			return plugintypes.RuntimeUpgradeProjection{}, err
		}
		if targetRelease != nil {
			targetSnapshot, err = s.ParseManifestSnapshot(targetRelease.ManifestSnapshot)
			if err != nil {
				return plugintypes.RuntimeUpgradeProjection{}, err
			}
		}
	} else if effectiveRelease != nil {
		targetRelease = effectiveRelease
		targetSnapshot = effectiveSnapshot
	}

	projection := BuildRuntimeUpgradeProjection(registry, manifest, effectiveSnapshot, targetSnapshot, targetRelease)
	if projection.State == plugintypes.RuntimeUpgradeStateUpgradeFailed && targetRelease != nil {
		failure, failureErr := s.BuildRuntimeUpgradeFailureWithLatestMigration(ctx, targetRelease)
		if failureErr != nil {
			return plugintypes.RuntimeUpgradeProjection{}, failureErr
		}
		projection.LastFailure = failure
	}
	return projection, nil
}

// BuildRuntimeUpgradeProjection compares effective registry state against the
// discovered manifest or archived snapshot.
func BuildRuntimeUpgradeProjection(
	registry *PluginRecord,
	manifest *catalog.Manifest,
	effectiveSnapshot *ManifestSnapshot,
	targetSnapshot *ManifestSnapshot,
	targetRelease *ReleaseRecord,
) plugintypes.RuntimeUpgradeProjection {
	// The explicit running marker has precedence over semantic-version drift so
	// management pages can show an in-progress state even before the target
	// release failure or success state is persisted.
	projection := plugintypes.RuntimeUpgradeProjection{
		State: plugintypes.RuntimeUpgradeStateNormal,
	}
	projection.EffectiveVersion = resolveEffectiveRuntimeVersion(registry, effectiveSnapshot)
	projection.DiscoveredVersion = resolveDiscoveredRuntimeVersion(manifest, targetSnapshot)

	if registry == nil || registry.Installed != plugintypes.InstalledYes {
		return projection
	}
	if projection.EffectiveVersion == "" || projection.DiscoveredVersion == "" {
		return projection
	}
	if registryRuntimeStateIsUpgradeRunning(registry) {
		projection.State = plugintypes.RuntimeUpgradeStateUpgradeRunning
		return projection
	}
	if registryRuntimeStateIsFailed(registry) {
		projection.State = plugintypes.RuntimeUpgradeStateUpgradeFailed
		projection.LastFailure = BuildRuntimeUpgradeFailure(targetRelease)
		return projection
	}
	if failed := BuildRuntimeUpgradeFailure(targetRelease); failed != nil {
		projection.State = plugintypes.RuntimeUpgradeStateUpgradeFailed
		projection.LastFailure = failed
		return projection
	}

	versionCompare, err := plugintypes.CompareSemanticVersions(
		projection.EffectiveVersion,
		projection.DiscoveredVersion,
	)
	if err != nil {
		projection.State = plugintypes.RuntimeUpgradeStateAbnormal
		projection.AbnormalReason = plugintypes.RuntimeUpgradeAbnormalReasonVersionCompareFailed
		return projection
	}

	switch {
	case versionCompare < 0:
		projection.UpgradeAvailable = true
		if failed := BuildRuntimeUpgradeFailure(targetRelease); failed != nil {
			projection.State = plugintypes.RuntimeUpgradeStateUpgradeFailed
			projection.LastFailure = failed
			return projection
		}
		projection.State = plugintypes.RuntimeUpgradeStatePendingUpgrade
	case versionCompare > 0:
		projection.State = plugintypes.RuntimeUpgradeStateAbnormal
		projection.AbnormalReason = plugintypes.RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective
	default:
		projection.State = plugintypes.RuntimeUpgradeStateNormal
	}
	return projection
}

// registryRuntimeStateIsUpgradeRunning reports whether the registry carries an
// explicit runtime-upgrade in-progress marker.
func registryRuntimeStateIsUpgradeRunning(registry *PluginRecord) bool {
	if registry == nil {
		return false
	}
	state := strings.TrimSpace(registry.CurrentState)
	return state == plugintypes.RuntimeUpgradeStateUpgradeRunning.String() ||
		state == plugintypes.HostStateReconciling.String()
}

// registryRuntimeStateIsFailed reports whether the registry carries an explicit
// failed runtime marker that must block business entries until repaired.
func registryRuntimeStateIsFailed(registry *PluginRecord) bool {
	if registry == nil {
		return false
	}
	return strings.TrimSpace(registry.CurrentState) == plugintypes.HostStateFailed.String()
}

// RegistryRuntimeStateIsUpgradeRunning reports whether a registry row is in the
// explicit runtime-upgrade running state.
func RegistryRuntimeStateIsUpgradeRunning(registry *PluginRecord) bool {
	return registryRuntimeStateIsUpgradeRunning(registry)
}

// RuntimeStateAllowsBusinessEntry reports whether plugin-owned routes, menus,
// cron jobs, and hooks may execute for the supplied runtime-upgrade state.
func RuntimeStateAllowsBusinessEntry(state plugintypes.RuntimeUpgradeState) bool {
	return state == "" || state == plugintypes.RuntimeUpgradeStateNormal
}

// BuildRuntimeUpgradeFailure projects a failed target release into diagnostic metadata.
func BuildRuntimeUpgradeFailure(release *ReleaseRecord) *plugintypes.RuntimeUpgradeFailure {
	if release == nil || strings.TrimSpace(release.Status) != plugintypes.ReleaseStatusFailed.String() {
		return nil
	}
	return &plugintypes.RuntimeUpgradeFailure{
		Phase:          plugintypes.RuntimeUpgradeFailurePhaseRelease,
		ErrorCode:      RuntimeUpgradeFailureCodeReleaseFailed,
		MessageKey:     RuntimeUpgradeFailureMessageKeyReleaseFailed,
		ReleaseID:      release.Id,
		ReleaseVersion: strings.TrimSpace(release.ReleaseVersion),
	}
}

// BuildRuntimeUpgradeFailureWithLatestMigration projects the failed release and
// augments it with the latest failed upgrade migration or callback phase.
func (s *serviceImpl) BuildRuntimeUpgradeFailureWithLatestMigration(
	ctx context.Context,
	release *ReleaseRecord,
) (*plugintypes.RuntimeUpgradeFailure, error) {
	failure := BuildRuntimeUpgradeFailure(release)
	if failure == nil {
		return nil, nil
	}
	migration, err := s.getLatestFailedUpgradeMigration(ctx, release)
	if err != nil {
		return nil, err
	}
	if migration == nil {
		return failure, nil
	}
	failure.Phase = normalizeRuntimeUpgradeFailurePhase(migration.MigrationKey)
	failure.ErrorCode = RuntimeUpgradeFailureCodeMigrationFailed
	failure.MessageKey = RuntimeUpgradeFailureMessageKeyMigrationFailed
	failure.Detail = strings.TrimSpace(migration.ErrorMessage)
	return failure, nil
}

// getLatestFailedUpgradeMigration returns the newest failed upgrade migration
// row for the target release.
func (s *serviceImpl) getLatestFailedUpgradeMigration(
	ctx context.Context,
	release *ReleaseRecord,
) (*MigrationRecord, error) {
	if release == nil {
		return nil, nil
	}
	var migration *MigrationRecord
	err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId:  release.PluginId,
			ReleaseId: release.Id,
			Phase:     plugintypes.MigrationDirectionUpgrade.String(),
			Status:    plugintypes.MigrationExecutionStatusFailed.String(),
		}).
		OrderDesc(dao.SysPluginMigration.Columns().UpdatedAt).
		OrderDesc(dao.SysPluginMigration.Columns().Id).
		Scan(&migration)
	return migration, err
}

// normalizeRuntimeUpgradeFailurePhase maps persisted migration keys back to
// stable runtime-upgrade failure phase enums.
func normalizeRuntimeUpgradeFailurePhase(migrationKey string) plugintypes.RuntimeUpgradeFailurePhase {
	normalizedKey := strings.TrimSpace(migrationKey)
	switch {
	case strings.Contains(normalizedKey, "before-upgrade") ||
		strings.Contains(normalizedKey, "before_upgrade"):
		return plugintypes.RuntimeUpgradeFailurePhaseBeforeUpgrade
	case strings.Contains(normalizedKey, "upgrade-callback") ||
		strings.Contains(normalizedKey, "upgrade_callback"):
		return plugintypes.RuntimeUpgradeFailurePhaseUpgradeCallback
	case strings.Contains(normalizedKey, "governance"):
		return plugintypes.RuntimeUpgradeFailurePhaseGovernance
	case strings.Contains(normalizedKey, "cache-invalidation") ||
		strings.Contains(normalizedKey, "cache_invalidation") ||
		strings.Contains(normalizedKey, "cache"):
		return plugintypes.RuntimeUpgradeFailurePhaseCacheInvalidation
	case strings.Contains(normalizedKey, "release-switch") ||
		strings.Contains(normalizedKey, "release_switch"):
		return plugintypes.RuntimeUpgradeFailurePhaseReleaseSwitch
	case strings.HasPrefix(normalizedKey, "upgrade-step-") ||
		strings.Contains(normalizedKey, "source-sql") ||
		strings.Contains(normalizedKey, "-sql"):
		return plugintypes.RuntimeUpgradeFailurePhaseSQL
	case strings.Contains(normalizedKey, "release"):
		return plugintypes.RuntimeUpgradeFailurePhaseRelease
	default:
		return plugintypes.RuntimeUpgradeFailurePhaseRelease
	}
}

// resolveEffectiveRuntimeVersion reads the database-effective version from the
// registry row and falls back to the effective release snapshot if needed.
func resolveEffectiveRuntimeVersion(
	registry *PluginRecord,
	snapshot *ManifestSnapshot,
) string {
	if registry != nil && strings.TrimSpace(registry.Version) != "" {
		return strings.TrimSpace(registry.Version)
	}
	if snapshot != nil {
		return strings.TrimSpace(snapshot.Version)
	}
	return ""
}

// resolveDiscoveredRuntimeVersion reads the file-discovered version from the
// manifest, falling back to the target release snapshot when no file is present.
func resolveDiscoveredRuntimeVersion(
	manifest *catalog.Manifest,
	snapshot *ManifestSnapshot,
) string {
	if manifest != nil {
		return strings.TrimSpace(manifest.Version)
	}
	if snapshot != nil {
		return strings.TrimSpace(snapshot.Version)
	}
	return ""
}
