// This file covers root-facade lifecycle methods defined in plugin_lifecycle.go.

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginhost"
)

// TestUpdateStatusEnablesBackendOnlyDynamicPluginWithoutFrontendAssets verifies
// that route-only runtime plugins can be enabled without bundled frontend files.
func TestUpdateStatusEnablesBackendOnlyDynamicPluginWithoutFrontendAssets(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-backend-only"
	)

	frontend.ResetBundleCache()
	t.Cleanup(frontend.ResetBundleCache)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		"Backend Only Dynamic Plugin",
		"v0.4.1",
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected backend-only artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected backend-only artifact sync to succeed, got error: %v", err)
	}
	if err = service.setPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected backend-only plugin install state to be set, got error: %v", err)
	}

	if err = service.UpdateStatus(ctx, pluginID, catalog.StatusEnabled, nil); err != nil {
		t.Fatalf("expected backend-only dynamic plugin enable to succeed, got error: %v", err)
	}
	if !service.IsEnabled(ctx, pluginID) {
		t.Fatalf("expected backend-only dynamic plugin to be enabled after status update")
	}
}

// TestApplyInstallModeSelectionRejectsInvalidMode verifies service-layer install
// validation rejects unsupported install-mode values before registry sync.
func TestApplyInstallModeSelectionRejectsInvalidMode(t *testing.T) {
	manifest := &catalog.Manifest{
		ID:                 "plugin-invalid-install-mode",
		ScopeNature:        catalog.ScopeNatureTenantAware.String(),
		DefaultInstallMode: catalog.InstallModeTenantScoped.String(),
	}

	err := applyInstallModeSelection(manifest, "per_tenant")
	if !bizerr.Is(err, CodePluginInstallModeInvalid) {
		t.Fatalf("expected invalid install mode bizerr, got %v", err)
	}
}

// TestApplyInstallModeSelectionRejectsPlatformOnlyTenantScoped verifies
// platform-only plugins cannot be installed with tenant-scoped enablement.
func TestApplyInstallModeSelectionRejectsPlatformOnlyTenantScoped(t *testing.T) {
	manifest := &catalog.Manifest{
		ID:                 "plugin-platform-only-install-mode",
		ScopeNature:        catalog.ScopeNaturePlatformOnly.String(),
		DefaultInstallMode: catalog.InstallModeGlobal.String(),
	}

	err := applyInstallModeSelection(manifest, catalog.InstallModeTenantScoped.String())
	if !bizerr.Is(err, CodePluginInstallModeInvalidForScopeNature) {
		t.Fatalf("expected scope/install-mode mismatch bizerr, got %v", err)
	}
}

// TestApplyInstallModeSelectionPersistsExplicitTenantAwareMode verifies an
// explicit platform selection overrides the manifest default before install.
func TestApplyInstallModeSelectionPersistsExplicitTenantAwareMode(t *testing.T) {
	manifest := &catalog.Manifest{
		ID:                 "plugin-tenant-aware-install-mode",
		ScopeNature:        catalog.ScopeNatureTenantAware.String(),
		DefaultInstallMode: catalog.InstallModeTenantScoped.String(),
	}

	if err := applyInstallModeSelection(manifest, catalog.InstallModeGlobal.String()); err != nil {
		t.Fatalf("expected explicit global install mode to be accepted, got %v", err)
	}
	if manifest.DefaultInstallMode != catalog.InstallModeGlobal.String() {
		t.Fatalf("expected explicit global install mode to be applied, got %s", manifest.DefaultInstallMode)
	}
}

// TestApplyInstallModeSelectionRejectsUnsupportedTenantScoped verifies explicit
// manifest opt-out from tenant governance also rejects tenant-scoped install.
func TestApplyInstallModeSelectionRejectsUnsupportedTenantScoped(t *testing.T) {
	supportsMultiTenant := false
	manifest := &catalog.Manifest{
		ID:                  "plugin-tenant-unsupported-install-mode",
		ScopeNature:         catalog.ScopeNatureTenantAware.String(),
		SupportsMultiTenant: &supportsMultiTenant,
		DefaultInstallMode:  catalog.InstallModeGlobal.String(),
	}

	err := applyInstallModeSelection(manifest, catalog.InstallModeTenantScoped.String())
	if !bizerr.Is(err, CodePluginInstallModeInvalidForScopeNature) {
		t.Fatalf("expected unsupported tenant-scoped install mode bizerr, got %v", err)
	}
	if manifest.DefaultInstallMode != catalog.InstallModeGlobal.String() {
		t.Fatalf("expected unsupported tenant governance to keep global install mode, got %s", manifest.DefaultInstallMode)
	}
}

