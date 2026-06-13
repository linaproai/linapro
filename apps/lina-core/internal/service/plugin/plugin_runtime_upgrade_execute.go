// This file exposes explicit plugin runtime-upgrade facade methods after the
// unified upgrade component owns execution orchestration.

package plugin

import "context"

// ListSourceUpgradeStatuses scans source manifests and returns one
// effective-versus-discovered upgrade-status item per source plugin.
func (s *serviceImpl) ListSourceUpgradeStatuses(ctx context.Context) ([]*SourceUpgradeStatus, error) {
	return s.upgradeSvc.ListSourceUpgradeStatuses(ctx)
}

// UpgradeSourcePlugin applies one explicit source-plugin runtime upgrade from
// the current effective version to the newer discovered source version.
func (s *serviceImpl) UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourceUpgradeResult, error) {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return nil, err
	}
	return s.upgradeSvc.UpgradeSourcePlugin(ctx, pluginID)
}

// ValidateSourcePluginUpgradeReadiness scans source-plugin version drift
// without failing on pending upgrades.
func (s *serviceImpl) ValidateSourcePluginUpgradeReadiness(ctx context.Context) error {
	return s.upgradeSvc.ValidateSourcePluginUpgradeReadiness(ctx)
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
	return s.upgradeSvc.ExecuteRuntimeUpgrade(ctx, pluginID, options)
}
