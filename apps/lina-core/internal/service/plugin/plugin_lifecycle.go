// This file exposes lifecycle and status methods on the root plugin facade.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
)

// Install executes the install lifecycle and optionally persists one host-confirmed
// host service authorization snapshot when the target is a dynamic plugin.
func (s *serviceImpl) Install(
	ctx context.Context,
	pluginID string,
	authorization *HostServiceAuthorizationInput,
) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if err = s.persistDynamicPluginAuthorization(ctx, manifest, authorization); err != nil {
		return err
	}
	return s.lifecycleSvc.Install(ctx, pluginID)
}

// Uninstall executes the uninstall lifecycle for an installed dynamic plugin.
func (s *serviceImpl) Uninstall(ctx context.Context, pluginID string) error {
	return s.lifecycleSvc.Uninstall(ctx, pluginID)
}

// UpdateStatus updates plugin status, where status is 1=enabled and 0=disabled,
// and optionally persists one host-confirmed host service authorization snapshot
// before enabling a dynamic plugin.
func (s *serviceImpl) UpdateStatus(
	ctx context.Context,
	pluginID string,
	status int,
	authorization *HostServiceAuthorizationInput,
) error {
	return s.updateStatus(ctx, pluginID, status, authorization)
}

func (s *serviceImpl) updateStatus(
	ctx context.Context,
	pluginID string,
	status int,
	authorization *HostServiceAuthorizationInput,
) error {
	if status != catalog.StatusDisabled && status != catalog.StatusEnabled {
		return gerror.New("插件状态仅支持0或1")
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if status == catalog.StatusEnabled && catalog.NormalizeType(manifest.Type) == catalog.TypeDynamic {
		if err = s.runtimeSvc.EnsureRuntimeArtifactAvailable(manifest, "启用"); err != nil {
			return err
		}
	}
	if err = s.SyncSourcePlugins(ctx); err != nil {
		return err
	}
	installed, err := s.runtimeSvc.CheckIsInstalled(ctx, pluginID)
	if err != nil {
		return err
	}
	if !installed {
		return gerror.New("插件未安装")
	}
	if catalog.NormalizeType(manifest.Type) == catalog.TypeDynamic {
		if status == catalog.StatusEnabled {
			if err = s.persistDynamicPluginAuthorization(ctx, manifest, authorization); err != nil {
				return err
			}
		}
		return s.reconcileDynamicPluginStatus(ctx, pluginID, status)
	}
	return s.catalogSvc.SetPluginStatus(ctx, pluginID, status)
}

// Enable enables the specified plugin.
func (s *serviceImpl) Enable(ctx context.Context, pluginID string) error {
	return s.updateStatus(ctx, pluginID, catalog.StatusEnabled, nil)
}

// Disable disables the specified plugin.
func (s *serviceImpl) Disable(ctx context.Context, pluginID string) error {
	return s.updateStatus(ctx, pluginID, catalog.StatusDisabled, nil)
}

// persistDynamicPluginAuthorization refreshes the release snapshot for dynamic
// plugins so install/enable flows can reuse one governance preparation path.
func (s *serviceImpl) persistDynamicPluginAuthorization(
	ctx context.Context,
	manifest *catalog.Manifest,
	authorization *HostServiceAuthorizationInput,
) error {
	if manifest == nil || catalog.NormalizeType(manifest.Type) != catalog.TypeDynamic {
		return nil
	}
	if _, err := s.catalogSvc.SyncManifest(ctx, manifest); err != nil {
		return err
	}
	if _, err := s.catalogSvc.PersistReleaseHostServiceAuthorization(ctx, manifest, authorization); err != nil {
		return err
	}
	return nil
}

// reconcileDynamicPluginStatus converts facade enable/disable requests into the
// runtime reconciler host state transitions used by dynamic plugins.
func (s *serviceImpl) reconcileDynamicPluginStatus(ctx context.Context, pluginID string, status int) error {
	targetState := catalog.HostStateInstalled.String()
	if status == catalog.StatusEnabled {
		targetState = catalog.HostStateEnabled.String()
	}
	return s.runtimeSvc.ReconcileDynamicPluginRequest(ctx, pluginID, targetState)
}

// IsInstalled returns whether a plugin is installed.
func (s *serviceImpl) IsInstalled(ctx context.Context, pluginID string) bool {
	installed, err := s.runtimeSvc.CheckIsInstalled(ctx, pluginID)
	return err == nil && installed
}

// IsEnabled returns whether a plugin is enabled.
func (s *serviceImpl) IsEnabled(ctx context.Context, pluginID string) bool {
	return s.integrationSvc.IsEnabled(ctx, pluginID)
}
