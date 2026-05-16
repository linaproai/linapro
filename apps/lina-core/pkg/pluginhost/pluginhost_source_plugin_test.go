// This file contains unit tests for extension-point registration rules and
// published callback input contracts defined by the pluginhost package.

package pluginhost

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"

	bridgecontract "lina-core/pkg/pluginbridge/contract"
)

// TestExtensionPointExecutionModes verifies hook and registrar points publish
// the expected execution-mode support matrix.
func TestExtensionPointExecutionModes(t *testing.T) {
	if !IsHookExtensionPoint(ExtensionPointAuthLoginSucceeded) {
		t.Fatalf(
			"expected %s to be hook extension point",
			ExtensionPointAuthLoginSucceeded,
		)
	}
	if !IsRegistrationExtensionPoint(ExtensionPointHTTPRouteRegister) {
		t.Fatalf(
			"expected %s to be registration extension point",
			ExtensionPointHTTPRouteRegister,
		)
	}
	if !IsExtensionPointExecutionModeSupported(
		ExtensionPointAuthLoginSucceeded,
		CallbackExecutionModeAsync,
	) {
		t.Fatalf(
			"expected %s to support %s mode",
			ExtensionPointAuthLoginSucceeded,
			CallbackExecutionModeAsync,
		)
	}
	if IsExtensionPointExecutionModeSupported(
		ExtensionPointHTTPRouteRegister,
		CallbackExecutionModeAsync,
	) {
		t.Fatalf(
			"expected %s to reject %s mode",
			ExtensionPointHTTPRouteRegister,
			CallbackExecutionModeAsync,
		)
	}
}

// TestCallbackInputContractsUseInterfaces verifies published callback inputs are
// exposed as interfaces rather than host concrete structs.
func TestCallbackInputContractsUseInterfaces(t *testing.T) {
	assertInterfaceType(t, (*SourcePlugin)(nil), "SourcePlugin")
	assertInterfaceType(t, (*SourcePluginAssets)(nil), "SourcePluginAssets")
	assertInterfaceType(t, (*SourcePluginLifecycle)(nil), "SourcePluginLifecycle")
	assertInterfaceType(t, (*SourcePluginHooks)(nil), "SourcePluginHooks")
	assertInterfaceType(t, (*SourcePluginHTTP)(nil), "SourcePluginHTTP")
	assertInterfaceType(t, (*SourcePluginCron)(nil), "SourcePluginCron")
	assertInterfaceType(t, (*SourcePluginGovernance)(nil), "SourcePluginGovernance")
	assertInterfaceType(t, (*SourcePluginDefinition)(nil), "SourcePluginDefinition")
	assertInterfaceType(t, (*HookPayload)(nil), "HookPayload")
	assertInterfaceType(t, (*SourcePluginLifecycleInput)(nil), "SourcePluginLifecycleInput")
	assertInterfaceType(t, (*SourcePluginTenantLifecycleInput)(nil), "SourcePluginTenantLifecycleInput")
	assertInterfaceType(t, (*SourcePluginInstallModeChangeInput)(nil), "SourcePluginInstallModeChangeInput")
	assertInterfaceType(t, (*ManifestSnapshot)(nil), "ManifestSnapshot")
	assertInterfaceType(t, (*SourcePluginUpgradeInput)(nil), "SourcePluginUpgradeInput")
	assertInterfaceType(t, (*SourcePluginUninstallInput)(nil), "SourcePluginUninstallInput")
	assertInterfaceType(t, (*HTTPRegistrar)(nil), "HTTPRegistrar")
	assertInterfaceType(t, (*RouteRegistrar)(nil), "RouteRegistrar")
	assertInterfaceType(t, (*GlobalMiddlewareRegistrar)(nil), "GlobalMiddlewareRegistrar")
	assertInterfaceType(t, (*CronRegistrar)(nil), "CronRegistrar")
	assertInterfaceType(t, (*MenuDescriptor)(nil), "MenuDescriptor")
	assertInterfaceType(t, (*PermissionDescriptor)(nil), "PermissionDescriptor")
}

