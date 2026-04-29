// This file coordinates startup-time plugin bootstrap so plugin.autoEnable can
// install and enable required plugins before later host wiring runs.

package plugin

import (
	"context"
	"strings"
	"time"

	"lina-core/internal/model/entity"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
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
// Per-entry mock-data opt-in flags from config flow into the InstallOptions
// passed down to Install.
func (s *serviceImpl) BootstrapAutoEnable(ctx context.Context) error {
	if err := s.SyncSourcePlugins(ctx); err != nil {
		return err
	}

	entries := s.configSvc.GetPluginAutoEnableEntries(ctx)
	if len(entries) == 0 {
		return nil
	}

	for _, entry := range entries {
		if err := s.bootstrapAutoEnablePlugin(ctx, entry); err != nil {
			return err
		}
	}

	if err := s.integrationSvc.RefreshEnabledSnapshot(ctx); err != nil {
		return bizerr.WrapCode(err, CodePluginEnabledSnapshotRefreshFailed)
	}
	return nil
}

// bootstrapAutoEnablePlugin routes one configured plugin entry into the
// matching source-plugin or dynamic-plugin startup bootstrap path. The entry
// carries both the ID and the per-plugin mock-data opt-in flag.
func (s *serviceImpl) bootstrapAutoEnablePlugin(ctx context.Context, entry configsvc.PluginAutoEnableEntry) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(entry.ID)
	if err != nil {
		return bizerr.WrapCode(err, CodePluginAutoEnableDiscoveryFailed, bizerr.P("pluginId", entry.ID))
	}
	if manifest == nil {
		return bizerr.NewCode(CodePluginAutoEnableManifestNotFound, bizerr.P("pluginId", entry.ID))
	}

	switch catalog.NormalizeType(manifest.Type) {
	case catalog.TypeSource:
		return s.bootstrapAutoEnableSourcePlugin(ctx, manifest, entry.WithMockData)
	case catalog.TypeDynamic:
		return s.bootstrapAutoEnableDynamicPlugin(ctx, manifest, entry.WithMockData)
	default:
		return bizerr.NewCode(
			CodePluginAutoEnableTypeUnsupported,
			bizerr.P("pluginId", entry.ID),
			bizerr.P("pluginType", manifest.Type),
		)
	}
}

// bootstrapAutoEnableSourcePlugin ensures one required source plugin reaches
// the enabled state during startup. When withMockData is true and the plugin
// is not yet installed, the install call also loads the plugin's mock-data
// SQL inside one transaction. Already-installed plugins do not re-run the
// mock-data load even if withMockData=true, since mock data is install-time
// only.
func (s *serviceImpl) bootstrapAutoEnableSourcePlugin(ctx context.Context, manifest *catalog.Manifest, withMockData bool) error {
	if manifest == nil {
		return bizerr.NewCode(CodePluginAutoEnableSourceManifestRequired)
	}

	return s.ensurePluginEnabledDuringStartup(ctx, manifest.ID, func() error {
		if err := s.Install(ctx, manifest.ID, InstallOptions{InstallMockData: withMockData}); err != nil {
			return bizerr.WrapCode(err, CodePluginSourceInstallFailed)
		}
		if err := s.Enable(ctx, manifest.ID); err != nil {
			return bizerr.WrapCode(err, CodePluginSourceEnableFailed)
		}
		return nil
	})
}

// bootstrapAutoEnableDynamicPlugin ensures one required dynamic plugin can
// reuse its confirmed authorization snapshot and then reaches the enabled state
// during startup. The mock-data opt-in flag flows through InstallOptions just
// like the source-plugin path.
func (s *serviceImpl) bootstrapAutoEnableDynamicPlugin(ctx context.Context, manifest *catalog.Manifest, withMockData bool) error {
	if manifest == nil {
		return bizerr.NewCode(CodePluginAutoEnableDynamicManifestRequired)
	}
	if err := s.ensureDynamicPluginAutoEnableAuthorization(ctx, manifest); err != nil {
		return bizerr.WrapCode(err, CodePluginAutoEnableFailed, bizerr.P("pluginId", manifest.ID))
	}

	return s.ensurePluginEnabledDuringStartup(ctx, manifest.ID, func() error {
		if err := s.Install(ctx, manifest.ID, InstallOptions{InstallMockData: withMockData}); err != nil {
			return bizerr.WrapCode(err, CodePluginDynamicInstallFailed)
		}
		if err := s.Enable(ctx, manifest.ID); err != nil {
			return bizerr.WrapCode(err, CodePluginDynamicEnableFailed)
		}
		return nil
	})
}

// ensureDynamicPluginAutoEnableAuthorization verifies that startup auto-enable
// can reuse one already confirmed host-service authorization snapshot instead
// of requesting authorization details from the host main config file.
func (s *serviceImpl) ensureDynamicPluginAutoEnableAuthorization(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return bizerr.NewCode(CodePluginDynamicManifestRequired)
	}
	if !catalog.HasResourceScopedHostServices(manifest.HostServices) {
		return nil
	}

	release, err := s.catalogSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return bizerr.NewCode(CodePluginDynamicAutoEnableReleaseMissing, bizerr.P("pluginId", manifest.ID))
	}

	snapshot, err := s.catalogSvc.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		return err
	}
	if snapshot == nil || !snapshot.HostServiceAuthConfirmed {
		return bizerr.NewCode(CodePluginDynamicAutoEnableAuthSnapshotMissing, bizerr.P("pluginId", manifest.ID))
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
			return bizerr.WrapCode(err, CodePluginRegistryReadFailed, bizerr.P("pluginId", pluginID))
		}
		if isPluginStartupEnabled(registry) {
			return nil
		}

		if !sharedExecuted && (!s.topology.IsEnabled() || s.topology.IsPrimary()) {
			sharedExecuted = true
			if executeShared == nil {
				return bizerr.NewCode(CodePluginAutoEnableSharedExecutorMissing, bizerr.P("pluginId", pluginID))
			}
			if err := executeShared(); err != nil {
				return bizerr.WrapCode(err, CodePluginAutoEnableFailed, bizerr.P("pluginId", pluginID))
			}
			continue
		}

		if time.Now().After(deadline) {
			return buildStartupAutoEnableTimeoutError(pluginID, registry)
		}

		select {
		case <-ctx.Done():
			return bizerr.WrapCode(ctx.Err(), CodePluginAutoEnableWaitCanceled, bizerr.P("pluginId", pluginID))
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
		return bizerr.NewCode(CodePluginAutoEnableTimeoutRegistryMissing, bizerr.P("pluginId", pluginID))
	}
	return bizerr.NewCode(
		CodePluginAutoEnableTimeoutState,
		bizerr.P("pluginId", pluginID),
		bizerr.P("installed", registry.Installed),
		bizerr.P("status", registry.Status),
		bizerr.P("desiredState", strings.TrimSpace(registry.DesiredState)),
		bizerr.P("currentState", strings.TrimSpace(registry.CurrentState)),
	)
}
