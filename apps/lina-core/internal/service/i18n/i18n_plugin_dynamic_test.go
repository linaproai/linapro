// This file verifies that dynamic plugin release artifacts contribute runtime
// i18n assets and resolve package paths from the expected storage roots.

package i18n

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gogf/gf/v2/os/gctx"

	pluginv1 "lina-core/api/plugin/v1"
	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/statusflag"
)

const testDynamicPluginI18NVersion = "v0.1.0"

// dynamicPluginI18NConfigService wraps the real config service with a
// test-scoped dynamic plugin storage path.
type dynamicPluginI18NConfigService struct {
	configsvc.Service
	dynamicStoragePath string
}

// GetPlugin returns plugin config with the test-scoped dynamic storage path.
func (s dynamicPluginI18NConfigService) GetPlugin(ctx context.Context) *configsvc.PluginConfig {
	cfg := s.Service.GetPlugin(ctx)
	if cfg == nil {
		cfg = &configsvc.PluginConfig{}
	}
	if s.dynamicStoragePath != "" {
		cfg.Dynamic.StoragePath = s.dynamicStoragePath
	}
	return cfg
}

// GetPluginDynamicStoragePath returns the test-scoped dynamic storage path.
func (s dynamicPluginI18NConfigService) GetPluginDynamicStoragePath(ctx context.Context) string {
	if s.dynamicStoragePath != "" {
		return s.dynamicStoragePath
	}
	return s.Service.GetPluginDynamicStoragePath(ctx)
}

// TestBuildRuntimeMessagesIncludesEnabledDynamicPluginAssets verifies that
// active dynamic plugin release artifacts participate in runtime i18n
// aggregation and follow enablement state changes.
func TestBuildRuntimeMessagesIncludesEnabledDynamicPluginAssets(t *testing.T) {
	resetRuntimeBundleCache()

	var (
		ctx      = context.Background()
		svc      = New(bizctx.New(), configsvc.New(), cachecoord.Default(nil))
		pluginID = "plugin-i18n-dynamic-runtime"
		key      = "plugin.plugin-i18n-dynamic-runtime.name"
		value    = "Dynamic Runtime Plugin"
	)

	artifactPath := writeDynamicPluginI18NArtifactForTest(t, pluginID, []*dynamicPluginI18NAsset{
		{
			Locale:  EnglishLocale,
			Content: "{\"plugin.plugin-i18n-dynamic-runtime.name\":\"Dynamic Runtime Plugin\"}",
		},
	})
	releaseID := insertDynamicPluginReleaseForTest(t, ctx, do.SysPluginRelease{
		PluginId:       pluginID,
		ReleaseVersion: testDynamicPluginI18NVersion,
		Type:           pluginv1.PluginTypeDynamic.String(),
		RuntimeKind:    protocol.RuntimeKindWasm,
		Status:         dynamicPluginReleaseStatusActive,
		PackagePath:    artifactPath,
		Checksum:       "dynamic-plugin-i18n-test-checksum",
	})
	pluginRowID := insertDynamicPluginRegistryForTest(t, ctx, do.SysPlugin{
		PluginId:     pluginID,
		Name:         "Dynamic Runtime Plugin",
		Version:      testDynamicPluginI18NVersion,
		Type:         pluginv1.PluginTypeDynamic.String(),
		Installed:    statusflag.Installed.Int(),
		Status:       statusflag.EnabledValue.Int(),
		DesiredState: "enabled",
		CurrentState: "enabled",
		Generation:   int64(1),
		ReleaseId:    releaseID,
		Checksum:     "dynamic-plugin-i18n-test-checksum",
	})
	t.Cleanup(func() {
		deleteDynamicPluginRegistryByID(t, ctx, pluginRowID)
		deleteDynamicPluginReleaseByID(t, ctx, releaseID)
		resetRuntimeBundleCache()
	})

	messages := svc.BuildRuntimeMessages(ctx, EnglishLocale)
	if actual, ok := lookupMessageString(messages, key); !ok || actual != value {
		t.Fatalf("expected dynamic plugin translation %q, got %q (exists=%v)", value, actual, ok)
	}

	updateDynamicPluginLifecycleStateForTest(t, ctx, pluginRowID, 1, 0, "installed", "installed")
	resetRuntimeBundleCache()
	messages = svc.BuildRuntimeMessages(ctx, EnglishLocale)
	if _, ok := lookupMessageString(messages, key); ok {
		t.Fatalf("expected dynamic plugin translation %q to disappear after disable", key)
	}

	updateDynamicPluginLifecycleStateForTest(t, ctx, pluginRowID, 1, 1, "enabled", "enabled")
	resetRuntimeBundleCache()
	messages = svc.BuildRuntimeMessages(ctx, EnglishLocale)
	if actual, ok := lookupMessageString(messages, key); !ok || actual != value {
		t.Fatalf("expected dynamic plugin translation %q after re-enable, got %q (exists=%v)", value, actual, ok)
	}

	updateDynamicPluginLifecycleStateForTest(t, ctx, pluginRowID, 0, 0, "uninstalled", "uninstalled")
	resetRuntimeBundleCache()
	messages = svc.BuildRuntimeMessages(ctx, EnglishLocale)
	if _, ok := lookupMessageString(messages, key); ok {
		t.Fatalf("expected dynamic plugin translation %q to disappear after uninstall", key)
	}
}

