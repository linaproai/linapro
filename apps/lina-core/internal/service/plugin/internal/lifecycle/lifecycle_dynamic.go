// This file implements dynamic plugin install lifecycle helpers that delegate
// convergence to the runtime reconciler.

package lifecycle

import (
	"context"
	pluginv1 "lina-core/api/plugin/v1"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/statusflag"
)

// installDynamic reconciles one discovered dynamic plugin into an installed runtime state.
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
	if plugintypes.NormalizeType(manifest.Type) == pluginv1.PluginTypeSource {
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
	if existingRegistry != nil && existingRegistry.Installed == statusflag.Installed.Int() {
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
	if registry.Installed == statusflag.Installed.Int() {
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
	if registry.Installed == statusflag.Installed.Int() && registry.Status == statusflag.EnabledValue.Int() {
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