// TestRegisterHookAcceptsAsyncMode verifies async execution is allowed for hook callbacks.
func TestRegisterHookAcceptsAsyncMode(t *testing.T) {
	plugin := NewSourcePlugin("test-plugin-hook")
	if err := plugin.Hooks().RegisterHook(
		ExtensionPointAuthLoginSucceeded,
		CallbackExecutionModeAsync,
		func(ctx context.Context, payload HookPayload) error {
			return nil
		},
	); err != nil {
		t.Fatalf("expected hook registration to succeed, got %v", err)
	}

	items := mustSourcePluginDefinition(t, plugin).GetHookHandlers()
	if len(items) != 1 {
		t.Fatalf("expected one hook handler, got %d", len(items))
	}
	if items[0].Mode != CallbackExecutionModeAsync {
		t.Fatalf("expected async mode, got %s", items[0].Mode)
	}
}

// TestRegisterRoutesRejectsAsyncMode verifies route registration returns an
// error when the caller requests an unsupported execution mode.
func TestRegisterRoutesRejectsAsyncMode(t *testing.T) {
	plugin := NewSourcePlugin("test-plugin-route")
	err := plugin.HTTP().RegisterRoutes(
		ExtensionPointHTTPRouteRegister,
		CallbackExecutionModeAsync,
		func(ctx context.Context, registrar HTTPRegistrar) error {
			return nil
		},
	)
	if err == nil {
		t.Fatalf("expected async route registration to return an error")
	}
}

// TestRegisterSourcePluginForTestReturnsGoFrameError verifies test fixture
// registration errors preserve GoFrame stack information.
func TestRegisterSourcePluginForTestReturnsGoFrameError(t *testing.T) {
	t.Parallel()

	cleanup, err := RegisterSourcePluginForTest(nil)
	if cleanup != nil {
		t.Fatalf("expected cleanup to be nil for invalid source plugin")
	}
	if err == nil {
		t.Fatalf("expected invalid source plugin to return an error")
	}
	stack := gerror.Stack(err)
	if !strings.Contains(stack, "RegisterSourcePluginForTest") {
		t.Fatalf("expected GoFrame stack to include registration helper, got %q", stack)
	}
}

// TestCronRegistrarReportsPrimaryNode verifies cron registrars expose the
// current primary-node status from the host callback.
func TestCronRegistrarReportsPrimaryNode(t *testing.T) {
	registrar := NewCronRegistrar(
		"test-plugin",
		nil,
		func() bool { return false },
		nil,
	)
	if registrar.IsPrimaryNode() {
		t.Fatalf("expected current node to be non-primary")
	}

	registrar = NewCronRegistrar(
		"test-plugin",
		nil,
		func() bool { return true },
		nil,
	)
	if !registrar.IsPrimaryNode() {
		t.Fatalf("expected current node to be primary")
	}
}

// TestHookPayloadHelpersBuildPublishedKeys verifies published hook payload maps
// contain the expected field names and values.
func TestHookPayloadHelpersBuildPublishedKeys(t *testing.T) {
	status := 1

	lifecycleValues := BuildPluginLifecycleHookPayloadValues(PluginLifecycleHookPayloadInput{
		PluginID: "plugin-demo",
		Name:     "Plugin Demo",
		Version:  "v0.1.0",
		Status:   &status,
	})
	if HookPayloadStringValue(lifecycleValues, HookPayloadKeyPluginID) != "plugin-demo" {
		t.Fatalf("expected lifecycle payload plugin id to be published")
	}
	if actualStatus, ok := HookPayloadIntValue(lifecycleValues, HookPayloadKeyStatus); !ok || actualStatus != status {
		t.Fatalf("expected lifecycle payload status=%d, got %d ok=%v", status, actualStatus, ok)
	}

	authValues := BuildAuthHookPayloadValues(AuthHookPayloadInput{
		UserName:   "admin",
		Status:     1,
		IP:         "127.0.0.1",
		ClientType: "web",
		Browser:    "Chrome",
		OS:         "macOS",
		Message:    "login ok",
		Reason:     AuthHookReasonLoginSuccessful,
	})
	if HookPayloadStringValue(authValues, HookPayloadKeyUserName) != "admin" {
		t.Fatalf("expected auth payload username to be published")
	}
	if HookPayloadStringValue(authValues, HookPayloadKeyClientType) != "web" {
		t.Fatalf("expected auth payload clientType to be published")
	}
	if HookPayloadStringValue(authValues, HookPayloadKeyReason) != AuthHookReasonLoginSuccessful {
		t.Fatalf("expected auth payload reason to be published")
	}
}