// TestInstallPersistsExplicitDynamicInstallMode verifies dynamic install does
// not let the runtime lifecycle's manifest reload reset the operator selection.
func TestInstallPersistsExplicitDynamicInstallMode(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-explicit-install-mode"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		"Dynamic Explicit Install Mode",
		"v0.4.2",
		nil,
		nil,
		nil,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	if err := service.Install(ctx, pluginID, InstallOptions{
		InstallMode: catalog.InstallModeGlobal.String(),
	}); err != nil {
		t.Fatalf("install dynamic plugin with explicit global mode: %v", err)
	}
	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("load plugin registry after install: %v", err)
	}
	if registry == nil {
		t.Fatal("expected plugin registry after install")
	}
	if registry.InstallMode != catalog.InstallModeGlobal.String() {
		t.Fatalf("expected explicit global install_mode to persist, got %s", registry.InstallMode)
	}
}

// TestUninstallDynamicUsesArchivedReleaseWhenStagingArtifactMissing verifies
// uninstall relies on the active release archive instead of the mutable staging
// artifact after a dynamic plugin has been installed.
func TestUninstallDynamicUsesArchivedReleaseWhenStagingArtifactMissing(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-uninstall-missing-staging"
		menuKey  = "plugin:plugin-dynamic-uninstall-missing-staging:entry"
		version  = "v0.4.3"
	)

	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	releaseRoot := filepath.Join(testutil.TestDynamicStorageDir(), "releases", pluginID)
	if err := os.RemoveAll(releaseRoot); err != nil {
		t.Fatalf("failed to clear release archive root: %v", err)
	}
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if err := os.RemoveAll(releaseRoot); err != nil {
			t.Fatalf("failed to cleanup release archive root: %v", err)
		}
	})

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithMenus(
		t,
		pluginID,
		"Dynamic Missing Staging Uninstall Plugin",
		version,
		[]*catalog.MenuSpec{
			{
				Key:   menuKey,
				Name:  "Dynamic Missing Staging Uninstall Plugin",
				Path:  "plugin-dynamic-uninstall-missing-staging",
				Perms: "plugin-dynamic-uninstall-missing-staging:view",
				Icon:  "ant-design:appstore-outlined",
				Type:  catalog.MenuTypePage.String(),
				Sort:  1,
			},
		},
		nil,
		nil,
	)

	if err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}
	release, err := service.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		t.Fatalf("expected dynamic plugin release lookup to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected dynamic plugin release row after install")
	}
	archivePath := resolveTestDynamicPackagePath(t, release.PackagePath)
	if _, err = os.Stat(archivePath); err != nil {
		t.Fatalf("expected active release archive to exist before staging removal: %v", err)
	}
	if err = os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove staging artifact: %v", err)
	}

	if err = service.Uninstall(ctx, pluginID); err != nil {
		t.Fatalf("expected dynamic plugin uninstall to use archived release, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected plugin registry lookup after uninstall to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected plugin registry row after uninstall")
	}
	if registry.Installed != catalog.InstalledNo || registry.Status != catalog.StatusDisabled || registry.ReleaseId != 0 {
		t.Fatalf("expected dynamic plugin to be uninstalled+disabled with release cleared, got installed=%d enabled=%d releaseID=%d", registry.Installed, registry.Status, registry.ReleaseId)
	}
	menu, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected plugin menu lookup after uninstall to succeed, got error: %v", err)
	}
	if menu != nil {
		t.Fatalf("expected dynamic plugin menu to be removed after uninstall")
	}
}

