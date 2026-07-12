// This file covers root-facade list methods defined in plugin_list.go.

package plugin

import (
	"context"
	"encoding/json"
	pluginv1 "lina-core/api/plugin/v1"
	"os"
	"path/filepath"
	"testing"

	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/internal/service/startupstats"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/statusflag"
)

// findPluginItem returns one plugin list item by plugin ID for list assertions.
func findPluginItem(out *ListOutput, pluginID string) *PluginItem {
	if out == nil {
		return nil
	}
	for _, current := range out.List {
		if current != nil && current.Id == pluginID {
			return current
		}
	}
	return nil
}

// TestManagementListCacheAvoidsRepeatedManifestScans verifies the management
// list read model is reused until an explicit plugin-runtime invalidation.
func TestManagementListCacheAvoidsRepeatedManifestScans(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = startupstats.WithCollector(context.Background(), startupstats.New())
		pluginID = "plugin-dev-source-management-list-cache"
	)

	createTestSourceDependencyPlugin(t, pluginID, "Source Management List Cache", "v0.1.0", "")
	cleanupTestPluginIDs(t, context.Background(), pluginID)

	// Filter by plugin ID so the assertion is independent of default list page size.
	// Official plugin workspaces with multi-cloud storage plugins exceed page size 20.
	first, err := service.List(ctx, ListInput{ID: pluginID})
	if err != nil {
		t.Fatalf("build first management list: %v", err)
	}
	if findPluginItem(first, pluginID) == nil {
		t.Fatalf("expected first management list to include %s", pluginID)
	}

	second, err := service.List(ctx, ListInput{ID: pluginID})
	if err != nil {
		t.Fatalf("read cached management list: %v", err)
	}
	if findPluginItem(second, pluginID) == nil {
		t.Fatalf("expected cached management list to include %s", pluginID)
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginScans); got != 1 {
		t.Fatalf("expected cached list to avoid repeated scans, got %d", got)
	}

	service.InvalidateManagementListCache(ctx, "test")
	third, err := service.List(ctx, ListInput{ID: pluginID})
	if err != nil {
		t.Fatalf("rebuild invalidated management list: %v", err)
	}
	if findPluginItem(third, pluginID) == nil {
		t.Fatalf("expected rebuilt management list to include %s", pluginID)
	}

	snapshot = startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginScans); got != 2 {
		t.Fatalf("expected invalidated list to rescan once, got %d", got)
	}
}

// TestRuntimeCacheChangeInvalidatesManagementList verifies lifecycle cache
// publications clear the plugin-management list read model.
func TestRuntimeCacheChangeInvalidatesManagementList(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = startupstats.WithCollector(context.Background(), startupstats.New())
		pluginID = "plugin-dev-source-management-list-runtime-invalidate"
	)

	createTestSourceDependencyPlugin(t, pluginID, "Source Management List Runtime Invalidate", "v0.1.0", "")
	cleanupTestPluginIDs(t, context.Background(), pluginID)

	if _, err := service.List(ctx, ListInput{}); err != nil {
		t.Fatalf("build management list: %v", err)
	}
	if _, err := service.List(ctx, ListInput{}); err != nil {
		t.Fatalf("read cached management list: %v", err)
	}
	if err := service.MarkRuntimeCacheChanged(ctx, "test_runtime_cache_changed"); err != nil {
		t.Fatalf("mark runtime cache changed: %v", err)
	}
	if _, err := service.List(ctx, ListInput{}); err != nil {
		t.Fatalf("rebuild after runtime cache change: %v", err)
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginScans); got != 2 {
		t.Fatalf("expected runtime cache change to invalidate list, got %d scans", got)
	}
}

// TestPrewarmManagementListPopulatesCache verifies startup prewarm fills the
// same complete read model later consumed by management list requests.
func TestPrewarmManagementListPopulatesCache(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = startupstats.WithCollector(context.Background(), startupstats.New())
		pluginID = "plugin-dev-source-management-list-prewarm"
	)

	createTestSourceDependencyPlugin(t, pluginID, "Source Management List Prewarm", "v0.1.0", "")
	cleanupTestPluginIDs(t, context.Background(), pluginID)

	if err := service.PrewarmManagementList(ctx); err != nil {
		t.Fatalf("prewarm management list: %v", err)
	}
	out, err := service.List(ctx, ListInput{ID: pluginID})
	if err != nil {
		t.Fatalf("read prewarmed management list: %v", err)
	}
	if len(out.List) != 1 || out.List[0] == nil || out.List[0].Id != pluginID {
		t.Fatalf("expected prewarmed filtered list for %s, got %#v", pluginID, out)
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginScans); got != 1 {
		t.Fatalf("expected prewarm plus list to scan once, got %d", got)
	}
}