// TestRegisterUninstallHandlerPublishesPolicySnapshot verifies uninstall
// handlers receive the host-confirmed policy snapshot interface.
func TestRegisterUninstallHandlerPublishesPolicySnapshot(t *testing.T) {
	plugin := NewSourcePlugin("test-plugin-uninstall")
	called := false

	if err := plugin.Lifecycle().RegisterUninstallHandler(func(ctx context.Context, input SourcePluginUninstallInput) error {
		called = true
		if input.PluginID() != "test-plugin-uninstall" {
			t.Fatalf("expected plugin id to be published, got %s", input.PluginID())
		}
		if !input.PurgeStorageData() {
			t.Fatalf("expected purgeStorageData to be true")
		}
		return nil
	}); err != nil {
		t.Fatalf("expected uninstall handler registration to succeed, got %v", err)
	}

	handler := mustSourcePluginDefinition(t, plugin).GetUninstallHandler()
	if handler == nil {
		t.Fatalf("expected uninstall handler to be registered")
	}
	if err := handler(context.Background(), NewSourcePluginUninstallInput("test-plugin-uninstall", true)); err != nil {
		t.Fatalf("expected uninstall handler to execute without error, got %v", err)
	}
	if !called {
		t.Fatalf("expected uninstall handler to be called")
	}
}

// TestRegisterUpgradeHandlersPublishesManifestSnapshots verifies source-plugin
// upgrade callbacks receive stable manifest snapshot interfaces.
func TestRegisterUpgradeHandlersPublishesManifestSnapshots(t *testing.T) {
	plugin := NewSourcePlugin("test-plugin-upgrade")
	called := false

	if err := plugin.Lifecycle().RegisterUpgradeHandler(func(ctx context.Context, input SourcePluginUpgradeInput) error {
		called = true
		if input.PluginID() != "test-plugin-upgrade" {
			t.Fatalf("expected upgrade plugin id to be published, got %s", input.PluginID())
		}
		if input.FromVersion() != "v0.1.0" || input.ToVersion() != "v0.2.0" {
			t.Fatalf("expected version pair v0.1.0/v0.2.0, got %s/%s", input.FromVersion(), input.ToVersion())
		}
		if input.FromManifest().Version() != "v0.1.0" || input.ToManifest().Version() != "v0.2.0" {
			t.Fatalf("expected manifest snapshot versions to be published")
		}
		if input.ToManifest().Values().MenuCount != 2 {
			t.Fatalf("expected manifest values copy to include menuCount")
		}
		return nil
	}); err != nil {
		t.Fatalf("expected upgrade handler registration to succeed, got %v", err)
	}

	handler := mustSourcePluginDefinition(t, plugin).GetUpgradeHandler()
	if handler == nil {
		t.Fatalf("expected upgrade handler to be registered")
	}
	input := NewSourcePluginUpgradeInput(
		"test-plugin-upgrade",
		"v0.1.0",
		"v0.2.0",
		NewManifestSnapshot(&bridgecontract.ManifestSnapshotV1{
			ID:      "test-plugin-upgrade",
			Name:    "Test Plugin Upgrade",
			Version: "v0.1.0",
			Type:    "source",
		}),
		NewManifestSnapshot(&bridgecontract.ManifestSnapshotV1{
			ID:        "test-plugin-upgrade",
			Name:      "Test Plugin Upgrade",
			Version:   "v0.2.0",
			Type:      "source",
			MenuCount: 2,
		}),
	)
	if err := handler(context.Background(), input); err != nil {
		t.Fatalf("expected upgrade handler to execute without error, got %v", err)
	}
	if !called {
		t.Fatalf("expected upgrade handler to be called")
	}
}