// TestUninstallForceClearsDynamicOrphanWhenArtifactsMissing verifies force
// uninstall can clear host governance when both staging and release artifacts
// are unavailable.
func TestUninstallForceClearsDynamicOrphanWhenArtifactsMissing(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-force-orphan-uninstall"
		menuKey  = "plugin:plugin-dynamic-force-orphan-uninstall:entry"
		version  = "v0.4.4"
	)

	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	releaseRoot := filepath.Join(testutil.TestDynamicStorageDir(), "releases", pluginID)
	if err := os.RemoveAll(releaseRoot); err != nil {
		t.Fatalf("failed to clear release archive root: %v", err)
	}
	t.Cleanup(func() {
		configsvc.SetPluginAllowForceUninstallOverride(nil)
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if err := os.RemoveAll(releaseRoot); err != nil {
			t.Fatalf("failed to cleanup release archive root: %v", err)
		}
	})

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithMenus(
		t,
		pluginID,
		"Dynamic Force Orphan Uninstall Plugin",
		version,
		[]*catalog.MenuSpec{
			{
				Key:   menuKey,
				Name:  "Dynamic Force Orphan Uninstall Plugin",
				Path:  "plugin-dynamic-force-orphan-uninstall",
				Perms: "plugin-dynamic-force-orphan-uninstall:view",
				Icon:  "ant-design:appstore-outlined",
				Type:  catalog.MenuTypePage.String(),
				Sort:  1,
			},
		},
		nil,
		nil,
	)

	if err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}
	release, err := service.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		t.Fatalf("expected dynamic plugin release lookup to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected dynamic plugin release row after install")
	}
	archivePath := resolveTestDynamicPackagePath(t, release.PackagePath)
	if err = os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove staging artifact: %v", err)
	}
	if err = os.Remove(archivePath); err != nil {
		t.Fatalf("failed to remove active release artifact: %v", err)
	}

	resourceCount, err := dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{PluginId: pluginID}).
		Count()
	if err != nil {
		t.Fatalf("expected governance resource count query to succeed, got error: %v", err)
	}
	if resourceCount == 0 {
		t.Fatalf("expected installed dynamic plugin to materialize governance resource refs")
	}

	err = service.UninstallWithOptions(ctx, pluginID, UninstallOptions{PurgeStorageData: true})
	if !bizerr.Is(err, CodePluginDynamicArtifactMissingForUninstall) {
		t.Fatalf("expected missing-artifact uninstall bizerr, got %v", err)
	}

	enabled := true
	configsvc.SetPluginAllowForceUninstallOverride(&enabled)
	if err = service.UninstallWithOptions(ctx, pluginID, UninstallOptions{
		PurgeStorageData: true,
		Force:            true,
	}); err != nil {
		t.Fatalf("expected force orphan uninstall to succeed, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected plugin registry lookup after force uninstall to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected plugin registry row after force uninstall")
	}
	if registry.Installed != catalog.InstalledNo || registry.Status != catalog.StatusDisabled || registry.ReleaseId != 0 {
		t.Fatalf("expected force orphan uninstall to clear runtime state, got installed=%d enabled=%d releaseID=%d", registry.Installed, registry.Status, registry.ReleaseId)
	}
	if registry.DesiredState != catalog.HostStateUninstalled.String() || registry.CurrentState != catalog.HostStateUninstalled.String() {
		t.Fatalf("expected force orphan uninstall to mark host states uninstalled, got desired=%s current=%s", registry.DesiredState, registry.CurrentState)
	}

	release, err = service.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		t.Fatalf("expected dynamic plugin release lookup after force uninstall to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected dynamic plugin release row after force uninstall")
	}
	if release.Status != catalog.ReleaseStatusUninstalled.String() {
		t.Fatalf("expected release to be marked uninstalled after force orphan uninstall, got %s", release.Status)
	}

	menu, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected plugin menu lookup after force uninstall to succeed, got error: %v", err)
	}
	if menu != nil {
		t.Fatalf("expected force orphan uninstall to remove plugin menu")
	}
	resourceCount, err = dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{PluginId: pluginID}).
		Count()
	if err != nil {
		t.Fatalf("expected governance resource count query after force uninstall to succeed, got error: %v", err)
	}
	if resourceCount != 0 {
		t.Fatalf("expected force orphan uninstall to clear governance resource refs, got count=%d", resourceCount)
	}
}

// resolveTestDynamicPackagePath resolves a release package path inside the
// shared dynamic-plugin test storage directory.
func resolveTestDynamicPackagePath(t *testing.T, packagePath string) string {
	t.Helper()

	if filepath.IsAbs(packagePath) {
		return filepath.Clean(packagePath)
	}
	return filepath.Join(testutil.TestDynamicStorageDir(), filepath.FromSlash(packagePath))
}

