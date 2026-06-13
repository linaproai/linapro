// This file keeps the root facade's runtime-upgrade preview API as a thin
// governance-facing delegate to the unified upgrade component.

package plugin

import (
	"context"

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
	return s.upgradeSvc.PreviewRuntimeUpgrade(ctx, pluginID)
}