// TestManagementListCacheIsLocaleScoped verifies localized plugin metadata
// cannot leak from startup prewarm or another request locale.
func TestManagementListCacheIsLocaleScoped(t *testing.T) {
	var (
		service   = newTestService()
		baseCtx   = context.Background()
		defaultID = "plugin-dev-source-management-list-default-locale"
		englishID = "plugin-dev-source-management-list-english-locale"
	)

	createTestSourceDependencyPlugin(t, defaultID, "Source Management List Default Locale", "v0.1.0", "")
	createTestSourceDependencyPlugin(t, englishID, "Source Management List English Locale", "v0.1.0", "")
	cleanupTestPluginIDs(t, context.Background(), defaultID, englishID)

	if _, err := service.List(baseCtx, ListInput{ID: defaultID}); err != nil {
		t.Fatalf("build default-locale management list: %v", err)
	}

	englishCtx := context.WithValue(
		context.Background(),
		gctx.StrKey("BizCtx"),
		&model.Context{Locale: i18nsvc.EnglishLocale},
	)
	if _, err := service.List(englishCtx, ListInput{ID: englishID}); err != nil {
		t.Fatalf("build english-locale management list: %v", err)
	}
	baseKey, err := service.managementListCacheKey(baseCtx)
	if err != nil {
		t.Fatalf("build default-locale cache key: %v", err)
	}
	baseRevision, err := service.runtimeCacheRevisionCtrl.CurrentRevision(baseCtx)
	if err != nil {
		t.Fatalf("read default-locale runtime revision: %v", err)
	}
	if baseKey.Locale != i18nsvc.DefaultLocale {
		t.Fatalf("expected default-locale cache key locale %s, got %s", i18nsvc.DefaultLocale, baseKey.Locale)
	}
	baseBundleRevision, err := service.i18nSvc.BundleRevision(baseCtx, i18nsvc.DefaultLocale)
	if err != nil {
		t.Fatalf("read default-locale runtime bundle revision: %v", err)
	}
	if baseKey.RuntimeBundleVersion != baseBundleRevision.Version {
		t.Fatalf("expected default-locale cache key bundle version to match runtime bundle, got %d", baseKey.RuntimeBundleVersion)
	}
	if baseKey.RuntimeRevision != baseRevision {
		t.Fatalf("expected default-locale cache key runtime revision %d, got %d", baseRevision, baseKey.RuntimeRevision)
	}
	if _, ok := service.managementListCache.Get(baseKey); !ok {
		t.Fatalf("expected default-locale management list cache")
	}
	englishKey, err := service.managementListCacheKey(englishCtx)
	if err != nil {
		t.Fatalf("build english-locale cache key: %v", err)
	}
	englishRevision, err := service.runtimeCacheRevisionCtrl.CurrentRevision(englishCtx)
	if err != nil {
		t.Fatalf("read english-locale runtime revision: %v", err)
	}
	if englishKey.Locale != i18nsvc.EnglishLocale {
		t.Fatalf("expected english-locale cache key locale %s, got %s", i18nsvc.EnglishLocale, englishKey.Locale)
	}
	englishBundleRevision, err := service.i18nSvc.BundleRevision(englishCtx, i18nsvc.EnglishLocale)
	if err != nil {
		t.Fatalf("read english-locale runtime bundle revision: %v", err)
	}
	if englishKey.RuntimeBundleVersion != englishBundleRevision.Version {
		t.Fatalf("expected english-locale cache key bundle version to match runtime bundle, got %d", englishKey.RuntimeBundleVersion)
	}
	if englishKey.RuntimeRevision != englishRevision {
		t.Fatalf("expected english-locale cache key runtime revision %d, got %d", englishRevision, englishKey.RuntimeRevision)
	}
	if englishKey.String() == baseKey.String() {
		t.Fatalf("expected locale-scoped cache keys to differ, got %s", englishKey.String())
	}
	if _, ok := service.managementListCache.Get(englishKey); !ok {
		t.Fatalf("expected english-locale management list cache")
	}
}

// TestListIncludesBuiltinByDefault verifies ordinary management list reads
// return builtin plugins with distribution projection and reuse the cached
// read model across repeated queries.
func TestListIncludesBuiltinByDefault(t *testing.T) {
	var (
		service   = newTestService()
		ctx       = startupstats.WithCollector(context.Background(), startupstats.New())
		managedID = "plugin-dev-list-managed-distribution"
		builtinID = "plugin-dev-list-builtin-distribution"
	)

	createTestSourceDependencyPlugin(t, managedID, "Managed Distribution", "v0.1.0", "")
	createTestSourceDependencyPlugin(
		t,
		builtinID,
		"Builtin Distribution",
		"v0.1.0",
		"distribution: builtin\n",
	)
	cleanupTestPluginIDs(t, context.Background(), managedID, builtinID)

	defaultOut, err := service.List(ctx, ListInput{ID: "plugin-dev-list-"})
	if err != nil {
		t.Fatalf("expected default list to succeed, got error: %v", err)
	}
	if findPluginItem(defaultOut, managedID) == nil {
		t.Fatalf("expected managed plugin in default list")
	}
	builtin := findPluginItem(defaultOut, builtinID)
	if builtin == nil {
		t.Fatalf("expected builtin plugin in default list")
	}
	if builtin.Distribution != pluginv1.PluginDistributionBuiltin.String() {
		t.Fatalf("expected builtin distribution, got %#v", builtin)
	}

	// Compatibility flag must not hide builtin plugins or force a rebuild.
	compatOut, err := service.List(ctx, ListInput{
		ID:             "plugin-dev-list-",
		IncludeBuiltin: false,
	})
	if err != nil {
		t.Fatalf("expected compatibility list to succeed, got error: %v", err)
	}
	if findPluginItem(compatOut, builtinID) == nil {
		t.Fatalf("expected builtin plugin even when includeBuiltin=false")
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginScans); got != 1 {
		t.Fatalf("expected repeated list reads to reuse cached read model, got %d scans", got)
	}
}

