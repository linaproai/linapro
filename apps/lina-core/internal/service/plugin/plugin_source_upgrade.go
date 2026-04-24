// This file exposes the plugin-service facade for source-plugin upgrade
// governance by delegating to the dedicated internal sourceupgrade component.

package plugin

import (
	"context"

	sourceupgradeinternal "lina-core/internal/service/plugin/internal/sourceupgrade"
	sourceupgradecontract "lina-core/pkg/sourceupgrade/contract"
)

type (
	// SourceUpgradeStatus aliases the stable source-plugin upgrade status contract.
	SourceUpgradeStatus = sourceupgradecontract.SourcePluginStatus

	// SourceUpgradeResult aliases the stable explicit source-plugin upgrade result contract.
	SourceUpgradeResult = sourceupgradecontract.SourcePluginUpgradeResult
)

// ListSourceUpgradeStatuses scans source manifests and returns one
// effective-versus-discovered upgrade-status item per source plugin.
func (s *serviceImpl) ListSourceUpgradeStatuses(ctx context.Context) ([]*SourceUpgradeStatus, error) {
	return s.sourceUpgradeSvc.ListSourceUpgradeStatuses(ctx)
}

// UpgradeSourcePlugin applies one explicit source-plugin upgrade from the
// current effective version to the newer discovered source version.
func (s *serviceImpl) UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourceUpgradeResult, error) {
	return s.sourceUpgradeSvc.UpgradeSourcePlugin(ctx, pluginID)
}

// ValidateSourcePluginUpgradeReadiness fails fast when any installed source
// plugin still has a newer discovered source version waiting to be upgraded.
func (s *serviceImpl) ValidateSourcePluginUpgradeReadiness(ctx context.Context) error {
	return s.sourceUpgradeSvc.ValidateSourcePluginUpgradeReadiness(ctx)
}

// Ensure the plugin facade keeps delegating to the dedicated source-upgrade component.
var _ sourceupgradeinternal.Service = (*serviceImpl)(nil)