// TestTranslateDynamicPluginSourceTextUsesReleaseArtifactBeforeEnable verifies
// pre-install review metadata can be localized from a dynamic-plugin artifact
// without adding inactive plugin resources to the global runtime bundle.
func TestTranslateDynamicPluginSourceTextUsesReleaseArtifactBeforeEnable(t *testing.T) {
	resetRuntimeBundleCache()

	var (
		ctx      = context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: DefaultLocale})
		svc      = New(bizctx.New(), configsvc.New(), cachecoord.Default(nil))
		pluginID = "plugin-i18n-dynamic-source-text"
		key      = "plugin.plugin-i18n-dynamic-source-text.preview.name"
	)

	artifactPath := writeDynamicPluginI18NArtifactForTest(t, pluginID, []*dynamicPluginI18NAsset{
		{
			Locale:  DefaultLocale,
			Content: `{"plugin":{"plugin-i18n-dynamic-source-text":{"preview":{"name":"动态插件预览"}}}}`,
		},
	})
	releaseID := insertDynamicPluginReleaseForTest(t, ctx, do.SysPluginRelease{
		PluginId:       pluginID,
		ReleaseVersion: testDynamicPluginI18NVersion,
		Type:           pluginv1.PluginTypeDynamic.String(),
		RuntimeKind:    protocol.RuntimeKindWasm,
		Status:         dynamicPluginReleaseStatusActive,
		PackagePath:    artifactPath,
		Checksum:       "dynamic-plugin-dev-source-text-test-checksum",
	})
	t.Cleanup(func() {
		deleteDynamicPluginReleaseByID(t, ctx, releaseID)
		resetRuntimeBundleCache()
	})

	actual := svc.TranslateDynamicPluginSourceText(ctx, pluginID, key, "Dynamic Plugin Preview")
	if actual != "动态插件预览" {
		t.Fatalf("expected pre-enable dynamic plugin translation, got %q", actual)
	}

	messages := svc.BuildRuntimeMessages(ctx, DefaultLocale)
	if _, ok := lookupMessageString(messages, key); ok {
		t.Fatalf("expected inactive dynamic plugin key %q to stay out of global runtime bundle", key)
	}
}

