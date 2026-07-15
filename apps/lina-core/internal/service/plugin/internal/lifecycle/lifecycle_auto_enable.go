// This file owns startup plugin.autoEnable orchestration inside lifecycle.

package lifecycle

import (
	"context"
	pluginv1 "lina-core/api/plugin/v1"
	"strings"
	"time"

	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/statusflag"
)

// startupAutoEnableWaitTimeout bounds how long host startup waits for one
// required plugin to reach the enabled state before failing fast.
const startupAutoEnableWaitTimeout = 15 * time.Second

// startupAutoEnablePollInterval is the registry polling cadence used while the
// current node waits to become primary or waits for another primary to converge
// one required plugin.
const startupAutoEnablePollInterval = 100 * time.Millisecond

// AutoEnableEntry is one normalized startup auto-enable target.
type AutoEnableEntry struct {
	// ID is the plugin ID configured in plugin.autoEnable.
	ID string
	// WithMockData reports whether startup install should load mock-data SQL.
	WithMockData bool
}

// BootstrapAutoEnableOptions carries startup auto-enable inputs owned outside
// lifecycle, such as host config and framework metadata.
type BootstrapAutoEnableOptions struct {
	// Entries are the normalized plugin.autoEnable entries to converge.
	Entries []AutoEnableEntry
	// FrameworkVersion is the current LinaPro framework version used for
	// dependency compatibility checks.
	FrameworkVersion string
}

// BootstrapAutoEnable ensures every configured plugin is installed and enabled
// before later host startup phases depend on plugin capabilities.
func (s *serviceImpl) BootstrapAutoEnable(ctx context.Context, options BootstrapAutoEnableOptions) error {
	if s == nil {
		return nil
	}
	if err := s.runtimeSvc.RefreshInstalledRuntimePluginReleases(ctx); err != nil {
		return err
	}
	if len(options.Entries) == 0 {
		return nil
	}
	for _, entry := range options.Entries {
		if err := s.bootstrapAutoEnablePlugin(ctx, entry, options.FrameworkVersion); err != nil {
			return err
		}
	}
	if err := s.integrationSvc.RefreshEnabledSnapshot(ctx); err != nil {
		return bizerr.WrapCode(err, CodePluginEnabledSnapshotRefreshFailed)
	}
	return nil
}

// bootstrapAutoEnablePlugin routes one configured plugin entry into the
// matching source-plugin or dynamic-plugin startup bootstrap path.
func (s *serviceImpl) bootstrapAutoEnablePlugin(
	ctx context.Context,
	entry AutoEnableEntry,
	frameworkVersion string,
) error {
	pluginID := strings.TrimSpace(entry.ID)
	if pluginID == "" {
		return nil
	}
	if err := s.checkStartupAutoEnableDependencies(ctx, pluginID, frameworkVersion); err != nil {
		return err
	}

	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return bizerr.WrapCode(err, CodePluginAutoEnableDiscoveryFailed, bizerr.P("pluginId", pluginID))
	}
	if manifest == nil {
		return bizerr.NewCode(CodePluginAutoEnableManifestNotFound, bizerr.P("pluginId", pluginID))
	}

	switch plugintypes.NormalizeType(manifest.Type) {
	case pluginv1.PluginTypeSource:
		return s.bootstrapAutoEnableSourcePlugin(ctx, manifest, entry.WithMockData, frameworkVersion)
	case pluginv1.PluginTypeDynamic:
		return s.bootstrapAutoEnableDynamicPlugin(ctx, manifest, entry.WithMockData, frameworkVersion)
	default:
		return bizerr.NewCode(
			CodePluginAutoEnableTypeUnsupported,
			bizerr.P("pluginId", pluginID),
			bizerr.P("pluginType", manifest.Type),
		)
	}
}

// checkStartupAutoEnableDependencies verifies that configured startup targets
// already have hard plugin dependencies installed and enabled before enable.
func (s *serviceImpl) checkStartupAutoEnableDependencies(
	ctx context.Context,
	pluginID string,
	frameworkVersion string,
) error {
	plan, err := s.resolveEnableDependencies(ctx, pluginID, frameworkVersion)
	if err != nil {
		return err
	}
	if plugindep.HasBlockers(plan.Blockers) {
		return buildDependencyBlockedError(pluginID, plan.Blockers)
	}
	return nil
}

// bootstrapAutoEnableSourcePlugin ensures one required source plugin reaches
// the enabled state during startup.
func (s *serviceImpl) bootstrapAutoEnableSourcePlugin(
	ctx context.Context,
	manifest *catalog.Manifest,
	withMockData bool,
	frameworkVersion string,
) error {
	if manifest == nil {
		return bizerr.NewCode(CodePluginAutoEnableSourceManifestRequired)
	}

	return s.ensurePluginStateDuringStartup(ctx, manifest.ID, isPluginStartupEnabled, func() error {
		if _, err := s.Install(ctx, manifest.ID, InstallOptions{
			InstallMockData:   withMockData,
			StartupAutoEnable: true,
			FrameworkVersion:  frameworkVersion,
		}); err != nil {
			return bizerr.WrapCode(err, CodePluginSourceInstallFailed)
		}
		if err := s.UpdateStatus(ctx, manifest.ID, statusflag.EnabledValue.Int(), UpdateStatusOptions{
			FrameworkVersion: frameworkVersion,
		}); err != nil {
			return bizerr.WrapCode(err, CodePluginSourceEnableFailed)
		}
		return nil
	})
}

