// This file verifies that enabled dynamic plugin release artifacts contribute
// runtime i18n assets and disappear after lifecycle status changes.

package i18n

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/pluginbridge"
)

const testDynamicPluginI18NVersion = "v0.1.0"

// TestBuildRuntimeMessagesIncludesEnabledDynamicPluginAssets verifies that
// active dynamic plugin release artifacts participate in runtime i18n
// aggregation and follow enablement state changes.
func TestBuildRuntimeMessagesIncludesEnabledDynamicPluginAssets(t *testing.T) {
	resetRuntimeBundleCache()

	var (
		ctx      = context.Background()
		svc      = New()
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
		Type:           dynamicPluginType,
		RuntimeKind:    pluginbridge.RuntimeKindWasm,
		Status:         dynamicPluginReleaseStatusActive,
		PackagePath:    artifactPath,
		Checksum:       "dynamic-plugin-i18n-test-checksum",
	})
	pluginRowID := insertDynamicPluginRegistryForTest(t, ctx, do.SysPlugin{
		PluginId:     pluginID,
		Name:         "Dynamic Runtime Plugin",
		Version:      testDynamicPluginI18NVersion,
		Type:         dynamicPluginType,
		Installed:    dynamicPluginInstalledYes,
		Status:       dynamicPluginStatusEnabled,
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

// writeDynamicPluginI18NArtifactForTest writes one minimal wasm artifact carrying a plugin i18n section.
func writeDynamicPluginI18NArtifactForTest(t *testing.T, pluginID string, assets []*dynamicPluginI18NAsset) string {
	t.Helper()

	payload, err := json.Marshal(assets)
	if err != nil {
		t.Fatalf("marshal dynamic plugin i18n assets: %v", err)
	}

	content := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	content = appendTestWasmCustomSection(content, pluginbridge.WasmSectionI18NAssets, payload)

	artifactPath := filepath.Join(t.TempDir(), pluginID+".wasm")
	if err = os.WriteFile(artifactPath, content, 0o644); err != nil {
		t.Fatalf("write dynamic plugin i18n artifact: %v", err)
	}
	return artifactPath
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