// TestSyncAndListRetainsMissingRuntimeRegistryAndReconcilesState verifies that
// missing runtime artifacts reconcile registry state without hiding the plugin.
func TestSyncAndListRetainsMissingRuntimeRegistryAndReconcilesState(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dev-dynamic-registry-missing"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		"Runtime Registry Missing Plugin",
		"v0.9.4",
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = service.setPluginInstalled(ctx, pluginID, statusflag.Installed.Int()); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = service.setPluginStatus(ctx, pluginID, statusflag.EnabledValue.Int()); err != nil {
		t.Fatalf("expected dynamic plugin enable state to be set, got error: %v", err)
	}
	if err = os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove dynamic artifact: %v", err)
	}

	out, err := service.SyncAndList(ctx)
	if err != nil {
		t.Fatalf("expected sync-and-list to tolerate missing dynamic artifact, got error: %v", err)
	}

	var item *PluginItem
	for _, current := range out.List {
		if current != nil && current.Id == pluginID {
			item = current
			break
		}
	}
	if item == nil {
		t.Fatalf("expected missing dynamic plugin to remain visible in plugin list")
	}
	if item.Installed != statusflag.Uninstalled.Int() {
		t.Fatalf("expected missing dynamic plugin installed state to reconcile to %d, got %d", statusflag.Uninstalled.Int(), item.Installed)
	}
	if item.Enabled != statusflag.Disabled.Int() {
		t.Fatalf("expected missing dynamic plugin enabled state to reconcile to %d, got %d", statusflag.Disabled.Int(), item.Enabled)
	}

	runtimeStates, err := service.ListRuntimeStates(ctx)
	if err != nil {
		t.Fatalf("expected runtime state list to succeed, got error: %v", err)
	}
	var runtimeState *PluginDynamicStateItem
	for _, current := range runtimeStates.List {
		if current != nil && current.Id == pluginID {
			runtimeState = current
			break
		}
	}
	if runtimeState == nil {
		t.Fatalf("expected missing dynamic plugin to remain visible in public runtime states")
	}
	if runtimeState.Installed != statusflag.Uninstalled.Int() || runtimeState.Enabled != statusflag.Disabled.Int() {
		t.Fatalf("expected public runtime state to reconcile to uninstalled+disabled, got installed=%d enabled=%d", runtimeState.Installed, runtimeState.Enabled)
	}
	if runtimeState.RuntimeState != RuntimeUpgradeStateNormal {
		t.Fatalf("expected missing dynamic plugin public runtime state to stay normal, got %s", runtimeState.RuntimeState)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected runtime registry row to remain after reconciliation")
	}
	if registry.Installed != statusflag.Uninstalled.Int() || registry.Status != statusflag.Disabled.Int() {
		t.Fatalf("expected runtime registry row to reconcile to uninstalled+disabled, got installed=%d enabled=%d", registry.Installed, registry.Status)
	}
}

// TestListProjectsMissingRuntimeRegistryWithoutWriting verifies the GET-list
// path can show a safe missing-artifact state without mutating governance rows.
func TestListProjectsMissingRuntimeRegistryWithoutWriting(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dev-dynamic-registry-readonly"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		"Runtime Registry Readonly Plugin",
		"v0.9.5",
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = service.setPluginInstalled(ctx, pluginID, statusflag.Installed.Int()); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = service.setPluginStatus(ctx, pluginID, statusflag.EnabledValue.Int()); err != nil {
		t.Fatalf("expected dynamic plugin enable state to be set, got error: %v", err)
	}

	registryBefore, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup before list to succeed, got error: %v", err)
	}
	if registryBefore == nil {
		t.Fatalf("expected runtime registry row before list")
	}
	if err = os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove dynamic artifact: %v", err)
	}

	out, err := service.List(ctx, ListInput{ID: pluginID})
	if err != nil {
		t.Fatalf("expected read-only list to tolerate missing dynamic artifact, got error: %v", err)
	}

	item := findPluginItem(out, pluginID)
	if item == nil {
		t.Fatalf("expected missing dynamic plugin to remain visible in read-only plugin list")
	}
	if item.Installed != statusflag.Uninstalled.Int() {
		t.Fatalf("expected read-only projection installed state to be %d, got %d", statusflag.Uninstalled.Int(), item.Installed)
	}
	if item.Enabled != statusflag.Disabled.Int() {
		t.Fatalf("expected read-only projection enabled state to be %d, got %d", statusflag.Disabled.Int(), item.Enabled)
	}

	registryAfter, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup after list to succeed, got error: %v", err)
	}
	if registryAfter == nil {
		t.Fatalf("expected runtime registry row to remain after read-only list")
	}
	if registryAfter.Installed != registryBefore.Installed ||
		registryAfter.Status != registryBefore.Status ||
		registryAfter.DesiredState != registryBefore.DesiredState ||
		registryAfter.CurrentState != registryBefore.CurrentState ||
		registryAfter.Generation != registryBefore.Generation ||
		registryAfter.ReleaseId != registryBefore.ReleaseId {
		t.Fatalf(
			"expected read-only list not to mutate registry, before installed=%d status=%d desired=%s current=%s generation=%d release=%d after installed=%d status=%d desired=%s current=%s generation=%d release=%d",
			registryBefore.Installed,
			registryBefore.Status,
			registryBefore.DesiredState,
			registryBefore.CurrentState,
			registryBefore.Generation,
			registryBefore.ReleaseId,
			registryAfter.Installed,
			registryAfter.Status,
			registryAfter.DesiredState,
			registryAfter.CurrentState,
			registryAfter.Generation,
			registryAfter.ReleaseId,
		)
	}
}

// TestNormalizePluginListPageBounds verifies plugin management list pagination
// applies stable defaults and the service-side maximum page size.
func TestNormalizePluginListPageBounds(t *testing.T) {
	pageNum, pageSize := normalizeListPage(0, 0)
	if pageNum != defaultListPageNum || pageSize != defaultListPageSize {
		t.Fatalf("expected default page %d/%d, got %d/%d", defaultListPageNum, defaultListPageSize, pageNum, pageSize)
	}

	pageNum, pageSize = normalizeListPage(2, maxListPageSize+1)
	if pageNum != 2 || pageSize != maxListPageSize {
		t.Fatalf("expected bounded page 2/%d, got %d/%d", maxListPageSize, pageNum, pageSize)
	}
}

