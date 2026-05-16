// This file covers root-facade lifecycle methods defined in plugin_lifecycle.go.

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

	if _, err := service.Install(ctx, pluginID, InstallOptions{
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

// TestBuildLifecycleAuthorizedHostServicesDropsUnconfirmedResources verifies
// lifecycle handlers do not receive resource-scoped host services unless the
// operation carries an explicit host authorization decision.
func TestBuildLifecycleAuthorizedHostServicesDropsUnconfirmedResources(t *testing.T) {
	hostServices := []*pluginbridge.HostServiceSpec{
		{
			Service: pluginbridge.HostServiceRuntime,
			Methods: []string{
				pluginbridge.HostServiceMethodRuntimeLogWrite,
			},
		},
		{
			Service: pluginbridge.HostServiceStorage,
			Methods: []string{
				pluginbridge.HostServiceMethodStorageGet,
			},
			Paths: []string{"private-files/"},
		},
	}

	withoutAuthorization, err := buildLifecycleAuthorizedHostServices(hostServices, nil)
	if err != nil {
		t.Fatalf("expected lifecycle host services to normalize, got error: %v", err)
	}
	if len(withoutAuthorization) != 1 || withoutAuthorization[0].Service != pluginbridge.HostServiceRuntime {
		t.Fatalf("expected only capability host service without authorization, got %#v", withoutAuthorization)
	}

	withAuthorization, err := buildLifecycleAuthorizedHostServices(
		hostServices,
		&HostServiceAuthorizationInput{
			Services: []*HostServiceAuthorizationDecision{
				{
					Service: pluginbridge.HostServiceStorage,
					Paths:   []string{"private-files/"},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("expected lifecycle host service authorization to normalize, got error: %v", err)
	}
	if len(withAuthorization) != 2 {
		t.Fatalf("expected runtime and authorized storage host services, got %#v", withAuthorization)
	}
}

// TestApplyTargetReleaseAuthorizedHostServicesFiltersMissingRelease verifies
// upgrade lifecycle execution does not expose resource-scoped host services
// when no target release authorization snapshot exists yet.
func TestApplyTargetReleaseAuthorizedHostServicesFiltersMissingRelease(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-lifecycle-missing-release-auth"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest := &catalog.Manifest{
		ID:      pluginID,
		Version: "v0.9.1",
		HostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceRuntime,
				Methods: []string{
					pluginbridge.HostServiceMethodRuntimeLogWrite,
				},
			},
			{
				Service: pluginbridge.HostServiceStorage,
				Methods: []string{
					pluginbridge.HostServiceMethodStorageGet,
				},
				Paths: []string{"private-files/"},
			},
		},
	}

	filtered, err := service.applyTargetReleaseAuthorizedHostServices(ctx, manifest, nil)
	if err != nil {
		t.Fatalf("expected missing release authorization to filter cleanly, got error: %v", err)
	}
	if filtered == nil {
		t.Fatal("expected filtered manifest")
	}
	if len(filtered.HostServices) != 1 || filtered.HostServices[0].Service != pluginbridge.HostServiceRuntime {
		t.Fatalf("expected missing release to keep only capability host services, got %#v", filtered.HostServices)
	}
	if _, ok := filtered.HostCapabilities[pluginbridge.CapabilityStorage]; ok {
		t.Fatalf("expected missing release to remove storage capability, got %#v", filtered.HostCapabilities)
	}
}

// TestDynamicLifecycleBeforeInstallFailsClosedBeforeInstall verifies dynamic
// BeforeInstall blocks the install path before registry state changes.
func TestDynamicLifecycleBeforeInstallFailsClosedBeforeInstall(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-before-install-fail-closed"
	)

	artifactPath := createDynamicLifecyclePreconditionArtifact(
		t,
		pluginID,
		"Dynamic Before Install Fail Closed Plugin",
		"v0.4.5",
		pluginbridge.LifecycleOperationBeforeInstall,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	_, err := service.Install(ctx, pluginID, InstallOptions{})
	if !bizerr.Is(err, CodePluginLifecyclePreconditionVetoed) {
		t.Fatalf("expected dynamic BeforeInstall precondition bizerr, got %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after vetoed install to succeed, got error: %v", err)
	}
	if registry == nil {
		return
	}
	if registry.Installed != catalog.InstalledNo || registry.Status != catalog.StatusDisabled {
		t.Fatalf("expected vetoed dynamic install to keep plugin uninstalled, got installed=%d status=%d", registry.Installed, registry.Status)
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

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
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

	if err = service.Uninstall(ctx, pluginID, UninstallOptions{PurgeStorageData: true}); err != nil {
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

// TestDynamicLifecycleBeforeDisableFailsClosedBeforeStatusChange verifies
// dynamic BeforeDisable runs from the active release and blocks status changes.
func TestDynamicLifecycleBeforeDisableFailsClosedBeforeStatusChange(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-before-disable-fail-closed"
	)

	artifactPath := createDynamicLifecyclePreconditionArtifact(
		t,
		pluginID,
		"Dynamic Before Disable Fail Closed Plugin",
		"v0.4.6",
		pluginbridge.LifecycleOperationBeforeDisable,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected dynamic plugin enable to succeed, got error: %v", err)
	}

	err := service.Disable(ctx, pluginID)
	if !bizerr.Is(err, CodePluginLifecyclePreconditionVetoed) {
		t.Fatalf("expected dynamic BeforeDisable precondition bizerr, got %v", err)
	}
	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after vetoed disable to succeed, got error: %v", err)
	}
	if registry == nil || registry.Status != catalog.StatusEnabled {
		t.Fatalf("expected vetoed dynamic disable to keep plugin enabled, got %#v", registry)
	}
}

// TestDynamicLifecycleBeforeUninstallUsesActiveReleaseWhenStagingMissing
// verifies archived dynamic lifecycle handlers still guard uninstall after the
// mutable staging artifact has been removed.
func TestDynamicLifecycleBeforeUninstallUsesActiveReleaseWhenStagingMissing(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-before-uninstall-active"
		version  = "v0.4.7"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	releaseRoot := filepath.Join(testutil.TestDynamicStorageDir(), "releases", pluginID)
	if err := os.RemoveAll(releaseRoot); err != nil {
		t.Fatalf("failed to clear release archive root: %v", err)
	}
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if err := os.RemoveAll(releaseRoot); err != nil {
			t.Fatalf("failed to cleanup release archive root: %v", err)
		}
	})

	artifactPath := createDynamicLifecyclePreconditionArtifact(
		t,
		pluginID,
		"Dynamic Before Uninstall Active Release Plugin",
		version,
		pluginbridge.LifecycleOperationBeforeUninstall,
	)
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}
	if err := os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove staging artifact: %v", err)
	}

	err := service.Uninstall(ctx, pluginID, UninstallOptions{PurgeStorageData: true})
	if !bizerr.Is(err, CodePluginLifecyclePreconditionVetoed) {
		t.Fatalf("expected active-release BeforeUninstall precondition bizerr, got %v", err)
	}
	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after vetoed uninstall to succeed, got error: %v", err)
	}
	if registry == nil || registry.Installed != catalog.InstalledYes {
		t.Fatalf("expected vetoed dynamic uninstall to keep plugin installed, got %#v", registry)
	}
}