// TestDisableRunsLifecycleGuards verifies disable requests fail closed when a
// lifecycle guard vetoes the operation.
func TestDisableRunsLifecycleGuards(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-disable-guarded"
	)
	pluginhost.RegisterLifecycleGuard(pluginID, lifecycleGuardDisableVeto{})
	t.Cleanup(func() {
		pluginhost.UnregisterLifecycleGuard(pluginID)
	})

	err := service.ensureLifecycleGuardAllowed(ctx, pluginID, pluginhost.GuardHookCanDisable, false)
	if !bizerr.Is(err, CodePluginLifecycleGuardVetoed) {
		t.Fatalf("expected lifecycle guard bizerr, got %v", err)
	}
}

// TestDisableIgnoresNonTargetLifecycleGuards verifies one plugin-owned guard
// cannot veto lifecycle actions for unrelated plugins.
func TestDisableIgnoresNonTargetLifecycleGuards(t *testing.T) {
	var (
		service        = newTestService()
		ctx            = context.Background()
		targetPluginID = "plugin-disable-target"
		guardPluginID  = "plugin-disable-other-guard"
	)
	pluginhost.RegisterLifecycleGuard(guardPluginID, lifecycleGuardDisableVeto{})
	t.Cleanup(func() {
		pluginhost.UnregisterLifecycleGuard(guardPluginID)
	})

	err := service.ensureLifecycleGuardAllowed(ctx, targetPluginID, pluginhost.GuardHookCanDisable, false)
	if err != nil {
		t.Fatalf("expected unrelated lifecycle guard to be ignored, got %v", err)
	}
}

// TestTenantDeleteRunsLifecycleGuards verifies tenant deletion fails closed
// when any plugin-owned lifecycle guard vetoes the tenant delete hook.
func TestTenantDeleteRunsLifecycleGuards(t *testing.T) {
	var (
		service = newTestService()
		ctx     = context.Background()
		guardID = "plugin-tenant-delete-guard"
	)
	pluginhost.RegisterLifecycleGuard(guardID, lifecycleGuardTenantDeleteVeto{})
	t.Cleanup(func() {
		pluginhost.UnregisterLifecycleGuard(guardID)
	})

	err := service.EnsureTenantDeleteAllowed(ctx, 8001)
	if !bizerr.Is(err, CodePluginLifecycleGuardVetoed) {
		t.Fatalf("expected tenant delete lifecycle guard bizerr, got %v", err)
	}
}

// TestTenantDeleteLifecycleGuardAllowsWhenNoParticipant verifies tenant
// deletion guard checks are a no-op when no plugin participates.
func TestTenantDeleteLifecycleGuardAllowsWhenNoParticipant(t *testing.T) {
	service := newTestService()

	if err := service.EnsureTenantDeleteAllowed(context.Background(), 8002); err != nil {
		t.Fatalf("expected tenant delete guard to allow without participants, got %v", err)
	}
}

// TestUninstallForceRequiresConfig verifies force uninstall is gated by
// plugin.allowForceUninstall before bypassing guard vetoes.
func TestUninstallForceRequiresConfig(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-force-guarded"
	)
	pluginhost.RegisterLifecycleGuard(pluginID, lifecycleGuardUninstallVeto{})
	t.Cleanup(func() {
		pluginhost.UnregisterLifecycleGuard(pluginID)
		configsvc.SetPluginAllowForceUninstallOverride(nil)
	})

	disabled := false
	configsvc.SetPluginAllowForceUninstallOverride(&disabled)
	err := service.ensureLifecycleGuardAllowed(ctx, pluginID, pluginhost.GuardHookCanUninstall, true)
	if !bizerr.Is(err, CodePluginForceUninstallDisabled) {
		t.Fatalf("expected force-disabled bizerr, got %v", err)
	}
}