// TestListPaginatesAndKeepsGovernanceDetailsOutOfSummary verifies the GET-list
// path is a paginated summary while exact detail lookup still exposes governance
// review payloads for the selected plugin only.
func TestListPaginatesAndKeepsGovernanceDetailsOutOfSummary(t *testing.T) {
	var (
		service       = newTestService()
		ctx           = context.Background()
		filterPrefix  = "plugin-dev-summary-page-"
		firstPluginID = filterPrefix + "a"
		secondID      = filterPrefix + "b"
		detailID      = "plugin-dev-dynamic-summary-detail"
		version       = "v0.1.0"
	)

	createTestSourceDependencyPlugin(t, firstPluginID, "Summary Page A", version, "")
	createTestSourceDependencyPlugin(t, secondID, "Summary Page B", version, "")
	cleanupTestPluginIDs(t, context.Background(), firstPluginID, secondID, detailID)

	out, err := service.List(ctx, ListInput{
		PageNum:  2,
		PageSize: 1,
		ID:       filterPrefix,
	})
	if err != nil {
		t.Fatalf("expected paginated summary list to succeed, got error: %v", err)
	}
	if out.Total != 2 || len(out.List) != 1 {
		t.Fatalf("expected second page of two filtered plugins, got total=%d len=%d", out.Total, len(out.List))
	}
	if out.List[0] == nil || out.List[0].Id != secondID {
		t.Fatalf("expected second plugin on page 2, got %#v", out.List)
	}
	emptyPage, err := service.List(ctx, ListInput{
		PageNum:  3,
		PageSize: 1,
		ID:       filterPrefix,
	})
	if err != nil {
		t.Fatalf("expected empty page summary list to succeed, got error: %v", err)
	}
	if emptyPage.Total != 2 || len(emptyPage.List) != 0 {
		t.Fatalf("expected empty page with retained total=2, got total=%d len=%d", emptyPage.Total, len(emptyPage.List))
	}

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), detailID+".wasm")
	t.Cleanup(func() {
		if err := os.Remove(artifactPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("failed to remove dynamic summary artifact %s: %v", artifactPath, err)
		}
	})
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:                  detailID,
			Name:                "Dynamic Summary Detail Plugin",
			Version:             version,
			Type:                pluginv1.PluginTypeDynamic.String(),
			ScopeNature:         pluginv1.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &testutil.DefaultTestSupportsMultiTenant,
			DefaultInstallMode:  pluginv1.InstallModeTenantScoped.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
			HostServices: []*protocol.HostServiceSpec{
				{
					Service: protocol.HostServiceStorage,
					Methods: []string{
						protocol.HostServiceMethodStorageGet,
					},
					Paths: []string{"reports/"},
				},
			},
			RouteCount: 1,
		},
		nil,
		nil,
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/governed-report",
				Method:      "GET",
				Access:      protocol.AccessPublic,
				RequestType: "DynamicSummaryDetailReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.ABIVersionV1,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
			AllocExport:    protocol.DefaultGuestAllocExport,
			ExecuteExport:  protocol.DefaultGuestExecuteExport,
		},
	)
	service.InvalidateManagementListCache(ctx, "dynamic_summary_artifact_created")

	summary, err := service.List(ctx, ListInput{ID: detailID})
	if err != nil {
		t.Fatalf("expected summary list for dynamic plugin to succeed, got error: %v", err)
	}
	summaryItem := findPluginItem(summary, detailID)
	if summaryItem == nil {
		t.Fatalf("expected dynamic plugin summary item")
	}
	if !summaryItem.AuthorizationRequired {
		t.Fatalf("expected summary to retain authorization-required status, got %#v", summaryItem)
	}
	if summaryItem.DependencyCheck != nil {
		t.Fatalf("expected summary list not to attach dependency check, got %#v", summaryItem.DependencyCheck)
	}
	if len(summaryItem.RequestedHostServices) != 0 ||
		len(summaryItem.AuthorizedHostServices) != 0 ||
		len(summaryItem.DeclaredRoutes) != 0 {
		t.Fatalf("expected summary list to omit detail governance payloads, got requested=%#v authorized=%#v routes=%#v", summaryItem.RequestedHostServices, summaryItem.AuthorizedHostServices, summaryItem.DeclaredRoutes)
	}

	full, err := service.ReadOnlyList(ctx)
	if err != nil {
		t.Fatalf("expected full read-only plugin list to succeed, got error: %v", err)
	}
	fullItem := findPluginItem(full, detailID)
	if fullItem == nil {
		t.Fatalf("expected dynamic plugin full read-only item")
	}
	if fullItem.DependencyCheck == nil {
		t.Fatalf("expected full read-only item to attach dependency check")
	}
	if len(fullItem.RequestedHostServices) != 1 || fullItem.RequestedHostServices[0].Service != protocol.HostServiceStorage {
		t.Fatalf("expected full read-only item requested host service, got %#v", fullItem.RequestedHostServices)
	}
	if len(fullItem.DeclaredRoutes) != 1 || fullItem.DeclaredRoutes[0].Path != "/governed-report" {
		t.Fatalf("expected full read-only item declared route, got %#v", fullItem.DeclaredRoutes)
	}

	detail, err := service.Get(ctx, detailID)
	if err != nil {
		t.Fatalf("expected plugin detail to succeed, got error: %v", err)
	}
	if detail.DependencyCheck == nil {
		t.Fatalf("expected detail to attach dependency check")
	}
	if len(detail.RequestedHostServices) != 1 || detail.RequestedHostServices[0].Service != protocol.HostServiceStorage {
		t.Fatalf("expected detail requested host service, got %#v", detail.RequestedHostServices)
	}
	if len(detail.DeclaredRoutes) != 1 || detail.DeclaredRoutes[0].Path != "/governed-report" {
		t.Fatalf("expected detail declared route, got %#v", detail.DeclaredRoutes)
	}

	_, err = service.Install(ctx, detailID, InstallOptions{
		Authorization: &HostServiceAuthorizationInput{
			Services: []*HostServiceAuthorizationDecision{
				{
					Service: protocol.HostServiceStorage,
					Methods: []string{
						protocol.HostServiceMethodStorageGet,
					},
					Paths: []string{"reports/"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected dynamic plugin install with authorization to succeed, got error: %v", err)
	}
	if err = os.Remove(artifactPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove staging artifact %s: %v", artifactPath, err)
	}

	installedDetail, err := service.Get(ctx, detailID)
	if err != nil {
		t.Fatalf("expected installed plugin detail to fall back to release snapshot, got error: %v", err)
	}
	if len(installedDetail.AuthorizedHostServices) != 1 ||
		installedDetail.AuthorizedHostServices[0].Service != protocol.HostServiceStorage {
		t.Fatalf("expected installed detail authorized snapshot, got %#v", installedDetail.AuthorizedHostServices)
	}
	if len(installedDetail.DeclaredRoutes) != 1 ||
		installedDetail.DeclaredRoutes[0].Path != "/governed-report" {
		t.Fatalf("expected installed detail declared route from release snapshot, got %#v", installedDetail.DeclaredRoutes)
	}
	installedFull, err := service.ReadOnlyList(ctx)
	if err != nil {
		t.Fatalf("expected installed plugin full read-only list to fall back to release snapshot, got error: %v", err)
	}
	installedFullItem := findPluginItem(installedFull, detailID)
	if installedFullItem == nil {
		t.Fatalf("expected installed plugin full read-only item")
	}
	if len(installedFullItem.AuthorizedHostServices) != 1 ||
		installedFullItem.AuthorizedHostServices[0].Service != protocol.HostServiceStorage {
		t.Fatalf("expected installed full read-only item authorized snapshot, got %#v", installedFullItem.AuthorizedHostServices)
	}
	if len(installedFullItem.DeclaredRoutes) != 1 ||
		installedFullItem.DeclaredRoutes[0].Path != "/governed-report" {
		t.Fatalf("expected installed full read-only item declared route from release snapshot, got %#v", installedFullItem.DeclaredRoutes)
	}
}

// TestListOwnerHostServiceSummarySkipsFullDependencyChecks verifies the first
// management list screen stays on the summary projection even when a dynamic
// plugin declares owner-aware host service dependencies.
func TestListOwnerHostServiceSummarySkipsFullDependencyChecks(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = startupstats.WithCollector(context.Background(), startupstats.New())
		ownerID    = "plugin-dev-source-owner-summary-base"
		consumerID = "plugin-dev-dynamic-owner-summary-consumer"
	)

	createTestSourceDependencyPlugin(t, ownerID, "Source Owner Summary Base", "v0.1.0", "")
	writeTestDynamicOwnerHostServiceArtifactWithDependencies(
		t,
		consumerID,
		"Dynamic Owner Summary Consumer",
		"v0.1.0",
		&plugintypes.DependencySpec{Plugins: []*plugintypes.PluginDependencySpec{
			testPluginDependencySpec(ownerID, ">=0.1.0"),
		}},
		[]*protocol.HostServiceSpec{{
			Owner:   ownerID,
			Service: "ai",
			Version: "v1",
			Methods: []string{
				"text.generate",
			},
		}},
		buildVersionedRuntimeFrontendAssets("owner-summary-consumer"),
	)
	cleanupTestPluginIDs(t, context.Background(), ownerID, consumerID)

	out, err := service.List(ctx, ListInput{ID: consumerID})
	if err != nil {
		t.Fatalf("expected owner summary list to succeed, got error: %v", err)
	}
	item := findPluginItem(out, consumerID)
	if item == nil {
		t.Fatalf("expected owner dynamic plugin summary item")
	}
	if item.DependencyCheck != nil {
		t.Fatalf("expected summary list not to attach dependency check, got %#v", item.DependencyCheck)
	}
	if len(item.RequestedHostServices) != 0 || len(item.AuthorizedHostServices) != 0 {
		t.Fatalf("expected summary list to omit owner host service detail, got requested=%#v authorized=%#v", item.RequestedHostServices, item.AuthorizedHostServices)
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginScans); got != 1 {
		t.Fatalf("expected summary list to scan manifests once, got %d", got)
	}
	if got := snapshot.CounterValue(startupstats.CounterCatalogSnapshotBuilds); got != 1 {
		t.Fatalf("expected summary list not to run full dependency snapshots, got %d catalog snapshots", got)
	}
	if got := snapshot.CounterValue(startupstats.CounterIntegrationSnapshotBuilds); got != 0 {
		t.Fatalf("expected summary list not to build integration snapshots, got %d", got)
	}
}

// TestReadOnlyListProjectionBatchesDependencyChecks verifies full management
// projections reuse manifest and store snapshots while attaching dependency
// status to every row.
func TestReadOnlyListProjectionBatchesDependencyChecks(t *testing.T) {
	var (
		service      = newTestService()
		ctx          = startupstats.WithCollector(context.Background(), startupstats.New())
		dependencyID = "plugin-dev-source-projection-batch-dependency"
		targetID     = "plugin-dev-source-projection-batch-target"
	)

	createTestSourceDependencyPlugin(t, dependencyID, "Projection Batch Dependency", "v0.1.0", "")
	createTestSourceDependencyPlugin(
		t,
		targetID,
		"Projection Batch Target",
		"v0.1.0",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: "+dependencyID+"\n"+
			"      version: \">=0.1.0\"\n",
	)
	cleanupTestPluginIDs(t, context.Background(), dependencyID, targetID)

	out, err := service.ReadOnlyList(ctx)
	if err != nil {
		t.Fatalf("expected full read-only plugin list to succeed, got error: %v", err)
	}
	dependencyItem := findPluginItem(out, dependencyID)
	if dependencyItem == nil || dependencyItem.DependencyCheck == nil {
		t.Fatalf("expected dependency plugin to include dependency check, got %#v", dependencyItem)
	}
	if dependencyItem.DependencyCheck.TargetID != dependencyID {
		t.Fatalf("expected dependency check target %s, got %#v", dependencyID, dependencyItem.DependencyCheck)
	}
	targetItem := findPluginItem(out, targetID)
	if targetItem == nil || targetItem.DependencyCheck == nil {
		t.Fatalf("expected target plugin to include dependency check, got %#v", targetItem)
	}
	if targetItem.DependencyCheck.TargetID != targetID {
		t.Fatalf("expected target dependency check target %s, got %#v", targetID, targetItem.DependencyCheck)
	}
	if len(targetItem.DependencyCheck.Blockers) != 1 ||
		targetItem.DependencyCheck.Blockers[0].DependencyID != dependencyID {
		t.Fatalf("expected target dependency check to expose missing dependency blocker, got %#v", targetItem.DependencyCheck)
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginScans); got != 1 {
		t.Fatalf("expected full projection to scan manifests once, got %d", got)
	}
	if got := snapshot.CounterValue(startupstats.CounterCatalogSnapshotBuilds); got != 1 {
		t.Fatalf("expected full projection to build one store snapshot, got %d", got)
	}
	if got := snapshot.CounterValue(startupstats.CounterIntegrationSnapshotBuilds); got != 0 {
		t.Fatalf("expected full read-only projection not to build integration snapshot, got %d", got)
	}
}

// TestGetReturnsStableNotFoundBizerr verifies exact detail lookup reports a
// stable business error when no discovered or registered plugin matches.
func TestGetReturnsStableNotFoundBizerr(t *testing.T) {
	service := newTestService()
	_, err := service.Get(context.Background(), "plugin-detail-missing")
	if !bizerr.Is(err, CodePluginNotFound) {
		t.Fatalf("expected plugin not-found bizerr, got %v", err)
	}
}

// TestListMarksInstalledDynamicPluginWithHigherArtifactPendingUpgrade verifies
// dynamic artifact replacement is exposed as a pending runtime upgrade without
// switching the effective registry version.
func TestListMarksInstalledDynamicPluginWithHigherArtifactPendingUpgrade(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-dynamic-runtime-upgrade-pending"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Pending Plugin",
		oldVersion,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}

	oldRelease, err := service.getPluginRelease(ctx, pluginID, oldVersion)
	if err != nil {
		t.Fatalf("expected old dynamic release lookup to succeed, got error: %v", err)
	}
	if oldRelease == nil {
		t.Fatal("expected old dynamic release row")
	}

	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Pending Plugin",
		newVersion,
		nil,
		nil,
	)
	newManifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected new dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected new dynamic manifest sync to succeed, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected dynamic registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatal("expected dynamic registry row")
	}
	if registry.Version != oldVersion {
		t.Fatalf("expected effective version %s to stay pinned, got %s", oldVersion, registry.Version)
	}
	if registry.ReleaseId != oldRelease.Id {
		t.Fatalf("expected effective release_id %d to stay pinned, got %d", oldRelease.Id, registry.ReleaseId)
	}

	out, err := service.List(ctx, ListInput{ID: pluginID})
	if err != nil {
		t.Fatalf("expected plugin list to succeed, got error: %v", err)
	}
	item := findPluginItem(out, pluginID)
	if item == nil {
		t.Fatal("expected dynamic plugin list item")
	}
	if item.RuntimeState != RuntimeUpgradeStatePendingUpgrade {
		t.Fatalf("expected runtime state %s, got %#v", RuntimeUpgradeStatePendingUpgrade, item)
	}
	if item.EffectiveVersion != oldVersion || item.DiscoveredVersion != newVersion {
		t.Fatalf("expected effective/discovered versions %s/%s, got %#v", oldVersion, newVersion, item)
	}
	if !item.UpgradeAvailable {
		t.Fatalf("expected dynamic plugin to report upgradeAvailable, got %#v", item)
	}

	detail, err := service.Get(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected plugin detail to succeed, got error: %v", err)
	}
	if detail.RuntimeState != RuntimeUpgradeStatePendingUpgrade {
		t.Fatalf("expected detail runtime state %s, got %#v", RuntimeUpgradeStatePendingUpgrade, detail)
	}
	if detail.EffectiveVersion != oldVersion || detail.DiscoveredVersion != newVersion {
		t.Fatalf("expected detail effective/discovered versions %s/%s, got %#v", oldVersion, newVersion, detail)
	}
	if !detail.UpgradeAvailable {
		t.Fatalf("expected detail to report upgradeAvailable, got %#v", detail)
	}
}

