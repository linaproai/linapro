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
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/pluginhost"
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
	// ownerCapabilities indexes plugin-owned dynamic capability descriptors.
	ownerCapabilities *capregistry.Registry
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
		validDeps.ownerCapabilities,
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
			name:    "owner-capabilities",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.ownerCapabilities = nil },
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
				deps.ownerCapabilities,
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
		recoveredDeps.ownerCapabilities,
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
		notifySvc:         notifysvc.New(tenantspi.New(nil, nil, nil, bizCtxSvc)),
		hostServices:      newRootTestCapabilities(bizCtxSvc, nil),
		ownerCapabilities: capregistry.NewRegistry(),
		configFactory:     NewPluginConfigFactory("", ""),
		hostConfigSvc:     NewHostConfigService(configSvc),
		manifestFactory:   manifestresource.NewFactory(""),
	}
}

// TestBuildSourceCapabilityRegistryIndexesDeclaredOwnerDescriptors verifies
// startup registry construction uses source-plugin descriptor declarations.
func TestBuildSourceCapabilityRegistryIndexesDeclaredOwnerDescriptors(t *testing.T) {
	sourcePlugin := pluginhost.NewDeclarations("plugin-dev-source-registry")
	descriptor := capregistry.Descriptor{
		OwnerPluginID: "plugin-dev-source-registry",
		Service:       "workflow",
		Version:       "v1",
		Methods: []capregistry.MethodDescriptor{
			{
				Method:       "run.execute",
				Capability:   "framework.workflow.v1",
				Risk:         capregistry.RiskLevelExecute,
				ResourceKind: capregistry.ResourceKindNone,
			},
		},
	}
	if err := sourcePlugin.Providers().ProvideCapability(descriptor); err != nil {
		t.Fatalf("declare owner capability descriptor: %v", err)
	}
	cleanup, err := pluginhost.RegisterSourcePluginForTest(sourcePlugin)
	if err != nil {
		t.Fatalf("register source plugin for test: %v", err)
	}
	t.Cleanup(cleanup)

	registry, err := buildSourceCapabilityRegistry()
	if err != nil {
		t.Fatalf("build source capability registry: %v", err)
	}
	method, ok := registry.LookupMethod("plugin-dev-source-registry", "workflow", "v1", "run.execute")
	if !ok {
		t.Fatal("expected source descriptor method to be indexed")
	}
	if method.OwnerPluginID != "plugin-dev-source-registry" || method.Service != "workflow" {
		t.Fatalf("unexpected source method index: %#v", method)
	}
}

// TestValidateSourceCapabilityDescriptorOwnerRejectsMismatch verifies startup
// descriptor validation rejects descriptors not owned by their declaring plugin.
func TestValidateSourceCapabilityDescriptorOwnerRejectsMismatch(t *testing.T) {
	descriptor := capregistry.Descriptor{
		OwnerPluginID: "linapro-ai-core",
		Service:       "ai",
		Version:       "v1",
		Methods: []capregistry.MethodDescriptor{
			{
				Method:       "text.generate",
				Capability:   "plugin.linapro-ai-core.ai.v1",
				Risk:         capregistry.RiskLevelExecute,
				ResourceKind: capregistry.ResourceKindNone,
			},
		},
	}

	err := validateSourceCapabilityDescriptorOwner("plugin-dev-source-owner-mismatch", descriptor)
	if err == nil {
		t.Fatal("expected mismatched source descriptor owner to fail")
	}
	for _, want := range []string{"plugin-dev-source-owner-mismatch", "linapro-ai-core", "ai", "v1"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to contain %q, got %v", want, err)
		}
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
