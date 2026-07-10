// This file exposes runtime and source upgrade orchestration through the
// lifecycle facade. The upgrade package keeps plan/execute implementation
// details; lifecycle owns construction attachment and the root-facing API.

package lifecycle

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/upgrade"
)

// BindUpgrade attaches the composition-root upgrade service.
func (s *serviceImpl) BindUpgrade(upgradeSvc upgrade.Service) error {
	if s == nil {
		return gerror.New("lifecycle service cannot bind upgrade through nil receiver")
	}
	if upgradeSvc == nil {
		return gerror.New("lifecycle service cannot bind nil upgrade service")
	}
	s.upgradeSvc = upgradeSvc
	return nil
}

// ListSourceUpgradeStatuses delegates to the bound upgrade service.
func (s *serviceImpl) ListSourceUpgradeStatuses(ctx context.Context) ([]*upgrade.SourceUpgradeStatus, error) {
	upgradeSvc, err := s.upgradeService()
	if err != nil {
		return nil, err
	}
	return upgradeSvc.ListSourceUpgradeStatuses(ctx)
}

// ValidateSourcePluginUpgradeReadiness delegates to the bound upgrade service.
func (s *serviceImpl) ValidateSourcePluginUpgradeReadiness(ctx context.Context) error {
	upgradeSvc, err := s.upgradeService()
	if err != nil {
		return err
	}
	return upgradeSvc.ValidateSourcePluginUpgradeReadiness(ctx)
}

// PreviewRuntimeUpgrade delegates to the bound upgrade service.
func (s *serviceImpl) PreviewRuntimeUpgrade(ctx context.Context, pluginID string) (*upgrade.RuntimeUpgradePreview, error) {
	upgradeSvc, err := s.upgradeService()
	if err != nil {
		return nil, err
	}
	return upgradeSvc.PreviewRuntimeUpgrade(ctx, pluginID)
}

// UpgradeSourcePlugin delegates to the bound upgrade service.
func (s *serviceImpl) UpgradeSourcePlugin(ctx context.Context, pluginID string) (*upgrade.SourceUpgradeResult, error) {
	upgradeSvc, err := s.upgradeService()
	if err != nil {
		return nil, err
	}
	return upgradeSvc.UpgradeSourcePlugin(ctx, pluginID)
}

// ExecuteRuntimeUpgrade delegates to the bound upgrade service.
func (s *serviceImpl) ExecuteRuntimeUpgrade(
	ctx context.Context,
	pluginID string,
	options upgrade.RuntimeUpgradeOptions,
) (*upgrade.RuntimeUpgradeResult, error) {
	upgradeSvc, err := s.upgradeService()
	if err != nil {
		return nil, err
	}
	return upgradeSvc.ExecuteRuntimeUpgrade(ctx, pluginID, options)
}

func (s *serviceImpl) upgradeService() (upgrade.Service, error) {
	if s == nil || s.upgradeSvc == nil {
		return nil, gerror.New("lifecycle upgrade service is not bound")
	}
	return s.upgradeSvc, nil
}