// TestTranslateDynamicPluginSourceTextReloadsLatestRelease verifies source-text
// translation does not keep a stale cross-request process cache when a newer
// dynamic plugin release is uploaded.
func TestTranslateDynamicPluginSourceTextReloadsLatestRelease(t *testing.T) {
	resetRuntimeBundleCache()

	var (
		ctx      = context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: DefaultLocale})
		svc      = New(bizctx.New(), configsvc.New(), cachecoord.Default(nil))
		pluginID = "plugin-i18n-dynamic-source-text-reload"
		key      = "plugin.plugin-i18n-dynamic-source-text-reload.preview.name"
	)

	firstArtifactPath := writeDynamicPluginI18NArtifactForTest(t, pluginID, []*dynamicPluginI18NAsset{
		{
			Locale:  DefaultLocale,
			Content: `{"plugin":{"plugin-i18n-dynamic-source-text-reload":{"preview":{"name":"旧动态插件预览"}}}}`,
		},
	})
	firstReleaseID := insertDynamicPluginReleaseForTest(t, ctx, do.SysPluginRelease{
		PluginId:       pluginID,
		ReleaseVersion: "v0.1.0",
		Type:           pluginv1.PluginTypeDynamic.String(),
		RuntimeKind:    protocol.RuntimeKindWasm,
		Status:         dynamicPluginReleaseStatusActive,
		PackagePath:    firstArtifactPath,
		Checksum:       "dynamic-plugin-dev-source-text-reload-test-checksum-1",
	})
	t.Cleanup(func() {
		deleteDynamicPluginReleaseByID(t, ctx, firstReleaseID)
		resetRuntimeBundleCache()
	})

	actual := svc.TranslateDynamicPluginSourceText(ctx, pluginID, key, "Dynamic Plugin Preview")
	if actual != "旧动态插件预览" {
		t.Fatalf("expected first dynamic plugin translation, got %q", actual)
	}

	secondArtifactPath := writeDynamicPluginI18NArtifactForTest(t, pluginID, []*dynamicPluginI18NAsset{
		{
			Locale:  DefaultLocale,
			Content: `{"plugin":{"plugin-i18n-dynamic-source-text-reload":{"preview":{"name":"新动态插件预览"}}}}`,
		},
	})
	secondReleaseID := insertDynamicPluginReleaseForTest(t, ctx, do.SysPluginRelease{
		PluginId:       pluginID,
		ReleaseVersion: "v0.2.0",
		Type:           pluginv1.PluginTypeDynamic.String(),
		RuntimeKind:    protocol.RuntimeKindWasm,
		Status:         dynamicPluginReleaseStatusActive,
		PackagePath:    secondArtifactPath,
		Checksum:       "dynamic-plugin-dev-source-text-reload-test-checksum-2",
	})
	t.Cleanup(func() {
		deleteDynamicPluginReleaseByID(t, ctx, secondReleaseID)
	})

	actual = svc.TranslateDynamicPluginSourceText(ctx, pluginID, key, "Dynamic Plugin Preview")
	if actual != "新动态插件预览" {
		t.Fatalf("expected latest dynamic plugin translation, got %q", actual)
	}
}

// TestTranslateDynamicPluginSourceTextFallsBackToStagingArtifact verifies
// inactive metadata localization can still use the current upload artifact when
// a stale registry release path is no longer readable.
func TestTranslateDynamicPluginSourceTextFallsBackToStagingArtifact(t *testing.T) {
	resetRuntimeBundleCache()

	var (
		ctx          = context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})
		pluginID     = "plugin-i18n-dynamic-source-text-staging"
		key          = "plugin.plugin-i18n-dynamic-source-text-staging.name"
		storageDir   = t.TempDir()
		stagingPath  = filepath.Join(storageDir, pluginID+".wasm")
		stalePackage = filepath.Join("missing-release", pluginID+".wasm")
	)
	svc := New(
		bizctx.New(),
		dynamicPluginI18NConfigService{Service: configsvc.New(), dynamicStoragePath: storageDir},
		cachecoord.Default(nil),
	)

	t.Cleanup(func() {
		resetRuntimeBundleCache()
	})

	writeDynamicPluginI18NArtifactAtPathForTest(t, stagingPath, []*dynamicPluginI18NAsset{
		{
			Locale:  EnglishLocale,
			Content: `{"plugin":{"plugin-i18n-dynamic-source-text-staging":{"name":"Dynamic Staging Plugin"}}}`,
		},
	})
	releaseID := insertDynamicPluginReleaseForTest(t, ctx, do.SysPluginRelease{
		PluginId:       pluginID,
		ReleaseVersion: testDynamicPluginI18NVersion,
		Type:           pluginv1.PluginTypeDynamic.String(),
		RuntimeKind:    protocol.RuntimeKindWasm,
		Status:         dynamicPluginReleaseStatusActive,
		PackagePath:    stalePackage,
		Checksum:       "dynamic-plugin-dev-source-text-staging-stale-checksum",
	})
	pluginRowID := insertDynamicPluginRegistryForTest(t, ctx, do.SysPlugin{
		PluginId:     pluginID,
		Name:         "动态暂存插件",
		Version:      testDynamicPluginI18NVersion,
		Type:         pluginv1.PluginTypeDynamic.String(),
		Installed:    0,
		Status:       0,
		DesiredState: "uninstalled",
		CurrentState: "uninstalled",
		Generation:   int64(1),
		ReleaseId:    releaseID,
		Checksum:     "dynamic-plugin-dev-source-text-staging-stale-checksum",
	})
	t.Cleanup(func() {
		deleteDynamicPluginRegistryByID(t, ctx, pluginRowID)
		deleteDynamicPluginReleaseByID(t, ctx, releaseID)
	})

	actual := svc.TranslateDynamicPluginSourceText(ctx, pluginID, key, "动态暂存插件")
	if actual != "Dynamic Staging Plugin" {
		t.Fatalf("expected staging artifact translation, got %q", actual)
	}
}