// TestListMarksInstalledDynamicPluginWithFailedTargetReleaseUpgradeFailed verifies
// failed target releases stay visible as retryable runtime-upgrade failures.
func TestListMarksInstalledDynamicPluginWithFailedTargetReleaseUpgradeFailed(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-dynamic-runtime-upgrade-failed"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Failed Plugin",
		oldVersion,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}

	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Failed Plugin",
		newVersion,
		nil,
		nil,
	)
	newManifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected new dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected new dynamic manifest sync to succeed, got error: %v", err)
	}

	targetRelease, err := service.getPluginRelease(ctx, pluginID, newVersion)
	if err != nil {
		t.Fatalf("expected target release lookup to succeed, got error: %v", err)
	}
	if targetRelease == nil {
		t.Fatal("expected target release row")
	}
	if err = service.storeSvc.UpdateReleaseState(
		ctx,
		targetRelease.Id,
		plugintypes.ReleaseStatusFailed,
		"",
	); err != nil {
		t.Fatalf("expected target release failure state update to succeed, got error: %v", err)
	}

	out, err := service.SyncAndList(ctx)
	if err != nil {
		t.Fatalf("expected sync-and-list to preserve failed target release, got error: %v", err)
	}
	item := findPluginItem(out, pluginID)
	if item == nil {
		t.Fatal("expected dynamic plugin list item")
	}
	if item.RuntimeState != RuntimeUpgradeStateUpgradeFailed {
		t.Fatalf("expected runtime state %s, got %#v", RuntimeUpgradeStateUpgradeFailed, item)
	}
	if item.LastUpgradeFailure == nil {
		t.Fatalf("expected last upgrade failure details, got %#v", item)
	}
	if item.LastUpgradeFailure.ReleaseID != targetRelease.Id ||
		item.LastUpgradeFailure.ReleaseVersion != newVersion {
		t.Fatalf("expected failed release %d/%s, got %#v", targetRelease.Id, newVersion, item.LastUpgradeFailure)
	}

	failedRelease, err := service.getPluginRelease(ctx, pluginID, newVersion)
	if err != nil {
		t.Fatalf("expected failed release lookup to succeed, got error: %v", err)
	}
	if failedRelease == nil || failedRelease.Status != plugintypes.ReleaseStatusFailed.String() {
		t.Fatalf("expected target release to remain failed after sync, got %#v", failedRelease)
	}
}

