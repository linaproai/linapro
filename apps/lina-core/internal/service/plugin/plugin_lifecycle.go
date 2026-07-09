// This file exposes lifecycle and status methods on the root plugin facade.

package plugin

import (
	"context"
	"errors"

	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/pkg/bizerr"
)

// Install executes the install lifecycle and returns the dependency plan/result
// generated before target plugin side effects. It optionally persists one
// host-confirmed host service authorization snapshot when the target is a
// dynamic plugin.
//
// On a rolled-back mock-data load the plugin is fully installed (registry, menus,
// release state) — only the mock data was reverted. Install returns a stable
// bizerr (CodePluginInstallMockDataFailed) carrying pluginId, failedFile,
// rolledBackFiles, and cause so the caller can render a precise message.
func (s *serviceImpl) Install(
	ctx context.Context,
	pluginID string,
	options InstallOptions,
) (result *DependencyCheckResult, err error) {
	if err = s.ensurePlatformGovernance(ctx); err != nil {
		return nil, err
	}
	if err = s.ensureBuiltinManagementActionAllowed(ctx, pluginID); err != nil {
		return nil, err
	}
	return s.install(ctx, pluginID, options)
}

// install executes plugin install side effects for platform-guarded callers
// and trusted startup reconciliation.
func (s *serviceImpl) install(
	ctx context.Context,
	pluginID string,
	options InstallOptions,
) (result *DependencyCheckResult, err error) {
	defer func() {
		err = wrapMockDataLoadError(err)
	}()
	return s.lifecycleSvc.Install(ctx, pluginID, lifecycle.InstallOptions{
		Authorization:    options.Authorization,
		InstallMode:      options.InstallMode,
		InstallMockData:  options.InstallMockData,
		FrameworkVersion: s.frameworkVersion(ctx),
	})
}

// wrapMockDataLoadError converts a migration.MockDataLoadError into the stable
// user-facing bizerr that carries all parameters into i18n templates. Returns
// the original err unchanged when the chain does not contain a mock-data load
// error so callers can pass through arbitrary install errors safely.
func wrapMockDataLoadError(err error) error {
	if err == nil {
		return nil
	}
	var mockErr *migration.MockDataLoadError
	if !errors.As(err, &mockErr) {
		return err
	}
	causeText := ""
	if mockErr.Cause != nil {
		causeText = mockErr.Cause.Error()
	}
	return bizerr.NewCode(
		CodePluginInstallMockDataFailed,
		bizerr.P("pluginId", mockErr.PluginID),
		bizerr.P("failedFile", mockErr.FailedFile),
		bizerr.P("rolledBackFiles", mockErr.RolledBackFiles),
		bizerr.P("cause", causeText),
	)
}

// Uninstall executes the uninstall lifecycle for an installed plugin using one explicit policy snapshot.
func (s *serviceImpl) Uninstall(
	ctx context.Context,
	pluginID string,
	options UninstallOptions,
) error {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return err
	}
	if err := s.ensureBuiltinManagementActionAllowed(ctx, pluginID); err != nil {
		return err
	}
	return s.lifecycleSvc.Uninstall(ctx, pluginID, lifecycle.UninstallOptions{
		PurgeStorageData:    options.PurgeStorageData,
		Force:               options.Force,
		AllowForceUninstall: s.configSvc.GetPlugin(ctx).AllowForceUninstall,
	})
}

// UpdateStatus updates plugin status, where status is 1=enabled and 0=disabled,
// and optionally persists one host-confirmed host service authorization
// snapshot before enabling a dynamic plugin.
func (s *serviceImpl) UpdateStatus(
	ctx context.Context,
	pluginID string,
	options UpdateStatusOptions,
) error {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return err
	}
	if err := s.ensureBuiltinManagementActionAllowed(ctx, pluginID); err != nil {
		return err
	}
	return s.updateStatus(ctx, pluginID, options.Status, options.Authorization)
}

// updateStatus centralizes enable/disable validation so source and dynamic
// plugins both honor installed-state checks before status transitions.
func (s *serviceImpl) updateStatus(
	ctx context.Context,
	pluginID string,
	status int,
	authorization *HostServiceAuthorizationInput,
) error {
	return s.lifecycleSvc.UpdateStatus(ctx, pluginID, status, lifecycle.UpdateStatusOptions{
		Authorization:    authorization,
		FrameworkVersion: s.frameworkVersion(ctx),
	})
}

// IsInstalled returns whether a plugin is installed.
func (s *serviceImpl) IsInstalled(ctx context.Context, pluginID string) bool {
	installed, err := s.runtimeSvc.CheckIsInstalled(ctx, pluginID)
	return err == nil && installed
}

// IsEnabled returns whether a plugin is enabled.
func (s *serviceImpl) IsEnabled(ctx context.Context, pluginID string) bool {
	s.ensureRuntimeCacheFreshBestEffort(ctx, "is_enabled")
	return s.integrationSvc.CanExposeBusinessEntries(ctx, pluginID)
}

// IsProviderEnabled returns whether pluginID is platform-enabled for framework
// capability provider use.
func (s *serviceImpl) IsProviderEnabled(ctx context.Context, pluginID string) bool {
	s.ensureRuntimeCacheFreshBestEffort(ctx, "provider_enabled")
	return s.integrationSvc.IsProviderEnabled(ctx, pluginID)
}

// IsEnabledAuthoritative returns whether pluginID is installed, enabled, and
// allowed to expose business entries after forcing a persisted registry read
// instead of reusing a process-local platform snapshot.
func (s *serviceImpl) IsEnabledAuthoritative(ctx context.Context, pluginID string) bool {
	ctx = integration.WithAuthoritativeEnablement(ctx)
	s.ensureRuntimeCacheFreshBestEffort(ctx, "is_enabled_authoritative")
	return s.integrationSvc.CanExposeBusinessEntries(ctx, pluginID)
}

// EnsureTenantPluginDisableAllowed runs source and dynamic lifecycle
// preconditions before one tenant loses access to a tenant-scoped plugin.
func (s *serviceImpl) EnsureTenantPluginDisableAllowed(ctx context.Context, pluginID string, tenantID int) error {
	return s.lifecycleSvc.EnsureTenantPluginDisableAllowed(ctx, pluginID, tenantID)
}

// NotifyTenantPluginDisabled runs best-effort source and dynamic lifecycle
// callbacks after one tenant loses access to a tenant-scoped plugin.
func (s *serviceImpl) NotifyTenantPluginDisabled(ctx context.Context, pluginID string, tenantID int) {
	s.lifecycleSvc.NotifyTenantPluginDisabled(ctx, pluginID, tenantID)
}

// EnsureTenantDeleteAllowed runs plugin lifecycle preconditions before tenant
// deletion continues in the tenant capability provider.
func (s *serviceImpl) EnsureTenantDeleteAllowed(ctx context.Context, tenantID int) error {
	return s.lifecycleSvc.EnsureTenantDeleteAllowed(ctx, tenantID)
}

// NotifyTenantDeleted runs best-effort source and dynamic lifecycle callbacks
// after a tenant has been deleted.
func (s *serviceImpl) NotifyTenantDeleted(ctx context.Context, tenantID int) {
	s.lifecycleSvc.NotifyTenantDeleted(ctx, tenantID)
}