// TestDynamicLifecycleUninstallRunsOnlyWhenPurgeRequested verifies dynamic
// Uninstall cleanup callbacks are skipped for keep-data uninstall and fail
// closed before state changes when data purge is requested.
func TestDynamicLifecycleUninstallRunsOnlyWhenPurgeRequested(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		blockingID = "plugin-dynamic-uninstall-cleanup-fail-closed"
		skipID     = "plugin-dynamic-uninstall-cleanup-skipped"
		version    = "v0.4.8"
	)

	for _, pluginID := range []string{blockingID, skipID} {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	}
	t.Cleanup(func() {
		for _, pluginID := range []string{blockingID, skipID} {
			testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		}
	})

	blockingArtifactPath := createDynamicLifecyclePreconditionArtifact(
		t,
		blockingID,
		"Dynamic Uninstall Cleanup Fail Closed Plugin",
		version,
		pluginbridge.LifecycleOperationUninstall,
	)
	if _, err := service.Install(ctx, blockingID, InstallOptions{}); err != nil {
		t.Fatalf("expected blocking dynamic plugin install to succeed, got error: %v", err)
	}
	err := service.Uninstall(ctx, blockingID, UninstallOptions{PurgeStorageData: true})
	if !bizerr.Is(err, CodePluginUninstallExecutionFailed) {
		t.Fatalf("expected purge uninstall lifecycle callback to return uninstall execution bizerr, got %v", err)
	}
	blockingRegistry, err := service.getPluginRegistry(ctx, blockingID)
	if err != nil {
		t.Fatalf("expected blocking plugin registry lookup to succeed, got error: %v", err)
	}
	if blockingRegistry == nil || blockingRegistry.Installed != catalog.InstalledYes {
		t.Fatalf("expected failed cleanup lifecycle to keep plugin installed, got %#v", blockingRegistry)
	}
	if cleanupErr := os.Remove(blockingArtifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
		t.Fatalf("failed to remove blocking artifact %s: %v", blockingArtifactPath, cleanupErr)
	}

	skipArtifactPath := createDynamicLifecyclePreconditionArtifact(
		t,
		skipID,
		"Dynamic Uninstall Cleanup Skipped Plugin",
		version,
		pluginbridge.LifecycleOperationUninstall,
	)
	if _, err = service.Install(ctx, skipID, InstallOptions{}); err != nil {
		t.Fatalf("expected skip dynamic plugin install to succeed, got error: %v", err)
	}
	if err = service.Uninstall(ctx, skipID, UninstallOptions{PurgeStorageData: false}); err != nil {
		t.Fatalf("expected keep-data uninstall to skip dynamic Uninstall lifecycle, got error: %v", err)
	}
	skipRegistry, err := service.getPluginRegistry(ctx, skipID)
	if err != nil {
		t.Fatalf("expected skip plugin registry lookup to succeed, got error: %v", err)
	}
	if skipRegistry == nil || skipRegistry.Installed != catalog.InstalledNo {
		t.Fatalf("expected keep-data uninstall to complete, got %#v", skipRegistry)
	}
	if cleanupErr := os.Remove(skipArtifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
		t.Fatalf("failed to remove skip artifact %s: %v", skipArtifactPath, cleanupErr)
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

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
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

	err = service.Uninstall(ctx, pluginID, UninstallOptions{PurgeStorageData: true})
	if !bizerr.Is(err, CodePluginDynamicArtifactMissingForUninstall) {
		t.Fatalf("expected missing-artifact uninstall bizerr, got %v", err)
	}

	enabled := true
	configsvc.SetPluginAllowForceUninstallOverride(&enabled)
	if err = service.Uninstall(ctx, pluginID, UninstallOptions{
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

// TestTenantDeleteRunsLifecyclePreconditions verifies tenant deletion fails
// closed when any plugin-owned lifecycle precondition vetoes tenant deletion.
func TestTenantDeleteRunsLifecyclePreconditions(t *testing.T) {
	var (
		service = newTestService()
		ctx     = context.Background()
	)
	plugin := pluginhost.NewSourcePlugin("plugin-tenant-delete-precondition")
	if err := plugin.Lifecycle().RegisterBeforeTenantDeleteHandler(func(
		ctx context.Context,
		input pluginhost.SourcePluginTenantLifecycleInput,
	) (bool, string, error) {
		if input.TenantID() != 8001 {
			t.Fatalf("expected tenant id to be published, got %d", input.TenantID())
		}
		return false, "plugin.test.tenant_delete_blocked", nil
	}); err != nil {
		t.Fatalf("register tenant delete lifecycle handler failed: %v", err)
	}
	cleanup, err := pluginhost.RegisterSourcePluginForTest(plugin)
	if err != nil {
		t.Fatalf("register tenant delete lifecycle callback failed: %v", err)
	}
	t.Cleanup(cleanup)

	err = service.EnsureTenantDeleteAllowed(ctx, 8001)
	if !bizerr.Is(err, CodePluginLifecyclePreconditionVetoed) {
		t.Fatalf("expected tenant delete lifecycle precondition bizerr, got %v", err)
	}
}

// TestTenantDeleteLifecyclePreconditionAllowsWhenNoParticipant verifies tenant
// deletion precondition checks are a no-op when no plugin participates.
func TestTenantDeleteLifecyclePreconditionAllowsWhenNoParticipant(t *testing.T) {
	service := newTestService()

	if err := service.EnsureTenantDeleteAllowed(context.Background(), 8002); err != nil {
		t.Fatalf("expected tenant delete precondition to allow without participants, got %v", err)
	}
}

// TestTenantPluginDisableRunsSourceLifecyclePrecondition verifies tenant-scoped
// plugin disable is routed to the target source plugin before state mutation.
func TestTenantPluginDisableRunsSourceLifecyclePrecondition(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-source-tenant-disable-precondition"
	)
	plugin := pluginhost.NewSourcePlugin(pluginID)
	if err := plugin.Lifecycle().RegisterBeforeTenantDisableHandler(func(
		ctx context.Context,
		input pluginhost.SourcePluginTenantLifecycleInput,
	) (bool, string, error) {
		if input.TenantID() != 8101 {
			t.Fatalf("expected tenant id to be published, got %d", input.TenantID())
		}
		return false, "plugin.test.tenant_disable_blocked", nil
	}); err != nil {
		t.Fatalf("register tenant disable lifecycle handler failed: %v", err)
	}
	cleanup, err := pluginhost.RegisterSourcePluginForTest(plugin)
	if err != nil {
		t.Fatalf("register tenant disable lifecycle callback failed: %v", err)
	}
	t.Cleanup(cleanup)

	err = service.EnsureTenantPluginDisableAllowed(ctx, pluginID, 8101)
	if !bizerr.Is(err, CodePluginLifecyclePreconditionVetoed) {
		t.Fatalf("expected tenant plugin disable lifecycle precondition bizerr, got %v", err)
	}
}

// TestDynamicLifecycleBeforeTenantDisableFailsClosed verifies tenant-scoped
// plugin disable also runs dynamic plugin lifecycle preconditions.
func TestDynamicLifecycleBeforeTenantDisableFailsClosed(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-before-tenant-disable"
	)

	artifactPath := createDynamicLifecyclePreconditionArtifact(
		t,
		pluginID,
		"Dynamic Before Tenant Disable Plugin",
		"v0.4.9",
		pluginbridge.LifecycleOperationBeforeTenantDisable,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected dynamic plugin enable to succeed, got error: %v", err)
	}
	err := service.EnsureTenantPluginDisableAllowed(ctx, pluginID, 8102)
	if !bizerr.Is(err, CodePluginLifecyclePreconditionVetoed) {
		t.Fatalf("expected dynamic tenant disable lifecycle precondition bizerr, got %v", err)
	}
}

// TestDynamicLifecycleBeforeTenantDeleteFailsClosed verifies tenant deletion
// runs dynamic plugin lifecycle preconditions for installed enabled plugins.
func TestDynamicLifecycleBeforeTenantDeleteFailsClosed(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-before-tenant-delete"
	)

	artifactPath := createDynamicLifecyclePreconditionArtifact(
		t,
		pluginID,
		"Dynamic Before Tenant Delete Plugin",
		"v0.4.10",
		pluginbridge.LifecycleOperationBeforeTenantDelete,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected dynamic plugin enable to succeed, got error: %v", err)
	}
	err := service.EnsureTenantDeleteAllowed(ctx, 8103)
	if !bizerr.Is(err, CodePluginLifecyclePreconditionVetoed) {
		t.Fatalf("expected dynamic tenant delete lifecycle precondition bizerr, got %v", err)
	}
}