// TestUninstallForceBypassesLifecycleGuardWhenConfigured verifies force
// uninstall can bypass guard vetoes when the host config explicitly allows it.
func TestUninstallForceBypassesLifecycleGuardWhenConfigured(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-force-missing-after-guard"
	)
	pluginhost.RegisterLifecycleGuard(pluginID, lifecycleGuardUninstallVeto{})
	t.Cleanup(func() {
		pluginhost.UnregisterLifecycleGuard(pluginID)
		configsvc.SetPluginAllowForceUninstallOverride(nil)
	})

	enabled := true
	configsvc.SetPluginAllowForceUninstallOverride(&enabled)
	err := service.ensureLifecycleGuardAllowed(ctx, pluginID, pluginhost.GuardHookCanUninstall, true)
	if bizerr.Is(err, CodePluginLifecycleGuardVetoed) || bizerr.Is(err, CodePluginForceUninstallDisabled) {
		t.Fatalf("expected force to bypass guard errors, got %v", err)
	}
	if err != nil {
		t.Fatalf("expected force bypass to continue, got %v", err)
	}
}

// lifecycleGuardDisableVeto is a test guard that blocks disable.
type lifecycleGuardDisableVeto struct{}

// CanDisable returns a deterministic test veto reason.
func (lifecycleGuardDisableVeto) CanDisable(ctx context.Context) (bool, string, error) {
	return false, "plugin.test.disable_blocked", nil
}

// lifecycleGuardUninstallVeto is a test guard that blocks uninstall.
type lifecycleGuardUninstallVeto struct{}

// CanUninstall returns a deterministic test veto reason.
func (lifecycleGuardUninstallVeto) CanUninstall(ctx context.Context) (bool, string, error) {
	return false, "plugin.test.uninstall_blocked", nil
}

// lifecycleGuardTenantDeleteVeto is a test guard that blocks tenant deletion.
type lifecycleGuardTenantDeleteVeto struct{}

// CanTenantDelete returns a deterministic tenant-delete veto reason.
func (lifecycleGuardTenantDeleteVeto) CanTenantDelete(ctx context.Context, tenantID int) (bool, string, error) {
	return false, "plugin.test.tenant_delete_blocked", nil
}

