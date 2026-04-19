// This file contains unit tests for extension-point registration rules and
// published callback input contracts defined by the pluginhost package.

package pluginhost

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
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
	assertInterfaceType(t, (*HookPayload)(nil), "HookPayload")
	assertInterfaceType(t, (*AfterAuthInput)(nil), "AfterAuthInput")
	assertInterfaceType(t, (*SourcePluginUninstallInput)(nil), "SourcePluginUninstallInput")
	assertInterfaceType(t, (*RouteRegistrar)(nil), "RouteRegistrar")
	assertInterfaceType(t, (*CronRegistrar)(nil), "CronRegistrar")
	assertInterfaceType(t, (*MenuDescriptor)(nil), "MenuDescriptor")
	assertInterfaceType(t, (*PermissionDescriptor)(nil), "PermissionDescriptor")
}

// TestRegisterHookAcceptsAsyncMode verifies async execution is allowed for hook callbacks.
func TestRegisterHookAcceptsAsyncMode(t *testing.T) {
	plugin := NewSourcePlugin("test-plugin-hook")
	plugin.RegisterHook(
		ExtensionPointAuthLoginSucceeded,
		CallbackExecutionModeAsync,
		func(ctx context.Context, payload HookPayload) error {
			return nil
		},
	)

	items := plugin.GetHookHandlers()
	if len(items) != 1 {
		t.Fatalf("expected one hook handler, got %d", len(items))
	}
	if items[0].Mode != CallbackExecutionModeAsync {
		t.Fatalf("expected async mode, got %s", items[0].Mode)
	}
}

// TestRegisterRoutesRejectsAsyncMode verifies route registration remains a
// blocking-only extension point.
func TestRegisterRoutesRejectsAsyncMode(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected async route registration to panic")
		}
	}()

	plugin := NewSourcePlugin("test-plugin-route")
	plugin.RegisterRoutes(
		ExtensionPointHTTPRouteRegister,
		CallbackExecutionModeAsync,
		func(ctx context.Context, registrar RouteRegistrar) error {
			return nil
		},
	)
}

// TestCronRegistrarReportsPrimaryNode verifies cron registrars expose the
// current primary-node status from the host callback.
func TestCronRegistrarReportsPrimaryNode(t *testing.T) {
	registrar := NewCronRegistrar(
		"test-plugin",
		nil,
		func() bool { return false },
	)
	if registrar.IsPrimaryNode() {
		t.Fatalf("expected current node to be non-primary")
	}

	registrar = NewCronRegistrar(
		"test-plugin",
		nil,
		func() bool { return true },
	)
	if !registrar.IsPrimaryNode() {
		t.Fatalf("expected current node to be primary")
	}
}

// TestRegisterJobHandlerPublishesScheduledJobMetadata verifies source plugins
// can publish scheduled-job handlers for host-side registration.
func TestRegisterJobHandlerPublishesScheduledJobMetadata(t *testing.T) {
	plugin := NewSourcePlugin("test-plugin-job-handler")
	plugin.RegisterJobHandler(JobHandlerRegistration{
		Name:         "echo",
		DisplayName:  "Echo",
		Description:  "Echoes the payload for scheduled-job tests.",
		ParamsSchema: `{"type":"object","properties":{"message":{"type":"string"}},"required":["message"]}`,
		Handler: func(ctx context.Context, params json.RawMessage) (result any, err error) {
			return map[string]any{"params": string(params)}, nil
		},
	})

	items := plugin.GetJobHandlers()
	if len(items) != 1 {
		t.Fatalf("expected one job handler, got %d", len(items))
	}
	if items[0].Name != "echo" {
		t.Fatalf("expected handler name echo, got %s", items[0].Name)
	}
	if items[0].DisplayName != "Echo" {
		t.Fatalf("expected handler display name Echo, got %s", items[0].DisplayName)
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
	})
	if HookPayloadStringValue(authValues, HookPayloadKeyUserName) != "admin" {
		t.Fatalf("expected auth payload username to be published")
	}
	if HookPayloadStringValue(authValues, HookPayloadKeyClientType) != "web" {
		t.Fatalf("expected auth payload clientType to be published")
	}
}

// TestRegisterUninstallHandlerPublishesPolicySnapshot verifies uninstall
// handlers receive the host-confirmed policy snapshot interface.
func TestRegisterUninstallHandlerPublishesPolicySnapshot(t *testing.T) {
	plugin := NewSourcePlugin("test-plugin-uninstall")
	called := false

	plugin.RegisterUninstallHandler(func(ctx context.Context, input SourcePluginUninstallInput) error {
		called = true
		if input.PluginID() != "test-plugin-uninstall" {
			t.Fatalf("expected plugin id to be published, got %s", input.PluginID())
		}
		if !input.PurgeStorageData() {
			t.Fatalf("expected purgeStorageData to be true")
		}
		return nil
	})

	handler := plugin.GetUninstallHandler()
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

// assertInterfaceType verifies the reflected type under test is an interface.
func assertInterfaceType(t *testing.T, value interface{}, name string) {
	t.Helper()

	if reflect.TypeOf(value).Elem().Kind() != reflect.Interface {
		t.Fatalf("expected %s to be declared as interface", name)
	}
}