// TestUninstallForceRequiresConfig verifies force uninstall is gated by
// plugin.allowForceUninstall before bypassing precondition vetoes.
func TestUninstallForceRequiresConfig(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-force-precondition"
	)
	registerUninstallLifecycleVetoForTest(t, pluginID)
	t.Cleanup(func() { configsvc.SetPluginAllowForceUninstallOverride(nil) })

	disabled := false
	configsvc.SetPluginAllowForceUninstallOverride(&disabled)
	err := service.ensureLifecyclePreconditionAllowed(ctx, pluginID, pluginhost.LifecycleHookBeforeUninstall, true)
	if !bizerr.Is(err, CodePluginForceUninstallDisabled) {
		t.Fatalf("expected force-disabled bizerr, got %v", err)
	}
}

// TestUninstallForceBypassesLifecyclePreconditionWhenConfigured verifies force
// uninstall can bypass precondition vetoes when the host config explicitly allows it.
func TestUninstallForceBypassesLifecyclePreconditionWhenConfigured(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-force-missing-after-precondition"
	)
	registerUninstallLifecycleVetoForTest(t, pluginID)
	t.Cleanup(func() { configsvc.SetPluginAllowForceUninstallOverride(nil) })

	enabled := true
	configsvc.SetPluginAllowForceUninstallOverride(&enabled)
	err := service.ensureLifecyclePreconditionAllowed(ctx, pluginID, pluginhost.LifecycleHookBeforeUninstall, true)
	if bizerr.Is(err, CodePluginLifecyclePreconditionVetoed) || bizerr.Is(err, CodePluginForceUninstallDisabled) {
		t.Fatalf("expected force to bypass lifecycle errors, got %v", err)
	}
	if err != nil {
		t.Fatalf("expected force bypass to continue, got %v", err)
	}
}

