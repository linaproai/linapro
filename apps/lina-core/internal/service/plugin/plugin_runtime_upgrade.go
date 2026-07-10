// This file exposes plugin runtime-upgrade facade methods and narrow cache
// adapters after the unified upgrade component owns preview and execution
// orchestration.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/upgrade"
)

const (
	// RuntimeUpgradeRiskHintUpgradeSQLRequiresReview warns that upgrade SQL
	// should be reviewed before the user confirms runtime side effects.
	RuntimeUpgradeRiskHintUpgradeSQLRequiresReview = upgrade.RuntimeUpgradeRiskHintUpgradeSQLRequiresReview
	// RuntimeUpgradeRiskHintMockSQLExcluded warns that mock SQL is never loaded by upgrade.
	RuntimeUpgradeRiskHintMockSQLExcluded = upgrade.RuntimeUpgradeRiskHintMockSQLExcluded
	// RuntimeUpgradeRiskHintHostServiceAuthorizationChanged warns that hostServices changed.
	RuntimeUpgradeRiskHintHostServiceAuthorizationChanged = upgrade.RuntimeUpgradeRiskHintHostServiceAuthorizationChanged
	// RuntimeUpgradeRiskHintDependencyBlockers warns that dependency checks found hard blockers.
	RuntimeUpgradeRiskHintDependencyBlockers = upgrade.RuntimeUpgradeRiskHintDependencyBlockers
)

// PreviewRuntimeUpgrade returns a side-effect-free upgrade preview for one
// plugin currently marked as pending or failed runtime upgrade.
func (s *serviceImpl) PreviewRuntimeUpgrade(ctx context.Context, pluginID string) (*RuntimeUpgradePreview, error) {
	return s.lifecycleSvc.PreviewRuntimeUpgrade(ctx, pluginID)
}

// ListSourceUpgradeStatuses scans source manifests and returns one
// effective-versus-discovered upgrade-status item per source plugin.
func (s *serviceImpl) ListSourceUpgradeStatuses(ctx context.Context) ([]*SourceUpgradeStatus, error) {
	return s.lifecycleSvc.ListSourceUpgradeStatuses(ctx)
}

// UpgradeSourcePlugin applies one explicit source-plugin runtime upgrade from
// the current effective version to the newer discovered source version.
func (s *serviceImpl) UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourceUpgradeResult, error) {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return nil, err
	}
	if err := s.ensureBuiltinManagementActionAllowed(ctx, pluginID); err != nil {
		return nil, err
	}
	return s.lifecycleSvc.UpgradeSourcePlugin(ctx, pluginID)
}

// ValidateSourcePluginUpgradeReadiness scans source-plugin version drift
// without failing on pending upgrades.
func (s *serviceImpl) ValidateSourcePluginUpgradeReadiness(ctx context.Context) error {
	return s.lifecycleSvc.ValidateSourcePluginUpgradeReadiness(ctx)
}

// ExecuteRuntimeUpgrade runs one explicit runtime upgrade after confirmation.
func (s *serviceImpl) ExecuteRuntimeUpgrade(
	ctx context.Context,
	pluginID string,
	options RuntimeUpgradeOptions,
) (*RuntimeUpgradeResult, error) {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return nil, err
	}
	if err := s.ensureBuiltinManagementActionAllowed(ctx, pluginID); err != nil {
		return nil, err
	}
	return s.lifecycleSvc.ExecuteRuntimeUpgrade(ctx, pluginID, options)
}

// upgradeCachePublisher publishes upgrade cache changes through the root
// facade's single plugin-change path.
type upgradeCachePublisher struct {
	service *serviceImpl
}

// PublishPluginChange publishes a plugin-scoped mutation reason.
func (p upgradeCachePublisher) PublishPluginChange(
	ctx context.Context,
	pluginID string,
	pluginType string,
	reason string,
) error {
	if p.service == nil {
		return gerror.New("plugin upgrade cache publisher is not configured")
	}
	return p.service.PublishPluginChange(ctx, pluginID, pluginType, reason)
}

// SyncEnabledSnapshotAndPublishRuntimeChange refreshes local enablement and
// publishes a scoped mutation through the root facade.
func (p upgradeCachePublisher) SyncEnabledSnapshotAndPublishRuntimeChange(
	ctx context.Context,
	pluginID string,
	reason string,
) error {
	if p.service == nil {
		return gerror.New("plugin upgrade cache publisher is not configured")
	}
	return p.service.syncEnabledSnapshotAndPublishRuntimeChange(ctx, pluginID, reason)
}

// upgradeCacheFreshener refreshes runtime caches before read-only upgrade paths.
type upgradeCacheFreshener struct {
	service *serviceImpl
}

// EnsureRuntimeCacheFresh synchronizes local runtime caches with the shared revision.
func (f upgradeCacheFreshener) EnsureRuntimeCacheFresh(ctx context.Context) error {
	if f.service == nil {
		return gerror.New("plugin upgrade cache freshener is not configured")
	}
	return f.service.ensureRuntimeCacheFresh(ctx)
}
