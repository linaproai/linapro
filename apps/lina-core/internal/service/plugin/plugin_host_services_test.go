// This file verifies root-level host service and Wasm host service wiring.

package plugin

import (
	"reflect"
	"strings"
	"testing"

	"lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
	notifysvc "lina-core/internal/service/notify"
	"lina-core/pkg/dialect"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	capabilityhostconfig "lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilitymanifest "lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/plugincap"
	capabilityconfig "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// wasmHostServiceTestDeps groups explicit dependencies for the root Wasm host
// service runtime constructor.
type wasmHostServiceTestDeps struct {
	// notifySvc is the notification service shared with dynamic plugins.
	notifySvc notifysvc.Service
	// hostServices provides plugin capability directories for host service dispatch.
	hostServices capability.Services
	// configFactory creates plugin-scoped config service views.
	configFactory plugincap.ConfigServiceFactory
	// hostConfigSvc exposes authorized host configuration reads.
	hostConfigSvc hostconfigcap.Service
	// manifestFactory creates plugin-scoped manifest service views.
	manifestFactory manifestcap.ServiceFactory
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
		notifySvc:       notifysvc.New(tenantspi.New(nil, nil, bizCtxSvc)),
		hostServices:    newRootTestCapabilities(bizCtxSvc, nil),
		configFactory:   capabilityconfig.NewConfigFactory("", ""),
		hostConfigSvc:   capabilityhostconfig.New(mustHostConfigRawReaderForTest(t, configSvc)),
		manifestFactory: capabilitymanifest.NewFactory(""),
	}
}

// mustHostConfigRawReaderForTest returns the raw host-config reader implemented
// by the test config service or fails the current test during fixture wiring.
func mustHostConfigRawReaderForTest(t *testing.T, configSvc configsvc.Service) capabilityhostconfig.RawConfigReader {
	t.Helper()

	reader, ok := configSvc.(capabilityhostconfig.RawConfigReader)
	if !ok {
		t.Fatal("test config service does not support raw host config reads")
	}
	return reader
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