// TestSourceLifecycleBeforeInstallBlocksInstall verifies source-plugin facade
// BeforeInstall callbacks run before install SQL or registry state changes.
func TestSourceLifecycleBeforeInstallBlocksInstall(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-source-before-install-veto"
		operations []string
	)

	testutil.CreateTestPluginDir(t, pluginID)
	registerSourceLifecycleCallbacksForTest(
		t,
		pluginID,
		&operations,
		sourceLifecycleCallbackOptions{vetoOperation: pluginhost.LifecycleHookBeforeInstall.String()},
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	_, err := service.Install(ctx, pluginID, InstallOptions{})
	if !bizerr.Is(err, CodePluginLifecyclePreconditionVetoed) {
		t.Fatalf("expected BeforeInstall veto bizerr, got %v", err)
	}
	if len(operations) != 1 || operations[0] != pluginhost.LifecycleHookBeforeInstall.String()+":"+pluginID {
		t.Fatalf("expected BeforeInstall operation to be published once, got %#v", operations)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected source plugin registry row after discovery")
	}
	if registry.Installed != catalog.InstalledNo {
		t.Fatalf("expected vetoed install to keep plugin uninstalled, got installed=%d", registry.Installed)
	}
}

// TestSourceLifecycleBeforeDisableBlocksDisable verifies source-plugin facade
// BeforeDisable callbacks run before the source registry status changes.
func TestSourceLifecycleBeforeDisableBlocksDisable(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-source-before-disable-veto"
		operations []string
	)

	testutil.CreateTestPluginDir(t, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected source plugin enable to succeed, got error: %v", err)
	}
	registerSourceLifecycleCallbacksForTest(
		t,
		pluginID,
		&operations,
		sourceLifecycleCallbackOptions{vetoOperation: pluginhost.LifecycleHookBeforeDisable.String()},
	)

	err := service.Disable(ctx, pluginID)
	if !bizerr.Is(err, CodePluginLifecyclePreconditionVetoed) {
		t.Fatalf("expected BeforeDisable veto bizerr, got %v", err)
	}
	if len(operations) != 1 || operations[0] != pluginhost.LifecycleHookBeforeDisable.String()+":"+pluginID {
		t.Fatalf("expected BeforeDisable operation to be published once, got %#v", operations)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup to succeed, got error: %v", err)
	}
	if registry == nil || registry.Status != catalog.StatusEnabled {
		t.Fatalf("expected vetoed disable to keep plugin enabled, got %#v", registry)
	}
}

