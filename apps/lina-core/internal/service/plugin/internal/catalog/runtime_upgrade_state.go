// This file computes runtime-upgrade state from discovered manifest metadata
// and effective registry/release rows without mutating governance tables.

package catalog

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
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

// RuntimeUpgradeFailure exposes the latest observable runtime-upgrade failure.
type RuntimeUpgradeFailure struct {
	// Phase is the upgrade phase associated with the failure.
	Phase RuntimeUpgradeFailurePhase
	// ErrorCode is a stable machine-readable failure code.
	ErrorCode string
	// MessageKey is the i18n key that management clients can render.
	MessageKey string
	// ReleaseID identifies the failed target release when known.
	ReleaseID int
	// ReleaseVersion identifies the failed target release version when known.
	ReleaseVersion string
	// Detail carries the latest persisted failure detail for operator diagnosis.
	Detail string
}

// RuntimeUpgradeProjection is the flattened version-drift state for one plugin.
type RuntimeUpgradeProjection struct {
	// State is the current runtime-upgrade state.
	State RuntimeUpgradeState
	// EffectiveVersion is the version currently active in sys_plugin.
	EffectiveVersion string
	// DiscoveredVersion is the version currently discovered from plugin files.
	DiscoveredVersion string
	// UpgradeAvailable reports whether a user can attempt a runtime upgrade.
	UpgradeAvailable bool
	// AbnormalReason stores a stable reason code when State is abnormal.
	AbnormalReason RuntimeUpgradeAbnormalReason
	// LastFailure stores the latest observable failed target release.
	LastFailure *RuntimeUpgradeFailure
}

// BuildRuntimeUpgradeState loads release snapshots and returns one runtime
// upgrade state projection for the supplied registry/discovered manifest pair.
func (s *serviceImpl) BuildRuntimeUpgradeState(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *Manifest,
) (RuntimeUpgradeProjection, error) {
	var (
		effectiveSnapshot *ManifestSnapshot
		targetRelease     *entity.SysPluginRelease
		targetSnapshot    *ManifestSnapshot
		err               error
	)

	effectiveRelease, err := s.GetRegistryRelease(ctx, registry)
	if err != nil {
		return RuntimeUpgradeProjection{}, err
	}
	if effectiveRelease != nil {
		effectiveSnapshot, err = s.ParseManifestSnapshot(effectiveRelease.ManifestSnapshot)
		if err != nil {
			return RuntimeUpgradeProjection{}, err
		}
	}
	if manifest != nil {
		targetRelease, err = s.GetRelease(ctx, manifest.ID, manifest.Version)
		if err != nil {
			return RuntimeUpgradeProjection{}, err
		}
		if targetRelease != nil {
			targetSnapshot, err = s.ParseManifestSnapshot(targetRelease.ManifestSnapshot)
			if err != nil {
				return RuntimeUpgradeProjection{}, err
			}
		}
	} else if effectiveRelease != nil {
		targetRelease = effectiveRelease
		targetSnapshot = effectiveSnapshot
	}

	projection := BuildRuntimeUpgradeProjection(registry, manifest, effectiveSnapshot, targetSnapshot, targetRelease)
	if projection.State == RuntimeUpgradeStateUpgradeFailed && targetRelease != nil {
		failure, failureErr := s.BuildRuntimeUpgradeFailureWithLatestMigration(ctx, targetRelease)
		if failureErr != nil {
			return RuntimeUpgradeProjection{}, failureErr
		}
		projection.LastFailure = failure
	}
	return projection, nil
}

// BuildRuntimeUpgradeProjection compares effective registry state against the
// discovered manifest or archived snapshot.
func BuildRuntimeUpgradeProjection(
	registry *entity.SysPlugin,
	manifest *Manifest,
	effectiveSnapshot *ManifestSnapshot,
	targetSnapshot *ManifestSnapshot,
	targetRelease *entity.SysPluginRelease,
) RuntimeUpgradeProjection {
	// The explicit running marker has precedence over semantic-version drift so
	// management pages can show an in-progress state even before the target
	// release failure or success state is persisted.
	projection := RuntimeUpgradeProjection{
		State: RuntimeUpgradeStateNormal,
	}
	projection.EffectiveVersion = resolveEffectiveRuntimeVersion(registry, effectiveSnapshot)
	projection.DiscoveredVersion = resolveDiscoveredRuntimeVersion(manifest, targetSnapshot)

	if registry == nil || registry.Installed != InstalledYes {
		return projection
	}
	if projection.EffectiveVersion == "" || projection.DiscoveredVersion == "" {
		return projection
	}
	if registryRuntimeStateIsUpgradeRunning(registry) {
		projection.State = RuntimeUpgradeStateUpgradeRunning
		return projection
	}

	versionCompare, err := CompareSemanticVersions(
		projection.EffectiveVersion,
		projection.DiscoveredVersion,
	)
	if err != nil {
		projection.State = RuntimeUpgradeStateAbnormal
		projection.AbnormalReason = RuntimeUpgradeAbnormalReasonVersionCompareFailed
		return projection
	}

	switch {
	case versionCompare < 0:
		projection.UpgradeAvailable = true
		if failed := BuildRuntimeUpgradeFailure(targetRelease); failed != nil {
			projection.State = RuntimeUpgradeStateUpgradeFailed
			projection.LastFailure = failed
			return projection
		}
		projection.State = RuntimeUpgradeStatePendingUpgrade
	case versionCompare > 0:
		projection.State = RuntimeUpgradeStateAbnormal
		projection.AbnormalReason = RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective
	default:
		projection.State = RuntimeUpgradeStateNormal
	}
	return projection
}

