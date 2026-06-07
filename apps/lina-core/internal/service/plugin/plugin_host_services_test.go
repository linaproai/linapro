// This file verifies root-level host service and Wasm host service wiring.

package plugin

import (
	"strings"
	"testing"

	"lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/hostlock"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/locker"
	notifysvc "lina-core/internal/service/notify"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	capabilityhostconfig "lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilitymanifest "lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/plugincap"
	capabilityconfig "lina-core/pkg/plugin/capability/plugincap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
)

// wasmHostServiceTestDeps groups explicit dependencies for the root Wasm host
// service configuration entry.
type wasmHostServiceTestDeps struct {
	// kvCacheSvc is the governed cache service shared with dynamic plugins.
	kvCacheSvc kvcache.Service
	// lockSvc is the governed host lock service shared with dynamic plugins.
	lockSvc hostlock.Service
	// notifySvc is the notification service shared with dynamic plugins.
	notifySvc notifysvc.Service
	// configSvc reads dynamic-plugin storage configuration.
	configSvc configsvc.PluginConfigReader
	// hostServices provides plugin capability directories for host service dispatch.
	hostServices capability.Services
	// configFactory creates plugin-scoped config service views.
	configFactory plugincap.ConfigServiceFactory
	// hostConfigSvc exposes authorized host configuration reads.
	hostConfigSvc hostconfigcap.Service
	// manifestFactory creates plugin-scoped manifest service views.
	manifestFactory manifestcap.ServiceFactory
}

// TestConfigureWasmHostServicesRequiresExplicitDependencies verifies the root
// startup entry configures every host service dependency and rejects missing
// required services before dispatchers can use package defaults.
func TestConfigureWasmHostServicesRequiresExplicitDependencies(t *testing.T) {
	t.Cleanup(func() {
		if err := configureWasmHostServicesForTest(newWasmHostServiceTestDeps(t)); err != nil {
			t.Fatalf("recover wasm host service configuration failed: %v", err)
		}
	})

	validDeps := newWasmHostServiceTestDeps(t)
	if err := configureWasmHostServicesForTest(validDeps); err != nil {
		t.Fatalf("expected complete wasm host service configuration to succeed, got error: %v", err)
	}

	cases := []struct {
		name    string
		mutate  func(*wasmHostServiceTestDeps)
		message string
	}{
		{
			name:    "cache",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.kvCacheSvc = nil },
			message: "configure wasm cache host service failed",
		},
		{
			name:    "lock",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.lockSvc = nil },
			message: "configure wasm lock host service failed",
		},
		{
			name:    "notify",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.notifySvc = nil },
			message: "configure wasm notify host service failed",
		},
		{
			name:    "storage",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.configSvc = nil },
			message: "configure wasm storage host service failed",
		},
		{
			name:    "capabilities",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.hostServices = nil },
			message: "configure wasm ai text host service failed",
		},
		{
			name:    "config",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.configFactory = nil },
			message: "configure wasm config host service failed",
		},
		{
			name:    "host-config",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.hostConfigSvc = nil },
			message: "configure wasm host config service failed",
		},
		{
			name:    "manifest",
			mutate:  func(deps *wasmHostServiceTestDeps) { deps.manifestFactory = nil },
			message: "configure wasm manifest host service failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := newWasmHostServiceTestDeps(t)
			tc.mutate(deps)
			err := configureWasmHostServicesForTest(deps)
			if err == nil {
				t.Fatalf("expected missing %s dependency to fail", tc.name)
			}
			if !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("expected error to contain %q, got %v", tc.message, err)
			}
		})
	}

	if err := configureWasmHostServicesForTest(newWasmHostServiceTestDeps(t)); err != nil {
		t.Fatalf("expected complete wasm host service configuration to recover after nil cases, got error: %v", err)
	}
}

// newWasmHostServiceTestDeps builds a complete dependency set for root Wasm
// host service configuration tests.
func newWasmHostServiceTestDeps(t *testing.T) *wasmHostServiceTestDeps {
	t.Helper()

	configSvc := configsvc.New()
	bizCtxSvc := bizctx.New()
	lockSvc, err := hostlock.New(locker.New())
	if err != nil {
		t.Fatalf("create host lock service failed: %v", err)
	}

	return &wasmHostServiceTestDeps{
		kvCacheSvc:      kvcache.New(),
		lockSvc:         lockSvc,
		notifySvc:       notifysvc.New(tenantcapsvc.New(nil, bizCtxSvc)),
		configSvc:       configSvc,
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

// configureWasmHostServicesForTest calls the production root configuration
// entry with one explicit dependency set.
func configureWasmHostServicesForTest(deps *wasmHostServiceTestDeps) error {
	return ConfigureWasmHostServices(
		deps.kvCacheSvc,
		deps.lockSvc,
		deps.notifySvc,
		deps.configSvc,
		deps.hostServices,
		deps.configFactory,
		deps.hostConfigSvc,
		deps.manifestFactory,
	)
}