// TestSourceLifecycleBeforeUninstallBlocksUninstall verifies source-plugin
// facade BeforeUninstall callbacks run before uninstall cleanup side effects.
func TestSourceLifecycleBeforeUninstallBlocksUninstall(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-source-before-uninstall-veto"
		operations []string
	)

	testutil.CreateTestPluginDir(t, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}
	registerSourceLifecycleCallbacksForTest(
		t,
		pluginID,
		&operations,
		sourceLifecycleCallbackOptions{vetoOperation: pluginhost.LifecycleHookBeforeUninstall.String()},
	)

	err := service.Uninstall(ctx, pluginID, UninstallOptions{PurgeStorageData: true})
	if !bizerr.Is(err, CodePluginLifecyclePreconditionVetoed) {
		t.Fatalf("expected BeforeUninstall veto bizerr, got %v", err)
	}
	if len(operations) != 1 || operations[0] != pluginhost.LifecycleHookBeforeUninstall.String()+":"+pluginID {
		t.Fatalf("expected BeforeUninstall operation to be published once, got %#v", operations)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup to succeed, got error: %v", err)
	}
	if registry == nil || registry.Installed != catalog.InstalledYes {
		t.Fatalf("expected vetoed uninstall to keep plugin installed, got %#v", registry)
	}
}