// TestNewManifestSnapshotUsesBridgeContract verifies source-plugin snapshots
// use the shared typed bridge lifecycle contract.
func TestNewManifestSnapshotUsesBridgeContract(t *testing.T) {
	input := &bridgecontract.ManifestSnapshotV1{
		ID:          "test-plugin-typed-snapshot",
		Name:        "Test Plugin Typed Snapshot",
		Version:     "v1.0.0",
		Type:        "source",
		Description: "typed contract",
	}
	snapshot := NewManifestSnapshot(input)
	input.Description = "mutated"
	values := snapshot.Values()
	values.Description = "mutated again"

	if snapshot.ID() != "test-plugin-typed-snapshot" ||
		snapshot.Name() != "Test Plugin Typed Snapshot" ||
		snapshot.Version() != "v1.0.0" ||
		snapshot.Type() != "source" ||
		snapshot.Values().Description != "typed contract" {
		t.Fatalf("expected typed bridge contract to be copied, got %#v", snapshot.Values())
	}
}

// TestNewManifestSnapshotReturnsNilForMissingContract verifies absent snapshots
// stay absent instead of creating empty wrappers.
func TestNewManifestSnapshotReturnsNilForMissingContract(t *testing.T) {
	if snapshot := NewManifestSnapshot(nil); snapshot != nil {
		t.Fatalf("expected nil manifest snapshot, got %#v", snapshot)
	}
}

// TestSourcePluginLifecycleCallbackAdapterRunsBeforeUpgrade verifies lifecycle
// facade callbacks are adapted into the shared callback runner.
func TestSourcePluginLifecycleCallbackAdapterRunsBeforeUpgrade(t *testing.T) {
	plugin := NewSourcePlugin("test-plugin-before-upgrade")
	if err := plugin.Lifecycle().RegisterBeforeUpgradeHandler(func(ctx context.Context, input SourcePluginUpgradeInput) (bool, string, error) {
		if input.PluginID() != "test-plugin-before-upgrade" {
			t.Fatalf("expected plugin id to be published, got %s", input.PluginID())
		}
		return false, "plugin.test.beforeUpgrade.blocked", nil
	}); err != nil {
		t.Fatalf("expected before-upgrade handler registration to succeed, got %v", err)
	}

	result := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook: LifecycleHookBeforeUpgrade,
		UpgradeInput: NewSourcePluginUpgradeInput(
			"test-plugin-before-upgrade",
			"v0.1.0",
			"v0.2.0",
			nil,
			nil,
		),
		Participants: []LifecycleParticipant{
			{
				PluginID: "test-plugin-before-upgrade",
				Callback: NewSourcePluginLifecycleCallbackAdapter(mustSourcePluginDefinition(t, plugin)),
			},
		},
	})
	if result.OK {
		t.Fatalf("expected before-upgrade callback to veto")
	}
	if len(result.Decisions) != 1 || result.Decisions[0].Reason != "plugin.test.beforeUpgrade.blocked" {
		t.Fatalf("expected veto reason to be preserved, got %#v", result.Decisions)
	}
}

// TestSourcePluginLifecycleCallbackAdapterRunsUpgrade verifies custom upgrade
// callbacks are exposed through the shared lifecycle runner.
func TestSourcePluginLifecycleCallbackAdapterRunsUpgrade(t *testing.T) {
	plugin := NewSourcePlugin("test-plugin-upgrade")
	called := false
	if err := plugin.Lifecycle().RegisterUpgradeHandler(func(ctx context.Context, input SourcePluginUpgradeInput) error {
		called = true
		if input.PluginID() != "test-plugin-upgrade" {
			t.Fatalf("expected plugin id to be published, got %s", input.PluginID())
		}
		if input.FromVersion() != "v0.1.0" || input.ToVersion() != "v0.2.0" {
			t.Fatalf("expected upgrade versions v0.1.0 -> v0.2.0, got %s -> %s", input.FromVersion(), input.ToVersion())
		}
		return nil
	}); err != nil {
		t.Fatalf("expected upgrade handler registration to succeed, got %v", err)
	}

	result := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook: LifecycleHookUpgrade,
		UpgradeInput: NewSourcePluginUpgradeInput(
			"test-plugin-upgrade",
			"v0.1.0",
			"v0.2.0",
			nil,
			nil,
		),
		Participants: []LifecycleParticipant{
			{
				PluginID: "test-plugin-upgrade",
				Callback: NewSourcePluginLifecycleCallbackAdapter(mustSourcePluginDefinition(t, plugin)),
			},
		},
	})
	if !result.OK || len(result.Decisions) != 1 || !called {
		t.Fatalf("expected upgrade callback to run successfully, result=%#v called=%v", result, called)
	}
}

