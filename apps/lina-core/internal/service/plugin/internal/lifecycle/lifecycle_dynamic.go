// This file implements dynamic plugin install and uninstall lifecycle entry
// points that delegate convergence to the runtime reconciler.

package lifecycle

import (
	"context"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
)

// Install executes the install lifecycle for a discovered dynamic plugin.
// Repeated installs are treated as idempotent unless the same version needs a refresh.
func (s *serviceImpl) Install(ctx context.Context, pluginID string) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if catalog.NormalizeType(manifest.Type) == catalog.TypeSource {
		return bizerr.NewCode(CodeSourcePluginInstallUnsupported)
	}
	if s.reconciler != nil {
		if err = s.reconciler.EnsureRuntimeArtifactAvailable(manifest, "install"); err != nil {
			return err
		}
	}

	registry, err := s.catalogSvc.SyncManifest(ctx, manifest)
	if err != nil {
		return err
	}
	if registry.Installed == catalog.InstalledYes {
		compareResult, compareErr := catalog.CompareSemanticVersions(manifest.Version, registry.Version)
		if compareErr != nil {
			return compareErr
		}
		if compareResult < 0 {
			return bizerr.NewCode(CodeDynamicPluginDowngradeUnsupported)
		}
		if compareResult == 0 {
			if s.reconciler != nil && !s.reconciler.ShouldRefreshInstalledDynamicRelease(ctx, registry, manifest) {
				return nil
			}
		}
	}

	desiredState := catalog.HostStateInstalled.String()
	if registry.Installed == catalog.InstalledYes && registry.Status == catalog.StatusEnabled {
		desiredState = catalog.HostStateEnabled.String()
	}
	if s.reconciler != nil {
		if err = s.reconciler.ReconcileDynamicPluginRequest(ctx, pluginID, desiredState); err != nil {
			return err
		}
	}
	return nil
}

// Uninstall executes the uninstall lifecycle for an installed dynamic plugin.
func (s *serviceImpl) Uninstall(ctx context.Context, pluginID string) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if catalog.NormalizeType(manifest.Type) == catalog.TypeSource {
		return bizerr.NewCode(CodeSourcePluginUninstallUnsupported)
	}

	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || registry.Installed != catalog.InstalledYes {
		return nil
	}
	if s.reconciler != nil {
		return s.reconciler.ReconcileDynamicPluginRequest(ctx, pluginID, catalog.HostStateUninstalled.String())
	}
	return nil
}