// TestSourceLifecycleBeforeUninstallForceBypassesWhenConfigured verifies the
// existing force-uninstall policy also applies to unified BeforeUninstall.
func TestSourceLifecycleBeforeUninstallForceBypassesWhenConfigured(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-source-before-uninstall-force"
		operations []string
	)

	testutil.CreateTestPluginDir(t, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		configsvc.SetPluginAllowForceUninstallOverride(nil)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}
	registerSourceLifecycleCallbacksForTest(
		t,
		pluginID,
		&operations,
		sourceLifecycleCallbackOptions{vetoOperation: pluginhost.LifecycleHookBeforeUninstall.String()},
	)
	enabled := true
	configsvc.SetPluginAllowForceUninstallOverride(&enabled)

	if err := service.Uninstall(ctx, pluginID, UninstallOptions{PurgeStorageData: true, Force: true}); err != nil {
		t.Fatalf("expected force uninstall to bypass BeforeUninstall veto, got error: %v", err)
	}
	if len(operations) != 1 || operations[0] != pluginhost.LifecycleHookBeforeUninstall.String()+":"+pluginID {
		t.Fatalf("expected BeforeUninstall operation to be published once, got %#v", operations)
	}
}

// sourceLifecycleCallbackOptions configures a source-plugin lifecycle test facade.
type sourceLifecycleCallbackOptions struct {
	vetoOperation string
}

// TestSourceLifecycleAfterInstallRunsAfterInstall verifies source-plugin
// AfterInstall callbacks run only after the install transition succeeds.
func TestSourceLifecycleAfterInstallRunsAfterInstall(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-source-after-install"
		operations []string
	)

	testutil.CreateTestPluginDir(t, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	registerSourceLifecycleCallbacksForTest(
		t,
		pluginID,
		&operations,
		sourceLifecycleCallbackOptions{},
	)

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}
	expected := []string{
		pluginhost.LifecycleHookBeforeInstall.String() + ":" + pluginID,
		pluginhost.LifecycleHookAfterInstall.String() + ":" + pluginID,
	}
	if !sourceLifecycleTestStringSlicesEqual(operations, expected) {
		t.Fatalf("expected lifecycle operations %#v, got %#v", expected, operations)
	}
}

// registerUninstallLifecycleVetoForTest registers a source-plugin before
// uninstall callback that always vetoes.
func registerUninstallLifecycleVetoForTest(t *testing.T, pluginID string) {
	t.Helper()

	plugin := pluginhost.NewSourcePlugin(pluginID)
	if err := plugin.Lifecycle().RegisterBeforeUninstallHandler(func(
		ctx context.Context,
		input pluginhost.SourcePluginLifecycleInput,
	) (bool, string, error) {
		return false, "plugin.test.uninstall_blocked", nil
	}); err != nil {
		t.Fatalf("register uninstall lifecycle handler failed: %v", err)
	}
	cleanup, err := pluginhost.RegisterSourcePluginForTest(plugin)
	if err != nil {
		t.Fatalf("register lifecycle callback failed: %v", err)
	}
	t.Cleanup(cleanup)
}