// TestSourcePluginLifecycleCallbackAdapterRunsAfterInstall verifies source
// plugin After* facade callbacks are adapted into the shared callback runner.
func TestSourcePluginLifecycleCallbackAdapterRunsAfterInstall(t *testing.T) {
	plugin := NewSourcePlugin("test-plugin-after-install")
	called := false
	if err := plugin.Lifecycle().RegisterAfterInstallHandler(func(ctx context.Context, input SourcePluginLifecycleInput) error {
		called = true
		if input.PluginID() != "test-plugin-after-install" {
			t.Fatalf("expected plugin id to be published, got %s", input.PluginID())
		}
		if input.Operation() != LifecycleHookAfterInstall.String() {
			t.Fatalf("expected operation %s, got %s", LifecycleHookAfterInstall, input.Operation())
		}
		return nil
	}); err != nil {
		t.Fatalf("expected after-install handler registration to succeed, got %v", err)
	}

	result := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook:        LifecycleHookAfterInstall,
		PluginInput: NewSourcePluginLifecycleInput("test-plugin-after-install", LifecycleHookAfterInstall.String()),
		Participants: []LifecycleParticipant{
			{
				PluginID: "test-plugin-after-install",
				Callback: NewSourcePluginLifecycleCallbackAdapter(mustSourcePluginDefinition(t, plugin)),
			},
		},
	})
	if !result.OK || len(result.Decisions) != 1 || !called {
		t.Fatalf("expected after-install callback to run successfully, result=%#v called=%v", result, called)
	}
}

// TestSourcePluginLifecycleCallbackAdapterRunsInstallModeChange verifies
// install-mode precondition callbacks are exposed through the lifecycle facade.
func TestSourcePluginLifecycleCallbackAdapterRunsInstallModeChange(t *testing.T) {
	plugin := NewSourcePlugin("test-plugin-before-install-mode")
	if err := plugin.Lifecycle().RegisterBeforeInstallModeChangeHandler(func(
		ctx context.Context,
		input SourcePluginInstallModeChangeInput,
	) (bool, string, error) {
		if input.PluginID() != "test-plugin-before-install-mode" {
			t.Fatalf("expected plugin id to be published, got %s", input.PluginID())
		}
		if input.FromMode() != "global" || input.ToMode() != "tenant_scoped" {
			t.Fatalf("expected install mode transition to be published, got %s -> %s", input.FromMode(), input.ToMode())
		}
		return false, "plugin.test.installMode.blocked", nil
	}); err != nil {
		t.Fatalf("expected install-mode handler registration to succeed, got %v", err)
	}

	result := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook: LifecycleHookBeforeInstallModeChange,
		InstallModeInput: NewSourcePluginInstallModeChangeInput(
			"test-plugin-before-install-mode",
			LifecycleHookBeforeInstallModeChange.String(),
			"global",
			"tenant_scoped",
		),
		Participants: []LifecycleParticipant{
			{
				PluginID: "test-plugin-before-install-mode",
				Callback: NewSourcePluginLifecycleCallbackAdapter(mustSourcePluginDefinition(t, plugin)),
			},
		},
	})
	if result.OK {
		t.Fatalf("expected before-install-mode callback to veto")
	}
	if len(result.Decisions) != 1 || result.Decisions[0].Reason != "plugin.test.installMode.blocked" {
		t.Fatalf("expected veto reason to be preserved, got %#v", result.Decisions)
	}
}

// assertInterfaceType verifies the reflected type under test is an interface.
func assertInterfaceType(t *testing.T, value interface{}, name string) {
	t.Helper()

	if reflect.TypeOf(value).Elem().Kind() != reflect.Interface {
		t.Fatalf("expected %s to be declared as interface", name)
	}
}

// mustSourcePluginDefinition narrows one published SourcePlugin to the host
// definition view used by registry and integration code.
func mustSourcePluginDefinition(t *testing.T, plugin SourcePlugin) SourcePluginDefinition {
	t.Helper()

	definition, ok := plugin.(SourcePluginDefinition)
	if !ok {
		t.Fatalf("expected source plugin to implement SourcePluginDefinition")
	}
	return definition
}
