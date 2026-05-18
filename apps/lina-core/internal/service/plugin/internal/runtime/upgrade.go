// This file exposes explicit dynamic-plugin runtime upgrade execution while
// keeping background reconciliation side-effect free for version drift.

package runtime

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
)

// UpgradeDynamicPluginRequest runs the version-switching upgrade path for one
// installed dynamic plugin. It is intentionally separate from ordinary
// reconciliation so staged higher-version artifacts do not become effective
// until a management caller confirms the runtime upgrade.
func (s *serviceImpl) UpgradeDynamicPluginRequest(ctx context.Context, pluginID string) error {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return gerror.New("dynamic plugin ID cannot be empty")
	}

	reconcileMu.Lock()
	defer reconcileMu.Unlock()

	registry, err := s.catalogSvc.GetRegistry(ctx, normalizedPluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		return gerror.Newf("dynamic plugin registry does not exist: %s", normalizedPluginID)
	}
	registry, err = s.reconcileRegistryArtifactState(ctx, registry)
	if err != nil {
		return err
	}
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return gerror.Newf("dynamic plugin registry is not dynamic: %s", normalizedPluginID)
	}
	if registry.Installed != catalog.InstalledYes {
		return gerror.Newf("dynamic plugin is not installed: %s", normalizedPluginID)
	}

	desiredManifest, err := s.catalogSvc.GetDesiredManifest(normalizedPluginID)
	if err != nil {
		return err
	}
	if desiredManifest == nil || catalog.NormalizeType(desiredManifest.Type) != catalog.TypeDynamic {
		return gerror.Newf("dynamic plugin desired manifest does not exist: %s", normalizedPluginID)
	}
	if strings.TrimSpace(desiredManifest.Version) == strings.TrimSpace(registry.Version) {
		return nil
	}

	desiredState := catalog.BuildStableHostState(registry)
	return s.applyUpgrade(ctx, registry, desiredManifest, desiredState)
}