// registerSourceLifecycleCallbacksForTest replaces the source-plugin fixture
// lifecycle callbacks while preserving its embedded filesystem declaration.
func registerSourceLifecycleCallbacksForTest(
	t *testing.T,
	pluginID string,
	operations *[]string,
	options sourceLifecycleCallbackOptions,
) {
	t.Helper()

	previous, ok := pluginhost.GetSourcePlugin(pluginID)
	if !ok || previous == nil {
		t.Fatalf("expected source plugin fixture %s to be registered", pluginID)
	}
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(previous.GetEmbeddedFiles())
	handler := func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
		if input == nil {
			t.Fatalf("expected lifecycle input to be published")
		}
		operation := strings.TrimSpace(input.Operation())
		*operations = append(*operations, operation+":"+input.PluginID())
		if input.PluginID() != pluginID {
			t.Fatalf("expected plugin id %s, got %s", pluginID, input.PluginID())
		}
		if options.vetoOperation == operation {
			return false, "plugin." + pluginID + "." + operation + ".veto", nil
		}
		return true, "", nil
	}
	if err := plugin.Lifecycle().RegisterBeforeInstallHandler(handler); err != nil {
		t.Fatalf("failed to register before-install lifecycle handler: %v", err)
	}
	if err := plugin.Lifecycle().RegisterAfterInstallHandler(func(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) error {
		if input == nil {
			t.Fatalf("expected lifecycle input to be published")
		}
		operation := strings.TrimSpace(input.Operation())
		*operations = append(*operations, operation+":"+input.PluginID())
		if input.PluginID() != pluginID {
			t.Fatalf("expected plugin id %s, got %s", pluginID, input.PluginID())
		}
		return nil
	}); err != nil {
		t.Fatalf("failed to register after-install lifecycle handler: %v", err)
	}
	if err := plugin.Lifecycle().RegisterBeforeDisableHandler(handler); err != nil {
		t.Fatalf("failed to register before-disable lifecycle handler: %v", err)
	}
	if err := plugin.Lifecycle().RegisterBeforeUninstallHandler(handler); err != nil {
		t.Fatalf("failed to register before-uninstall lifecycle handler: %v", err)
	}
	cleanup, err := pluginhost.RegisterSourcePluginForTest(plugin)
	if err != nil {
		t.Fatalf("failed to replace source plugin fixture %s: %v", pluginID, err)
	}
	t.Cleanup(cleanup)
}

// sourceLifecycleTestStringSlicesEqual reports whether two ordered string slices are equal.
func sourceLifecycleTestStringSlicesEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

// createDynamicLifecyclePreconditionArtifact writes a dynamic artifact with one
// lifecycle contract and intentionally missing guest bridge exports so host
// execution fails closed before lifecycle side effects.
func createDynamicLifecyclePreconditionArtifact(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	operation pluginbridge.LifecycleOperation,
) string {
	t.Helper()

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), pluginID+".wasm")
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    pluginName,
			Version: version,
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: pluginbridge.RuntimeKindWasm,
			ABIVersion:  pluginbridge.SupportedABIVersion,
			LifecycleContracts: []*pluginbridge.LifecycleContract{
				{
					Operation:    operation,
					RequestType:  "DynamicLifecycleReq",
					InternalPath: "/__lifecycle/" + strings.ToLower(strings.TrimPrefix(operation.String(), "Before")),
					TimeoutMs:    1000,
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		&pluginbridge.BridgeSpec{
			ABIVersion:     pluginbridge.ABIVersionV1,
			RuntimeKind:    pluginbridge.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   pluginbridge.CodecProtobuf,
			ResponseCodec:  pluginbridge.CodecProtobuf,
		},
	)
	return artifactPath
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

	if _, err := service.Install(ctx, pluginID, InstallOptions{Authorization: authorization}); err != nil {
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
			"scope_nature: tenant_aware\n"+
			"supports_multi_tenant: false\n"+
			"default_install_mode: global\n"+
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

	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
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

	if err = service.Uninstall(ctx, pluginID, UninstallOptions{PurgeStorageData: true}); err != nil {
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
