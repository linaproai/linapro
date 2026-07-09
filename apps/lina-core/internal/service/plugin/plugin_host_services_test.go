// This file verifies root-level host service and Wasm host service wiring.

package plugin

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
	notifysvc "lina-core/internal/service/notify"
	"lina-core/internal/service/plugin/internal/manifestresource"
	"lina-core/pkg/dialect"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// TestStorageProviderRuntimeUsesInternalStateReader verifies provider storage
// gating still uses the host-internal bool plugin state contract.
func TestStorageProviderRuntimeUsesInternalStateReader(t *testing.T) {
	runtime := NewStorageProviderRuntime(testPluginStateReader{
		providers: map[string]bool{" provider-a ": true},
	})

	if !runtime.ProviderPluginAvailable(context.Background(), " provider-a ") {
		t.Fatal("expected provider-a to be available through internal state reader")
	}
	if runtime.ProviderPluginAvailable(context.Background(), "provider-b") {
		t.Fatal("expected missing provider to be unavailable")
	}
}

type testPluginStateReader struct {
	providers map[string]bool
}

func (r testPluginStateReader) IsEnabled(context.Context, string) bool {
	return false
}

func (r testPluginStateReader) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return r.providers[pluginID]
}

func (r testPluginStateReader) IsEnabledAuthoritative(context.Context, string) bool {
	return false
}

// wasmHostServiceTestDeps groups explicit dependencies for the root Wasm host
// service runtime constructor.
type wasmHostServiceTestDeps struct {
	// notifySvc is the notification service shared with dynamic plugins.
	notifySvc notifysvc.Service
	// hostServices provides plugin capability directories for host service dispatch.
	hostServices capability.Services
	// configFactory creates plugin-scoped config service views.
	configFactory PluginConfigFactory
	// hostConfigSvc exposes authorized host configuration reads.
	hostConfigSvc hostconfigcap.Service
	// manifestFactory creates plugin-scoped manifest service views.
	manifestFactory manifestresource.Factory
}

// TestNewWasmHostServiceRuntimeRequiresExplicitDependencies verifies the root
// startup runtime constructor rejects missing required services before
// dispatchers can use package defaults.
func TestNewWasmHostServiceRuntimeRequiresExplicitDependencies(t *testing.T) {
	validDeps := newWasmHostServiceTestDeps(t)
	if _, err := newWasmHostServiceRuntime(
		validDeps.hostServices,
		validDeps.configFactory,
		validDeps.hostConfigSvc,
		validDeps.manifestFactory,
	); err != nil {
		t.Fatalf("expected complete wasm host service runtime construction to succeed, got error: %v", err)
	}

	cases := []struct {
		name    string
		mutate  func(*wasmHostServiceTestDeps)
		message string
	}{
		{
			name:    "domain-capabilities",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.hostServices = nil },
			message: "create wasm host service runtime failed",
		},
		{
			name:    "config",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.configFactory = nil },
			message: "create wasm host service runtime failed",
		},
		{
			name:    "host-config",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.hostConfigSvc = nil },
			message: "create wasm host service runtime failed",
		},
		{
			name:    "manifest",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.manifestFactory = nil },
			message: "create wasm host service runtime failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := newWasmHostServiceTestDeps(t)
			tc.mutate(deps)
			_, err := newWasmHostServiceRuntime(
				deps.hostServices,
				deps.configFactory,
				deps.hostConfigSvc,
				deps.manifestFactory,
			)
			if err == nil {
				t.Fatalf("expected missing %s dependency to fail", tc.name)
			}
			if !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("expected error to contain %q, got %v", tc.message, err)
			}
		})
	}

	recoveredDeps := newWasmHostServiceTestDeps(t)
	if _, err := newWasmHostServiceRuntime(
		recoveredDeps.hostServices,
		recoveredDeps.configFactory,
		recoveredDeps.hostConfigSvc,
		recoveredDeps.manifestFactory,
	); err != nil {
		t.Fatalf("expected complete wasm host service runtime construction after nil cases, got error: %v", err)
	}
}

// newWasmHostServiceTestDeps builds a complete dependency set for root Wasm
// host service configuration tests.
func newWasmHostServiceTestDeps(t *testing.T) *wasmHostServiceTestDeps {
	t.Helper()

	configSvc := configsvc.New()
	bizCtxSvc := bizctx.New()
	return &wasmHostServiceTestDeps{
		notifySvc:       notifysvc.New(tenantspi.New(nil, nil, nil, bizCtxSvc)),
		hostServices:    newRootTestCapabilities(bizCtxSvc, nil),
		configFactory:   NewPluginConfigFactory("", ""),
		hostConfigSvc:   NewHostConfigService(configSvc),
		manifestFactory: manifestresource.NewFactory(""),
	}
}

// TestNormalizeDataTableNamesTrimsAndDeduplicates verifies metadata lookups
// query each non-empty table name at most once.
func TestNormalizeDataTableNamesTrimsAndDeduplicates(t *testing.T) {
	got := normalizeDataTableNames([]string{" sys_plugin ", "", "sys_user", "sys_plugin", "  "})
	want := []string{"sys_plugin", "sys_user"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected normalized table names %v, got %v", want, got)
	}
}

// TestDataTableCommentsFromMetadataMapsNonBlankComments verifies dialect
// metadata results are converted into the governance display map.
func TestDataTableCommentsFromMetadataMapsNonBlankComments(t *testing.T) {
	got := dataTableCommentsFromMetadata([]dialect.TableMeta{
		{TableName: " sys_plugin ", TableComment: " Plugin registry "},
		{TableName: "sys_user", TableComment: ""},
		{TableName: "", TableComment: "ignored"},
	})
	want := map[string]string{"sys_plugin": "Plugin registry"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected metadata comments %v, got %v", want, got)
	}
}