// registryRuntimeStateIsUpgradeRunning reports whether the registry carries an
// explicit runtime-upgrade in-progress marker.
func registryRuntimeStateIsUpgradeRunning(registry *entity.SysPlugin) bool {
	if registry == nil {
		return false
	}
	state := strings.TrimSpace(registry.CurrentState)
	return state == RuntimeUpgradeStateUpgradeRunning.String() ||
		state == HostStateReconciling.String()
}

// RegistryRuntimeStateIsUpgradeRunning reports whether a registry row is in the
// explicit runtime-upgrade running state.
func RegistryRuntimeStateIsUpgradeRunning(registry *entity.SysPlugin) bool {
	return registryRuntimeStateIsUpgradeRunning(registry)
}

// RuntimeStateAllowsBusinessEntry reports whether plugin-owned routes, menus,
// cron jobs, and hooks may execute for the supplied runtime-upgrade state.
func RuntimeStateAllowsBusinessEntry(state RuntimeUpgradeState) bool {
	return state == "" || state == RuntimeUpgradeStateNormal
}

// BuildRuntimeUpgradeFailure projects a failed target release into diagnostic metadata.
func BuildRuntimeUpgradeFailure(release *entity.SysPluginRelease) *RuntimeUpgradeFailure {
	if release == nil || strings.TrimSpace(release.Status) != ReleaseStatusFailed.String() {
		return nil
	}
	return &RuntimeUpgradeFailure{
		Phase:          RuntimeUpgradeFailurePhaseRelease,
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
	release *entity.SysPluginRelease,
) (*RuntimeUpgradeFailure, error) {
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
	release *entity.SysPluginRelease,
) (*entity.SysPluginMigration, error) {
	if release == nil {
		return nil, nil
	}
	var migration *entity.SysPluginMigration
	err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId:  release.PluginId,
			ReleaseId: release.Id,
			Phase:     MigrationDirectionUpgrade.String(),
			Status:    MigrationExecutionStatusFailed.String(),
		}).
		OrderDesc(dao.SysPluginMigration.Columns().UpdatedAt).
		OrderDesc(dao.SysPluginMigration.Columns().Id).
		Scan(&migration)
	return migration, err
}

// normalizeRuntimeUpgradeFailurePhase maps persisted migration keys back to
// stable runtime-upgrade failure phase enums.
func normalizeRuntimeUpgradeFailurePhase(migrationKey string) RuntimeUpgradeFailurePhase {
	normalizedKey := strings.TrimSpace(migrationKey)
	switch {
	case strings.Contains(normalizedKey, "before-upgrade"):
		return RuntimeUpgradeFailurePhaseBeforeUpgrade
	case strings.Contains(normalizedKey, "upgrade-callback"):
		return RuntimeUpgradeFailurePhaseUpgradeCallback
	case strings.Contains(normalizedKey, "governance-"):
		return RuntimeUpgradeFailurePhaseGovernance
	case strings.Contains(normalizedKey, "release"):
		return RuntimeUpgradeFailurePhaseReleaseSwitch
	case strings.Contains(normalizedKey, "cache"):
		return RuntimeUpgradeFailurePhaseCacheInvalidation
	case strings.HasPrefix(normalizedKey, "upgrade-step-"):
		return RuntimeUpgradeFailurePhaseSQL
	default:
		return RuntimeUpgradeFailurePhaseRelease
	}
}

// resolveEffectiveRuntimeVersion reads the database-effective version from the
// registry row and falls back to the effective release snapshot if needed.
func resolveEffectiveRuntimeVersion(
	registry *entity.SysPlugin,
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
	manifest *Manifest,
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
