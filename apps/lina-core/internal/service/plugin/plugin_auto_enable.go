// This file coordinates startup-time plugin bootstrap so plugin.autoEnable can
// install and enable required plugins before later host wiring runs.

package plugin

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

// startupAutoEnableWaitTimeout bounds how long host startup waits for one
// required plugin to reach the enabled state before failing fast.
const startupAutoEnableWaitTimeout = 15 * time.Second

// startupAutoEnablePollInterval is the registry polling cadence used while the
// current node waits to become primary or waits for another primary to converge
// one required plugin.
const startupAutoEnablePollInterval = 100 * time.Millisecond

// BootstrapAutoEnable synchronizes manifests and ensures every plugin listed
// in plugin.autoEnable is installed and enabled before later host wiring runs.
func (s *serviceImpl) BootstrapAutoEnable(ctx context.Context) error {
	if err := s.SyncSourcePlugins(ctx); err != nil {
		return err
	}

	pluginIDs := s.configSvc.GetPluginAutoEnable(ctx)
	if len(pluginIDs) == 0 {
		return nil
	}

	for _, pluginID := range pluginIDs {
		if err := s.bootstrapAutoEnablePlugin(ctx, pluginID); err != nil {
			return err
		}
	}

	if err := s.integrationSvc.RefreshEnabledSnapshot(ctx); err != nil {
		return gerror.Wrap(err, "刷新插件启用快照失败")
	}
	return nil
}

// bootstrapAutoEnablePlugin routes one configured plugin ID into the matching
// source-plugin or dynamic-plugin startup bootstrap path.
func (s *serviceImpl) bootstrapAutoEnablePlugin(ctx context.Context, pluginID string) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return gerror.Wrapf(err, "启动期自动启用插件 %s 失败：发现插件", pluginID)
	}
	if manifest == nil {
		return gerror.Newf("启动期自动启用插件 %s 失败：插件清单不存在", pluginID)
	}

	switch catalog.NormalizeType(manifest.Type) {
	case catalog.TypeSource:
		return s.bootstrapAutoEnableSourcePlugin(ctx, manifest)
	case catalog.TypeDynamic:
		return s.bootstrapAutoEnableDynamicPlugin(ctx, manifest)
	default:
		return gerror.Newf("启动期自动启用插件 %s 失败：不支持的插件类型 %s", pluginID, manifest.Type)
	}
}

// bootstrapAutoEnableSourcePlugin ensures one required source plugin reaches
// the enabled state during startup, while cluster followers wait for the
// elected primary to perform shared lifecycle work.
func (s *serviceImpl) bootstrapAutoEnableSourcePlugin(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return gerror.New("启动期自动启用源码插件失败：插件清单不能为空")
	}

	return s.ensurePluginEnabledDuringStartup(ctx, manifest.ID, func() error {
		if err := s.Install(ctx, manifest.ID, nil); err != nil {
			return gerror.Wrap(err, "安装源码插件失败")
		}
		if err := s.Enable(ctx, manifest.ID); err != nil {
			return gerror.Wrap(err, "启用源码插件失败")
		}
		return nil
	})
}

// bootstrapAutoEnableDynamicPlugin ensures one required dynamic plugin can
// reuse its confirmed authorization snapshot and then reaches the enabled state
// during startup.
func (s *serviceImpl) bootstrapAutoEnableDynamicPlugin(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return gerror.New("启动期自动启用动态插件失败：插件清单不能为空")
	}
	if err := s.ensureDynamicPluginAutoEnableAuthorization(ctx, manifest); err != nil {
		return gerror.Wrapf(err, "启动期自动启用插件 %s 失败", manifest.ID)
	}

	return s.ensurePluginEnabledDuringStartup(ctx, manifest.ID, func() error {
		if err := s.Install(ctx, manifest.ID, nil); err != nil {
			return gerror.Wrap(err, "安装动态插件失败")
		}
		if err := s.Enable(ctx, manifest.ID); err != nil {
			return gerror.Wrap(err, "启用动态插件失败")
		}
		return nil
	})
}

// ensureDynamicPluginAutoEnableAuthorization verifies that startup auto-enable
// can reuse one already confirmed host-service authorization snapshot instead
// of requesting authorization details from the host main config file.
func (s *serviceImpl) ensureDynamicPluginAutoEnableAuthorization(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return gerror.New("动态插件清单不能为空")
	}
	if !catalog.HasResourceScopedHostServices(manifest.HostServices) {
		return nil
	}

	release, err := s.catalogSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("动态插件 %s 缺少发布记录，无法复用授权快照", manifest.ID)
	}

	snapshot, err := s.catalogSvc.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		return err
	}
	if snapshot == nil || !snapshot.HostServiceAuthConfirmed {
		return gerror.Newf(
			"动态插件 %s 缺少已确认的宿主服务授权快照，请先通过常规安装或启用流程完成审核",
			manifest.ID,
		)
	}
	return nil
}

// ensurePluginEnabledDuringStartup waits for one plugin to reach the enabled
// state. The current node performs the shared lifecycle action once it becomes
// primary; otherwise it keeps waiting for the shared registry state to converge.
func (s *serviceImpl) ensurePluginEnabledDuringStartup(
	ctx context.Context,
	pluginID string,
	executeShared func() error,
) error {
	var (
		deadline = time.Now().Add(startupAutoEnableWaitTimeout)
		ticker   = time.NewTicker(startupAutoEnablePollInterval)
	)
	defer ticker.Stop()

	sharedExecuted := false

	for {
		registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
		if err != nil {
			return gerror.Wrapf(err, "读取插件 %s 注册表失败", pluginID)
		}
		if isPluginStartupEnabled(registry) {
			return nil
		}

		if !sharedExecuted && (!s.topology.IsEnabled() || s.topology.IsPrimary()) {
			sharedExecuted = true
			if executeShared == nil {
				return gerror.Newf("启动期自动启用插件 %s 失败：缺少共享执行器", pluginID)
			}
			if err := executeShared(); err != nil {
				return gerror.Wrapf(err, "启动期自动启用插件 %s 失败", pluginID)
			}
			continue
		}

		if time.Now().After(deadline) {
			return buildStartupAutoEnableTimeoutError(pluginID, registry)
		}

		select {
		case <-ctx.Done():
			return gerror.Wrapf(ctx.Err(), "启动期等待插件 %s 自动启用被取消", pluginID)
		case <-ticker.C:
		}
	}
}

// isPluginStartupEnabled reports whether one registry row already reflects the
// stable installed-and-enabled state expected by plugin.autoEnable.
func isPluginStartupEnabled(registry *entity.SysPlugin) bool {
	if registry == nil {
		return false
	}
	if registry.Installed != catalog.InstalledYes || registry.Status != catalog.StatusEnabled {
		return false
	}
	if catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return true
	}
	return strings.TrimSpace(registry.CurrentState) == catalog.HostStateEnabled.String()
}

// buildStartupAutoEnableTimeoutError formats one fail-fast timeout error with
// the last observed registry state so operators can identify the stuck phase.
func buildStartupAutoEnableTimeoutError(pluginID string, registry *entity.SysPlugin) error {
	if registry == nil {
		return gerror.Newf("启动期自动启用插件 %s 超时：插件注册表不存在", pluginID)
	}
	return gerror.Newf(
		"启动期自动启用插件 %s 超时：installed=%d status=%d desiredState=%s currentState=%s",
		pluginID,
		registry.Installed,
		registry.Status,
		strings.TrimSpace(registry.DesiredState),
		strings.TrimSpace(registry.CurrentState),
	)
}