// bootstrapAutoEnableDynamicPlugin ensures one required dynamic plugin reuses
// its confirmed authorization snapshot and reaches enabled state during startup.
func (s *serviceImpl) bootstrapAutoEnableDynamicPlugin(
	ctx context.Context,
	manifest *catalog.Manifest,
	withMockData bool,
	frameworkVersion string,
) error {
	if manifest == nil {
		return bizerr.NewCode(CodePluginAutoEnableDynamicManifestRequired)
	}
	if err := s.ensureDynamicPluginAutoEnableAuthorization(ctx, manifest); err != nil {
		return bizerr.WrapCode(err, CodePluginAutoEnableFailed, bizerr.P("pluginId", manifest.ID))
	}

	return s.ensurePluginStateDuringStartup(ctx, manifest.ID, isPluginStartupEnabled, func() error {
		if _, err := s.Install(ctx, manifest.ID, InstallOptions{
			InstallMockData:  withMockData,
			FrameworkVersion: frameworkVersion,
		}); err != nil {
			return bizerr.WrapCode(err, CodePluginDynamicInstallFailed)
		}
		if err := s.UpdateStatus(ctx, manifest.ID, statusflag.EnabledValue.Int(), UpdateStatusOptions{
			FrameworkVersion: frameworkVersion,
		}); err != nil {
			return bizerr.WrapCode(err, CodePluginDynamicEnableFailed)
		}
		return nil
	})
}

// ensureDynamicPluginAutoEnableAuthorization verifies startup auto-enable can
// reuse one already confirmed host-service authorization snapshot.
func (s *serviceImpl) ensureDynamicPluginAutoEnableAuthorization(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return bizerr.NewCode(CodePluginDynamicManifestRequired)
	}
	if !store.HasResourceScopedHostServices(manifest.HostServices) {
		return nil
	}

	release, err := s.storeSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return bizerr.NewCode(CodePluginDynamicAutoEnableReleaseMissing, bizerr.P("pluginId", manifest.ID))
	}

	snapshot, err := s.storeSvc.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		return err
	}
	if snapshot == nil || !snapshot.HostServiceAuthConfirmed {
		return bizerr.NewCode(CodePluginDynamicAutoEnableAuthSnapshotMissing, bizerr.P("pluginId", manifest.ID))
	}
	return nil
}

// ensurePluginStateDuringStartup waits for one plugin to reach a
// caller-defined registry state. The current node performs the shared lifecycle
// action once it becomes primary; otherwise it waits for shared state to converge.
func (s *serviceImpl) ensurePluginStateDuringStartup(
	ctx context.Context,
	pluginID string,
	stateSatisfied func(*store.PluginRecord) bool,
	executeShared func() error,
) error {
	return s.ensurePluginStateDuringStartupWithPolicy(ctx, pluginID, stateSatisfied, executeShared, true)
}

// ensurePluginStateDuringStartupUnwrapped waits like ensurePluginStateDuringStartup
// but returns the shared executor error directly. Builtin startup reconciliation
// uses it so runtime-upgrade diagnostics are not hidden behind auto-enable codes.
func (s *serviceImpl) ensurePluginStateDuringStartupUnwrapped(
	ctx context.Context,
	pluginID string,
	stateSatisfied func(*store.PluginRecord) bool,
	executeShared func() error,
) error {
	return s.ensurePluginStateDuringStartupWithPolicy(ctx, pluginID, stateSatisfied, executeShared, false)
}

// ensurePluginStateDuringStartupWithPolicy implements shared startup waiting
// and primary-only execution for plugin bootstrap flows.
func (s *serviceImpl) ensurePluginStateDuringStartupWithPolicy(
	ctx context.Context,
	pluginID string,
	stateSatisfied func(*store.PluginRecord) bool,
	executeShared func() error,
	wrapSharedError bool,
) error {
	var (
		deadline = time.Now().Add(startupAutoEnableWaitTimeout)
		ticker   = time.NewTicker(startupAutoEnablePollInterval)
	)
	defer ticker.Stop()

	sharedExecuted := false

	for {
		registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
		if err != nil {
			return bizerr.WrapCode(err, CodePluginRegistryReadFailed, bizerr.P("pluginId", pluginID))
		}
		if stateSatisfied != nil && stateSatisfied(registry) {
			return nil
		}

		if !sharedExecuted && (!s.isClusterModeEnabled() || s.isPrimaryNode()) {
			sharedExecuted = true
			if executeShared == nil {
				return bizerr.NewCode(CodePluginAutoEnableSharedExecutorMissing, bizerr.P("pluginId", pluginID))
			}
			if err := executeShared(); err != nil {
				if !wrapSharedError {
					return err
				}
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

// isClusterModeEnabled is a nil-safe wrapper around startup topology.
func (s *serviceImpl) isClusterModeEnabled() bool {
	return s != nil && s.topology != nil && s.topology.IsEnabled()
}

// isPrimaryNode is a nil-safe wrapper around startup topology.
func (s *serviceImpl) isPrimaryNode() bool {
	return s == nil || s.topology == nil || s.topology.IsPrimary()
}

// isPluginStartupEnabled reports whether one registry row already reflects the
// stable installed-and-enabled state expected by plugin.autoEnable.
func isPluginStartupEnabled(registry *store.PluginRecord) bool {
	if registry == nil {
		return false
	}
	if registry.Installed != statusflag.Installed.Int() || registry.Status != statusflag.EnabledValue.Int() {
		return false
	}
	if plugintypes.NormalizeType(registry.Type) != pluginv1.PluginTypeDynamic {
		return true
	}
	return strings.TrimSpace(registry.CurrentState) == plugintypes.HostStateEnabled.String()
}

// buildStartupAutoEnableTimeoutError formats one fail-fast timeout error with
// the last observed registry state so operators can identify the stuck phase.
func buildStartupAutoEnableTimeoutError(pluginID string, registry *store.PluginRecord) error {
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