// writeDynamicPluginI18NArtifactForTest writes one minimal wasm artifact carrying a plugin i18n section.
func writeDynamicPluginI18NArtifactForTest(t *testing.T, pluginID string, assets []*dynamicPluginI18NAsset) string {
	t.Helper()

	artifactPath := filepath.Join(t.TempDir(), pluginID+".wasm")
	writeDynamicPluginI18NArtifactAtPathForTest(t, artifactPath, assets)
	return artifactPath
}

// writeDynamicPluginI18NArtifactAtPathForTest writes one minimal wasm artifact
// with plugin i18n assets to an explicit filesystem path.
func writeDynamicPluginI18NArtifactAtPathForTest(t *testing.T, artifactPath string, assets []*dynamicPluginI18NAsset) {
	t.Helper()

	payload, err := json.Marshal(assets)
	if err != nil {
		t.Fatalf("marshal dynamic plugin i18n assets: %v", err)
	}

	content := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	content = appendTestWasmCustomSection(content, protocol.WasmSectionI18NAssets, payload)

	if err = os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
		t.Fatalf("create dynamic plugin i18n artifact dir: %v", err)
	}
	if err = os.WriteFile(artifactPath, content, 0o644); err != nil {
		t.Fatalf("write dynamic plugin i18n artifact: %v", err)
	}
}

// appendTestWasmCustomSection appends one custom section to a synthetic wasm payload.
func appendTestWasmCustomSection(content []byte, name string, payload []byte) []byte {
	section := make([]byte, 0, len(name)+len(payload)+8)
	section = appendTestWasmULEB128(section, uint32(len(name)))
	section = append(section, []byte(name)...)
	section = append(section, payload...)

	content = append(content, 0x00)
	content = appendTestWasmULEB128(content, uint32(len(section)))
	content = append(content, section...)
	return content
}

// appendTestWasmULEB128 encodes one unsigned LEB128 integer for synthetic wasm test data.
func appendTestWasmULEB128(content []byte, value uint32) []byte {
	current := value
	for {
		part := byte(current & 0x7f)
		current >>= 7
		if current != 0 {
			part |= 0x80
		}
		content = append(content, part)
		if current == 0 {
			return content
		}
	}
}

// insertDynamicPluginReleaseForTest inserts one dynamic plugin release row for i18n tests.
func insertDynamicPluginReleaseForTest(t *testing.T, ctx context.Context, data do.SysPluginRelease) int {
	t.Helper()

	insertedID, err := dao.SysPluginRelease.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert dynamic plugin release: %v", err)
	}
	return int(insertedID)
}

// insertDynamicPluginRegistryForTest inserts one dynamic plugin registry row for i18n tests.
func insertDynamicPluginRegistryForTest(t *testing.T, ctx context.Context, data do.SysPlugin) int {
	t.Helper()

	insertedID, err := dao.SysPlugin.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert dynamic plugin registry: %v", err)
	}
	return int(insertedID)
}

