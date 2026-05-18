// This file contains the concrete source-plugin upgrade delegation methods.
// It keeps host facade forwarding out of the package entrypoint while retaining
// the same exported service contract and error behavior.

package sourceupgrade

import "context"

// ListSourcePluginStatuses returns the current effective/discovered source-plugin version pairs.
func (s *serviceImpl) ListSourcePluginStatuses(ctx context.Context) ([]*SourcePluginStatus, error) {
	return s.pluginSvc.ListSourceUpgradeStatuses(ctx)
}

// UpgradeSourcePlugin applies one explicit source-plugin upgrade.
func (s *serviceImpl) UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourcePluginUpgradeResult, error) {
	return s.pluginSvc.UpgradeSourcePlugin(ctx, pluginID)
}

// ValidateSourcePluginUpgradeReadiness scans source-plugin version drift without failing on pending upgrades.
func (s *serviceImpl) ValidateSourcePluginUpgradeReadiness(ctx context.Context) error {
	return s.pluginSvc.ValidateSourcePluginUpgradeReadiness(ctx)
}