// TestFilterMenusHidesPendingUpgradePluginMenus verifies plugin-owned menus are
// hidden while a plugin waits for runtime upgrade.
func TestFilterMenusHidesPendingUpgradePluginMenus(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-dynamic-menu-pending-upgrade"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Menu Pending Upgrade Plugin",
		oldVersion,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}

	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Menu Pending Upgrade Plugin",
		newVersion,
		nil,
		nil,
	)
	newManifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected new dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected new dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = service.integrationSvc.RefreshEnabledSnapshot(ctx); err != nil {
		t.Fatalf("expected enabled snapshot refresh to succeed, got error: %v", err)
	}

	filtered := service.FilterMenus(ctx, []*entity.SysMenu{
		{
			Id:      1,
			MenuKey: "plugin:" + pluginID + ":entry",
			Name:    "runtime menu",
			Type:    catalog.MenuTypePage.String(),
			Status:  1,
			Visible: 1,
		},
	})
	if len(filtered) != 0 {
		t.Fatalf("expected pending-upgrade plugin menu to be hidden, got %d entries", len(filtered))
	}
}

// TestFilterMenusUsesAuthoritativeRegistryState verifies user-facing menu
// projection does not reuse stale process-local enablement snapshots after
// direct lifecycle-state changes have reached the persisted registry.
func TestFilterMenusUsesAuthoritativeRegistryState(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dev-source-menu-authoritative"
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
	service.integrationSvc.SetPluginEnabledState(pluginID, false)
	if err := service.storeSvc.SetPluginStatus(ctx, pluginID, statusflag.EnabledValue.Int()); err != nil {
		t.Fatalf("expected persisted plugin status update to succeed, got error: %v", err)
	}

	menus := []*entity.SysMenu{
		{
			Id:      1,
			MenuKey: "plugin:" + pluginID + ":entry",
			Name:    "source menu",
			Type:    catalog.MenuTypePage.String(),
			Status:  1,
			Visible: 1,
		},
	}
	filtered := service.FilterMenus(ctx, menus)
	if len(filtered) != 1 {
		t.Fatalf("expected enabled source plugin menu to stay visible, got %d entries", len(filtered))
	}
	filteredPermissions := service.FilterPermissionMenus(ctx, menus)
	if len(filteredPermissions) != 1 {
		t.Fatalf("expected enabled source plugin permission menu to stay visible, got %d entries", len(filteredPermissions))
	}
}

