// This file implements dynamic plugin install and uninstall lifecycle entry
// points that delegate convergence to the runtime reconciler.

package lifecycle

import (
	"context"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/bizerr"
)

// InstallDynamic executes the install lifecycle for a discovered dynamic plugin.
// Repeated installs are treated as idempotent unless the same version needs a refresh.
func (s *serviceImpl) InstallDynamic(ctx context.Context, pluginID string) error {
	return s.installDynamic(ctx, pluginID, nil, runtime.DynamicReconcileOptions{})
}

func (s *serviceImpl) installDynamic(
	ctx context.Context,
	pluginID string,
	desiredManifest *catalog.Manifest,
	options runtime.DynamicReconcileOptions,
) error {
	manifest := catalog.CloneManifest(desiredManifest)
	var err error
	if manifest == nil {
		if s.catalogSvc != nil {
			s.catalogSvc.InvalidateManifestCache(pluginID)
		}
		manifest, err = s.catalogSvc.GetDesiredManifest(pluginID)
		if err != nil {
			return err
		}
	}
	if plugintypes.NormalizeType(manifest.Type) == plugintypes.TypeSource {
		return bizerr.NewCode(CodeSourcePluginInstallUnsupported)
	}
	if s.runtimeSvc != nil {
		if err = s.runtimeSvc.EnsureRuntimeArtifactAvailable(manifest, "install"); err != nil {
			return err
		}
	}

	existingRegistry, err := s.storeSvc.GetRegistry(ctx, manifest.ID)
	if err != nil {
		return err
	}
	if existingRegistry != nil && existingRegistry.Installed == plugintypes.InstalledYes {
		compareResult, compareErr := plugintypes.CompareSemanticVersions(manifest.Version, existingRegistry.Version)
		if compareErr != nil {
			return compareErr
		}
		if compareResult < 0 {
			return bizerr.NewCode(CodeDynamicPluginDowngradeUnsupported)
		}
	}

	registry, err := s.storeSvc.SyncManifest(ctx, manifest)
	if err != nil {
		return err
	}
	if registry.Installed == plugintypes.InstalledYes {
		compareResult, compareErr := plugintypes.CompareSemanticVersions(manifest.Version, registry.Version)
		if compareErr != nil {
			return compareErr
		}
		if compareResult == 0 {
			if s.runtimeSvc != nil && !s.runtimeSvc.ShouldRefreshInstalledDynamicRelease(ctx, registry, manifest) {
				return nil
			}
		}
	}

	desiredState := plugintypes.HostStateInstalled.String()
	if registry.Installed == plugintypes.InstalledYes && registry.Status == plugintypes.StatusEnabled {
		desiredState = plugintypes.HostStateEnabled.String()
	}
	if s.runtimeSvc != nil {
		options.DesiredManifest = catalog.CloneManifest(manifest)
		if err = s.runtimeSvc.ReconcileDynamicPluginRequest(ctx, pluginID, desiredState, options); err != nil {
			return err
		}
	}
	return nil
}

// UninstallDynamic executes the low-level uninstall lifecycle for an installed
// dynamic plugin.
func (s *serviceImpl) UninstallDynamic(ctx context.Context, pluginID string) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if plugintypes.NormalizeType(manifest.Type) == plugintypes.TypeSource {
		return bizerr.NewCode(CodeSourcePluginUninstallUnsupported)
	}

	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || registry.Installed != plugintypes.InstalledYes {
		return nil
	}
	if s.runtimeSvc != nil {
		return s.runtimeSvc.ReconcileDynamicPluginRequest(ctx, pluginID, plugintypes.HostStateUninstalled.String(), runtime.DynamicReconcileOptions{})
	}
	return nil
}