// updateDynamicPluginLifecycleStateForTest updates one plugin registry row to emulate lifecycle transitions.
func updateDynamicPluginLifecycleStateForTest(
	t *testing.T,
	ctx context.Context,
	id int,
	installed int,
	status int,
	desiredState string,
	currentState string,
) {
	t.Helper()

	if _, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{Id: id}).
		Data(do.SysPlugin{
			Installed:    installed,
			Status:       status,
			DesiredState: desiredState,
			CurrentState: currentState,
		}).
		Update(); err != nil {
		t.Fatalf("update dynamic plugin lifecycle state: %v", err)
	}
}

// deleteDynamicPluginRegistryByID removes one dynamic plugin registry row used by i18n tests.
func deleteDynamicPluginRegistryByID(t *testing.T, ctx context.Context, id int) {
	t.Helper()

	if _, err := dao.SysPlugin.Ctx(ctx).Unscoped().Where(do.SysPlugin{Id: id}).Delete(); err != nil {
		t.Fatalf("delete dynamic plugin registry %d: %v", id, err)
	}
}

// deleteDynamicPluginReleaseByID removes one dynamic plugin release row used by i18n tests.
func deleteDynamicPluginReleaseByID(t *testing.T, ctx context.Context, id int) {
	t.Helper()

	if _, err := dao.SysPluginRelease.Ctx(ctx).Unscoped().Where(do.SysPluginRelease{Id: id}).Delete(); err != nil {
		t.Fatalf("delete dynamic plugin release %d: %v", id, err)
	}
}

// TestResolveDynamicPluginPackagePathAnchorsRelativeStoragePathAtRepoRoot verifies
// that runtime i18n resolves temp/output against the repository root instead of
// the apps/lina-core working directory.
func TestResolveDynamicPluginPackagePathAnchorsRelativeStoragePathAtRepoRoot(t *testing.T) {
	t.Helper()

	originalWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	repoRoot := t.TempDir()
	if err = os.WriteFile(filepath.Join(repoRoot, "go.work"), []byte("go 1.26\n"), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}

	var (
		storageRoot  = filepath.Join(repoRoot, "temp", "output")
		packagePath  = filepath.Join("releases", "linapro-demo-dynamic", "v0.1.0", "linapro-demo-dynamic.wasm")
		expectedPath = filepath.Join(storageRoot, packagePath)
	)
	if err = os.MkdirAll(filepath.Dir(expectedPath), 0o755); err != nil {
		t.Fatalf("create runtime storage dir: %v", err)
	}
	if err = os.WriteFile(expectedPath, []byte("wasm"), 0o644); err != nil {
		t.Fatalf("write runtime artifact: %v", err)
	}

	workingDir := filepath.Join(repoRoot, "apps", "lina-core")
	if err = os.MkdirAll(workingDir, 0o755); err != nil {
		t.Fatalf("create working directory: %v", err)
	}
	if err = os.Chdir(workingDir); err != nil {
		t.Fatalf("chdir to fake apps/lina-core: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWorkingDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	svc := New(
		bizctx.New(),
		dynamicPluginI18NConfigService{Service: configsvc.New(), dynamicStoragePath: "temp/output"},
		cachecoord.Default(nil),
	).(*serviceImpl)

	resolvedPath, err := svc.resolveDynamicPluginPackagePath(context.Background(), filepath.ToSlash(packagePath))
	if err != nil {
		t.Fatalf("resolve dynamic plugin package path: %v", err)
	}
	expectedRealPath, err := filepath.EvalSymlinks(expectedPath)
	if err != nil {
		t.Fatalf("eval expected path symlink: %v", err)
	}
	resolvedRealPath, err := filepath.EvalSymlinks(resolvedPath)
	if err != nil {
		t.Fatalf("eval resolved path symlink: %v", err)
	}
	if resolvedRealPath != expectedRealPath {
		t.Fatalf("expected resolved path %q, got %q", expectedRealPath, resolvedRealPath)
	}
}