// TestListLocalizesUninstalledDynamicPluginMetadataInEnglish verifies that
// plugin management can display artifact-owned metadata before installation.
func TestListLocalizesUninstalledDynamicPluginMetadataInEnglish(t *testing.T) {
	var (
		service      = newTestService()
		ctx          = context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: i18nsvc.EnglishLocale})
		pluginID     = "plugin-dev-dynamic-list-i18n"
		artifactPath = filepath.Join(testutil.TestDynamicStorageDir(), pluginID+".wasm")
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		if err := os.Remove(artifactPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("failed to remove dynamic i18n test artifact %s: %v", artifactPath, err)
		}
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:          pluginID,
			Name:        "动态插件列表中文名",
			Version:     "v0.9.8",
			Type:        pluginv1.PluginTypeDynamic.String(),
			Description: "未安装动态插件的中文描述",
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        protocol.RuntimeKindWasm,
			ABIVersion:         protocol.SupportedABIVersion,
			FrontendAssetCount: len(testutil.DefaultTestRuntimeFrontendAssets()),
		},
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	appendRuntimeI18NSectionForPluginListTest(
		t,
		artifactPath,
		[]map[string]string{
			{
				"locale": "en-US",
				"content": `{
  "plugin": {
    "plugin-dev-dynamic-list-i18n": {
      "name": "Dynamic List I18N Plugin",
      "description": "English dynamic plugin description before installation."
    }
  }
}`,
			},
		},
	)

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic i18n artifact manifest to load, got error: %v", err)
	}
	registry, err := service.syncPluginManifest(ctx, manifest)
	if err != nil {
		t.Fatalf("expected dynamic i18n manifest sync to succeed, got error: %v", err)
	}
	if registry == nil || registry.Installed != statusflag.Uninstalled.Int() {
		t.Fatalf("expected dynamic i18n plugin to remain uninstalled after sync, got %#v", registry)
	}

	out, err := service.List(ctx, ListInput{ID: pluginID})
	if err != nil {
		t.Fatalf("expected plugin list to succeed, got error: %v", err)
	}
	item := findPluginItem(out, pluginID)
	if item == nil {
		t.Fatalf("expected dynamic i18n plugin to appear in plugin list")
	}
	if item.Name != "Dynamic List I18N Plugin" {
		t.Fatalf("expected English plugin name before install, got %q", item.Name)
	}
	if item.Description != "English dynamic plugin description before installation." {
		t.Fatalf("expected English plugin description before install, got %q", item.Description)
	}
	if item.Installed != statusflag.Uninstalled.Int() {
		t.Fatalf("expected plugin to remain not installed, got %d", item.Installed)
	}
}

