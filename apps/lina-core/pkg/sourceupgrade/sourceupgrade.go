// Package sourceupgrade exposes stable source-plugin runtime upgrade contracts
// and a host-facing facade.
package sourceupgrade

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	sourceupgradecontract "lina-core/pkg/sourceupgrade/contract"
)

type (
	// Service re-exports the stable source-plugin upgrade governance contract.
	Service = sourceupgradecontract.Service

	// SourcePluginStatus re-exports the stable source-plugin upgrade status contract.
	SourcePluginStatus = sourceupgradecontract.SourcePluginStatus

	// SourcePluginUpgradeResult re-exports the stable explicit source-plugin
	// upgrade result contract.
	SourcePluginUpgradeResult = sourceupgradecontract.SourcePluginUpgradeResult
)

const (
	// SourcePluginEnabledNo marks a source plugin as disabled in the upgrade snapshot.
	SourcePluginEnabledNo = sourceupgradecontract.SourcePluginEnabledNo
	// SourcePluginEnabledYes marks a source plugin as enabled in the upgrade snapshot.
	SourcePluginEnabledYes = sourceupgradecontract.SourcePluginEnabledYes
	// SourcePluginInstalledNo marks a source plugin as not installed in the upgrade snapshot.
	SourcePluginInstalledNo = sourceupgradecontract.SourcePluginInstalledNo
	// SourcePluginInstalledYes marks a source plugin as installed in the upgrade snapshot.
	SourcePluginInstalledYes = sourceupgradecontract.SourcePluginInstalledYes
)

// Ensure serviceImpl satisfies the published source-plugin upgrade contract.
var _ Service = (*serviceImpl)(nil)

// UpgradeGovernanceService narrows the host plugin facade to the operations
// required by the published source-upgrade helper.
type UpgradeGovernanceService interface {
	// ListSourceUpgradeStatuses scans source manifests and returns source-plugin upgrade status.
	ListSourceUpgradeStatuses(ctx context.Context) ([]*sourceupgradecontract.SourcePluginStatus, error)
	// UpgradeSourcePlugin applies one explicit source-plugin upgrade.
	UpgradeSourcePlugin(ctx context.Context, pluginID string) (*sourceupgradecontract.SourcePluginUpgradeResult, error)
	// ValidateSourcePluginUpgradeReadiness scans source-plugin version drift without failing on pending upgrades.
	ValidateSourcePluginUpgradeReadiness(ctx context.Context) error
}

// serviceImpl delegates to the host plugin service while exposing only the
// stable source-upgrade contract needed by callers.
type serviceImpl struct {
	// pluginSvc is the host source-plugin upgrade governance facade.
	pluginSvc UpgradeGovernanceService
}

// New creates and returns a new source-plugin upgrade helper service.
func New(pluginSvc UpgradeGovernanceService) (Service, error) {
	if pluginSvc == nil {
		return nil, gerror.New("sourceupgrade service requires a non-nil plugin upgrade service")
	}
	return &serviceImpl{
		pluginSvc: pluginSvc,
	}, nil
}
