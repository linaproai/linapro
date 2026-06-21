// This file contains distribution-aware governance guards for public plugin
// management write paths. Startup-internal reconciliation intentionally bypasses
// these helpers and uses narrower lifecycle contracts.

package plugin

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/bizerr"
)

// ensureBuiltinManagementActionAllowed rejects ordinary management mutations
// for project built-in plugins. It consults both the registry and the desired
// manifest so uninstalled builtin plugins are guarded before a registry row
// exists, while registry-only dynamic cleanup paths can continue when the
// mutable manifest is unavailable.
func (s *serviceImpl) ensureBuiltinManagementActionAllowed(ctx context.Context, pluginID string) error {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil
	}
	if s == nil {
		return nil
	}

	if s.storeSvc != nil {
		registry, err := s.storeSvc.GetRegistry(ctx, normalizedPluginID)
		if err != nil {
			return err
		}
		if registry != nil && registry.Distribution == plugintypes.DistributionBuiltin.String() {
			return bizerr.NewCode(CodePluginBuiltinManagementActionDenied, bizerr.P("pluginId", normalizedPluginID))
		}
	}

	if s.catalogSvc == nil {
		return nil
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(normalizedPluginID)
	if err != nil || manifest == nil {
		return nil
	}
	if plugintypes.NormalizeDistribution(manifest.Distribution) == plugintypes.DistributionBuiltin {
		return bizerr.NewCode(CodePluginBuiltinManagementActionDenied, bizerr.P("pluginId", normalizedPluginID))
	}
	return nil
}