// TestSyncAndListReportsPendingHostServiceAuthorization verifies that list
// projections expose dynamic plugin authorization review requirements.
func TestSyncAndListReportsPendingHostServiceAuthorization(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-host-auth-pending"
	)

	artifactPath := filepath.Join(
		testutil.TestDynamicStorageDir(),
		pluginID+".wasm",
	)
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Pending Authorization Plugin",
			Version: "v0.5.0",
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: pluginbridge.RuntimeKindWasm,
			ABIVersion:  pluginbridge.SupportedABIVersion,
			HostServices: []*pluginbridge.HostServiceSpec{
				{
					Service: pluginbridge.HostServiceRuntime,
					Methods: []string{pluginbridge.HostServiceMethodRuntimeInfoNow},
				},
				{
					Service: pluginbridge.HostServiceNetwork,
					Methods: []string{pluginbridge.HostServiceMethodNetworkRequest},
					Resources: []*pluginbridge.HostServiceResourceSpec{
						{
							Ref: "https://example.com/api",
						},
					},
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	out, err := service.SyncAndList(ctx)
	if err != nil {
		t.Fatalf("expected sync-and-list to succeed, got error: %v", err)
	}

	var item *PluginItem
	for _, current := range out.List {
		if current != nil && current.Id == pluginID {
			item = current
			break
		}
	}
	if item == nil {
		t.Fatalf("expected pending authorization plugin in list")
	}
	if !item.AuthorizationRequired {
		t.Fatalf("expected pending authorization plugin to require review")
	}
	if item.AuthorizationStatus != runtime.AuthorizationStatusPending {
		t.Fatalf("expected authorization status pending, got %s", item.AuthorizationStatus)
	}
	if len(item.RequestedHostServices) != 2 {
		t.Fatalf("expected requested host services to be exposed, got %#v", item.RequestedHostServices)
	}
	if len(item.AuthorizedHostServices) != 0 {
		t.Fatalf("expected no authorized host services before confirmation, got %#v", item.AuthorizedHostServices)
	}
}

// TestEnableWithAuthorizationAppliesConfirmedHostServiceSnapshot verifies that
// install and enable persist the host-confirmed authorization snapshot.
func TestEnableWithAuthorizationAppliesConfirmedHostServiceSnapshot(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-host-auth-enabled"
	)

	artifactPath := filepath.Join(
		testutil.TestDynamicStorageDir(),
		pluginID+".wasm",
	)
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Confirmed Authorization Plugin",
			Version: "v0.5.1",
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: pluginbridge.RuntimeKindWasm,
			ABIVersion:  pluginbridge.SupportedABIVersion,
			HostServices: []*pluginbridge.HostServiceSpec{
				{
					Service: pluginbridge.HostServiceRuntime,
					Methods: []string{pluginbridge.HostServiceMethodRuntimeInfoNow},
				},
				{
					Service: pluginbridge.HostServiceNetwork,
					Methods: []string{pluginbridge.HostServiceMethodNetworkRequest},
					Resources: []*pluginbridge.HostServiceResourceSpec{
						{
							Ref: "https://example.com/api",
						},
					},
				},
				{
					Service: pluginbridge.HostServiceStorage,
					Methods: []string{pluginbridge.HostServiceMethodStorageGet},
					Paths:   []string{"private-files/"},
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	authorization := &HostServiceAuthorizationInput{
		Services: []*HostServiceAuthorizationDecision{
			{
				Service: pluginbridge.HostServiceStorage,
				Paths:   []string{"private-files/"},
			},
		},
	}

	if err := service.Install(ctx, pluginID, InstallOptions{Authorization: authorization}); err != nil {
		t.Fatalf("expected install with authorization to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, pluginID, catalog.StatusEnabled, authorization); err != nil {
		t.Fatalf("expected enable with authorization to succeed, got error: %v", err)
	}

	release, err := service.getPluginRelease(ctx, pluginID, "v0.5.1")
	if err != nil {
		t.Fatalf("expected release lookup to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected release row after enable")
	}

	snapshot, err := service.catalogSvc.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		t.Fatalf("expected manifest snapshot parse to succeed, got error: %v", err)
	}
	if snapshot == nil || !snapshot.HostServiceAuthConfirmed {
		t.Fatalf("expected confirmed host service authorization snapshot, got %#v", snapshot)
	}

	activeManifest, err := service.getActivePluginManifest(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected active manifest lookup to succeed, got error: %v", err)
	}
	if activeManifest == nil {
		t.Fatalf("expected active manifest after enable")
	}
	if len(activeManifest.HostServices) != 2 {
		t.Fatalf("expected active manifest to use narrowed host services, got %#v", activeManifest.HostServices)
	}
	if activeManifest.HostServices[0].Service != pluginbridge.HostServiceRuntime &&
		activeManifest.HostServices[1].Service != pluginbridge.HostServiceRuntime {
		t.Fatalf("expected runtime host service to remain authorized, got %#v", activeManifest.HostServices)
	}
	for _, spec := range activeManifest.HostServices {
		if spec == nil {
			continue
		}
		if spec.Service == pluginbridge.HostServiceNetwork {
			t.Fatalf("expected network host service to be removed from authorized snapshot, got %#v", activeManifest.HostServices)
		}
	}
	if _, ok := activeManifest.HostCapabilities[pluginbridge.CapabilityHTTPRequest]; ok {
		t.Fatalf("expected network capability to be removed with rejected authorization")
	}
}

// TestSourcePluginInstallAndUninstallRequireExplicitLifecycle verifies that
// source plugins stay discovered-only until the host explicitly installs them.
func TestSourcePluginInstallAndUninstallRequireExplicitLifecycle(t *testing.T) {
	var (
		service = newTestService()
		ctx     = context.Background()
	)

	const (
		pluginID = "plugin-source-explicit-lifecycle"
		menuKey  = "plugin:plugin-source-explicit-lifecycle:entry"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	testutil.WriteTestFile(
		t,
		manifestPath,
		"id: "+pluginID+"\n"+
			"name: Source Explicit Lifecycle Plugin\n"+
			"version: v0.1.0\n"+
			"type: source\n"+
			"menus:\n"+
			"  - key: "+menuKey+"\n"+
			"    name: Source Explicit Lifecycle Plugin\n"+
			"    path: plugin-source-explicit-lifecycle\n"+
			"    component: system/plugin/dynamic-page\n"+
			"    perms: plugin-source-explicit-lifecycle:view\n"+
			"    icon: ant-design:appstore-outlined\n"+
			"    type: M\n"+
			"    sort: -1\n",
	)

	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected source plugin registry row to exist after discovery")
	}
	if registry.Installed != catalog.InstalledNo || registry.Status != catalog.StatusDisabled {
		t.Fatalf("expected source plugin to stay uninstalled+disabled after discovery, got installed=%d enabled=%d", registry.Installed, registry.Status)
	}

	menu, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected source plugin menu query to succeed, got error: %v", err)
	}
	if menu != nil {
		t.Fatalf("expected source plugin menu to remain absent before install")
	}

	release, err := service.getPluginRelease(ctx, pluginID, "v0.1.0")
	if err != nil {
		t.Fatalf("expected source plugin release lookup after discovery to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected source plugin release row after discovery")
	}
	if release.Status != catalog.ReleaseStatusUninstalled.String() {
		t.Fatalf("expected discovered source plugin release to stay uninstalled, got %s", release.Status)
	}

	if err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}

	registry, err = service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup after install to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected source plugin registry row after install")
	}
	if registry.Installed != catalog.InstalledYes || registry.Status != catalog.StatusDisabled {
		t.Fatalf("expected source plugin install to yield installed+disabled, got installed=%d enabled=%d", registry.Installed, registry.Status)
	}
	if registry.InstalledAt == nil {
		t.Fatalf("expected source plugin install to record installed_at")
	}

	menu, err = testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected source plugin menu query after install to succeed, got error: %v", err)
	}
	if menu == nil {
		t.Fatalf("expected source plugin menu to be created on install")
	}

	release, err = service.getPluginRelease(ctx, pluginID, "v0.1.0")
	if err != nil {
		t.Fatalf("expected source plugin release lookup after install to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected source plugin release row after install")
	}
	if release.Status != catalog.ReleaseStatusInstalled.String() {
		t.Fatalf("expected source plugin release to become installed, got %s", release.Status)
	}

	migrationCount, err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId: pluginID,
			Phase:    catalog.MigrationDirectionInstall.String(),
		}).
		Count()
	if err != nil {
		t.Fatalf("expected source plugin install migration count query to succeed, got error: %v", err)
	}
	if migrationCount != 1 {
		t.Fatalf("expected one source plugin install migration row, got count=%d", migrationCount)
	}

	resourceCount, err := dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{PluginId: pluginID, ReleaseId: release.Id}).
		Count()
	if err != nil {
		t.Fatalf("expected source plugin resource ref count query to succeed, got error: %v", err)
	}
	if resourceCount == 0 {
		t.Fatalf("expected source plugin install to materialize governance resource refs")
	}

	if err = service.Uninstall(ctx, pluginID); err != nil {
		t.Fatalf("expected source plugin uninstall to succeed, got error: %v", err)
	}

	registry, err = service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup after uninstall to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected source plugin registry row after uninstall")
	}
	if registry.Installed != catalog.InstalledNo || registry.Status != catalog.StatusDisabled {
		t.Fatalf("expected source plugin uninstall to yield uninstalled+disabled, got installed=%d enabled=%d", registry.Installed, registry.Status)
	}

	menu, err = testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected source plugin menu query after uninstall to succeed, got error: %v", err)
	}
	if menu != nil {
		t.Fatalf("expected source plugin menu to be removed on uninstall")
	}

	release, err = service.getPluginRelease(ctx, pluginID, "v0.1.0")
	if err != nil {
		t.Fatalf("expected source plugin release lookup after uninstall to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected source plugin release row after uninstall")
	}
	if release.Status != catalog.ReleaseStatusUninstalled.String() {
		t.Fatalf("expected source plugin release to become uninstalled, got %s", release.Status)
	}

	resourceCount, err = dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{PluginId: pluginID, ReleaseId: release.Id}).
		Count()
	if err != nil {
		t.Fatalf("expected source plugin resource ref count query after uninstall to succeed, got error: %v", err)
	}
	if resourceCount != 0 {
		t.Fatalf("expected source plugin uninstall to clear governance resource refs, got count=%d", resourceCount)
	}

	migrationCount, err = dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId: pluginID,
			Phase:    catalog.MigrationDirectionUninstall.String(),
		}).
		Count()
	if err != nil {
		t.Fatalf("expected source plugin uninstall migration count query to succeed, got error: %v", err)
	}
	if migrationCount != 1 {
		t.Fatalf("expected one source plugin uninstall migration row, got count=%d", migrationCount)
	}
}