// TestSyncAndListDoesNotRestoreUninstalledDynamicGovernanceProjection verifies
// that sync does not recreate release-bound governance after uninstall.
func TestSyncAndListDoesNotRestoreUninstalledDynamicGovernanceProjection(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dev-dynamic-uninstall-governance"
	)

	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithMenus(
		t,
		pluginID,
		"Dynamic Uninstall Governance Plugin",
		"v0.3.1",
		[]*catalog.MenuSpec{
			{
				Key:    "plugin:plugin-dev-dynamic-uninstall-governance:entry",
				Name:   "Dynamic Uninstall Governance Plugin",
				Path:   "plugin-dev-dynamic-uninstall-governance-entry",
				Perms:  "plugin-dev-dynamic-uninstall-governance:view",
				Icon:   "ant-design:appstore-outlined",
				Type:   catalog.MenuTypePage.String(),
				Sort:   1,
				Remark: "Runtime uninstall governance verification menu.",
			},
		},
		nil,
		nil,
	)

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}
	if err = service.Uninstall(ctx, pluginID, UninstallOptions{PurgeStorageData: true}); err != nil {
		t.Fatalf("expected dynamic plugin uninstall to succeed, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected runtime registry row to exist after uninstall")
	}
	if registry.ReleaseId != 0 {
		t.Fatalf("expected runtime registry release_id to be cleared after uninstall, got %d", registry.ReleaseId)
	}

	resourceCount, err := dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{PluginId: pluginID}).
		Count()
	if err != nil {
		t.Fatalf("expected governance resource count query to succeed, got error: %v", err)
	}
	if resourceCount != 0 {
		t.Fatalf("expected uninstall to clear governance resource refs, got count=%d", resourceCount)
	}

	if _, err = service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected sync-and-list to succeed after uninstall, got error: %v", err)
	}

	registry, err = service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup after sync-and-list to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected runtime registry row to remain after sync-and-list")
	}
	if registry.ReleaseId != 0 {
		t.Fatalf("expected sync-and-list not to restore release_id for uninstalled plugin, got %d", registry.ReleaseId)
	}

	resourceCount, err = dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{PluginId: pluginID}).
		Count()
	if err != nil {
		t.Fatalf("expected governance resource count query after sync-and-list to succeed, got error: %v", err)
	}
	if resourceCount != 0 {
		t.Fatalf("expected sync-and-list not to recreate governance resource refs for uninstalled plugin, got count=%d", resourceCount)
	}
}

// appendRuntimeI18NSectionForPluginListTest appends one runtime i18n custom
// section to the synthetic wasm artifact used by plugin list localization tests.
func appendRuntimeI18NSectionForPluginListTest(
	t *testing.T,
	artifactPath string,
	payload any,
) {
	t.Helper()

	content, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact read to succeed, got error: %v", err)
	}
	sectionPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("expected runtime i18n payload marshal to succeed, got error: %v", err)
	}
	content = appendPluginListTestWasmCustomSection(
		content,
		protocol.WasmSectionI18NAssets,
		sectionPayload,
	)
	if err = os.WriteFile(artifactPath, content, 0o644); err != nil {
		t.Fatalf("expected runtime artifact write to succeed, got error: %v", err)
	}
}

// appendPluginListTestWasmCustomSection appends one custom section using WASM
// section-length encoding.
func appendPluginListTestWasmCustomSection(content []byte, name string, payload []byte) []byte {
	sectionPayload := append([]byte{}, encodePluginListTestWasmULEB128(uint32(len(name)))...)
	sectionPayload = append(sectionPayload, []byte(name)...)
	sectionPayload = append(sectionPayload, payload...)

	result := append([]byte{}, content...)
	result = append(result, 0x00)
	result = append(result, encodePluginListTestWasmULEB128(uint32(len(sectionPayload)))...)
	result = append(result, sectionPayload...)
	return result
}

// encodePluginListTestWasmULEB128 encodes one unsigned integer for custom sections.
func encodePluginListTestWasmULEB128(value uint32) []byte {
	result := make([]byte, 0, 5)
	for {
		current := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			current |= 0x80
		}
		result = append(result, current)
		if value == 0 {
			return result
		}
	}
}
