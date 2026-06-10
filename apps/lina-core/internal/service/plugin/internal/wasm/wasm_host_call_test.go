// This file tests the shared host call entrypoint dispatch and error
// propagation behavior.

package wasm

import (
	"testing"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// assertPanic verifies the supplied function panics with the expected message.
func assertPanic(t *testing.T, expected string, fn func()) {
	t.Helper()

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("expected panic %q", expected)
		}
		if recovered != expected {
			t.Fatalf("expected panic %q, got %#v", expected, recovered)
		}
	}()

	fn()
}

// TestValidateCapabilitiesAcceptsValid verifies known capabilities pass schema
// validation.
func TestValidateCapabilitiesAcceptsValid(t *testing.T) {
	err := protocol.ValidateCapabilities([]string{
		protocol.CapabilityRuntime,
		protocol.CapabilityDataRead,
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestValidateCapabilitiesRejectsUnknown verifies unknown capability names are
// rejected by validation.
func TestValidateCapabilitiesRejectsUnknown(t *testing.T) {
	err := protocol.ValidateCapabilities([]string{protocol.CapabilityRuntime, "host:unknown"})
	if err == nil {
		t.Error("expected error for unknown capability")
	}
}

// TestValidateCapabilitiesRejectsEmpty verifies empty capability entries are
// rejected during validation.
func TestValidateCapabilitiesRejectsEmpty(t *testing.T) {
	err := protocol.ValidateCapabilities([]string{""})
	if err == nil {
		t.Error("expected error for empty capability")
	}
}

// TestCapabilitiesFromHostServicesDerivesRuntimeCapability verifies runtime
// host services imply the runtime capability grant.
func TestCapabilitiesFromHostServicesDerivesRuntimeCapability(t *testing.T) {
	capabilities := protocol.CapabilitiesFromHostServices(
		[]*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{
					protocol.HostServiceMethodRuntimeLogWrite,
					protocol.HostServiceMethodRuntimeInfoUUID,
				},
			},
		},
	)
	if len(capabilities) != 1 || capabilities[0] != protocol.CapabilityRuntime {
		t.Fatalf("expected derived runtime capability, got %#v", capabilities)
	}
}

// TestHostCallContextHasCapability verifies direct capability lookups against
// the precomputed capability set.
func TestHostCallContextHasCapability(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilityRuntime: {},
		},
	}
	if !hcc.hasCapability(protocol.CapabilityRuntime) {
		t.Error("expected host:runtime to be granted")
	}
	if hcc.hasCapability(protocol.CapabilityStorage) {
		t.Error("expected host:storage to not be granted")
	}
}

// TestHostCallContextHasHostServiceAccess verifies host-service method
// authorization honors the declared method allowlist.
func TestHostCallContextHasHostServiceAccess(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{
					protocol.HostServiceMethodRuntimeLogWrite,
					protocol.HostServiceMethodRuntimeInfoUUID,
				},
			},
		},
	}
	if !hcc.hasHostServiceAccess(protocol.HostServiceRuntime, protocol.HostServiceMethodRuntimeLogWrite, "", "") {
		t.Error("expected runtime log.write to be authorized")
	}
	if hcc.hasHostServiceAccess(protocol.HostServiceRuntime, protocol.HostServiceMethodRuntimeStateGet, "", "") {
		t.Error("expected runtime state.get to be denied")
	}
}

// TestHostCallContextAuthorizesPluginsConfigMethod verifies plugin config
// access is governed as an explicit plugins domain method.
func TestHostCallContextAuthorizesPluginsConfigMethod(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServicePlugins,
				Methods: []string{
					protocol.HostServiceMethodPluginsConfigGet,
				},
			},
		},
	}
	if !hcc.hasHostServiceAccess(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsConfigGet, "", "") {
		t.Error("expected plugins config.get to be authorized when explicitly declared")
	}
	if hcc.hasHostServiceAccess(protocol.HostServicePlugins, "config.exists", "", "") {
		t.Error("expected config exists helper method to be unauthorized")
	}
	if hcc.hasHostServiceAccess(protocol.HostServicePlugins, "config.set", "", "") {
		t.Error("expected unsupported config method to remain unauthorized")
	}
}

// TestHostCallContextHasHostConfigKeyAccess verifies hostConfig authorization
// uses the resourceRef key from the request envelope.
func TestHostCallContextHasHostConfigKeyAccess(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		hostServices: []*protocol.HostServiceSpec{{
			Service: protocol.HostServiceHostConfig,
			Methods: []string{protocol.HostServiceMethodHostConfigGet},
			Keys:    []string{"workspace.basePath"},
		}},
	}
	if !hcc.hasHostServiceAccess(protocol.HostServiceHostConfig, protocol.HostServiceMethodHostConfigGet, "workspace.basePath", "") {
		t.Error("expected authorized hostConfig key to be allowed")
	}
	if hcc.hasHostServiceAccess(protocol.HostServiceHostConfig, protocol.HostServiceMethodHostConfigGet, "database.default.link", "") {
		t.Error("expected unauthorized hostConfig key to be denied")
	}
}

// TestHostCallContextHasManifestPathAccess verifies manifest authorization
// accepts exact and globbed manifest-relative paths.
func TestHostCallContextHasManifestPathAccess(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		hostServices: []*protocol.HostServiceSpec{{
			Service: protocol.HostServiceManifest,
			Methods: []string{protocol.HostServiceMethodManifestGet},
			Paths:   []string{"metadata.yaml", "resources/*.yaml", "config/config.example.yaml"},
		}},
	}
	if !hcc.hasHostServiceAccess(protocol.HostServiceManifest, protocol.HostServiceMethodManifestGet, "metadata.yaml", "") {
		t.Error("expected exact manifest path to be allowed")
	}
	if !hcc.hasHostServiceAccess(protocol.HostServiceManifest, protocol.HostServiceMethodManifestGet, "resources/policy.yaml", "") {
		t.Error("expected globbed manifest path to be allowed")
	}
	if !hcc.hasHostServiceAccess(protocol.HostServiceManifest, protocol.HostServiceMethodManifestGet, "config/config.example.yaml", "") {
		t.Error("expected authorized config manifest path to be allowed")
	}
	if hcc.hasHostServiceAccess(protocol.HostServiceManifest, protocol.HostServiceMethodManifestGet, "config/config.yaml", "") {
		t.Error("expected unauthorized config manifest path to be denied")
	}
}

// TestHostCallContextHasDataTableAccess verifies data-table authorization is
// limited to explicitly granted tables.
func TestHostCallContextHasDataTableAccess(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceData,
				Methods: []string{protocol.HostServiceMethodDataList},
				Tables:  []string{"sys_plugin_node_state"},
			},
		},
	}
	if !hcc.hasHostServiceAccess(protocol.HostServiceData, protocol.HostServiceMethodDataList, "", "sys_plugin_node_state") {
		t.Error("expected data list on authorized table to be allowed")
	}
	if hcc.hasHostServiceAccess(protocol.HostServiceData, protocol.HostServiceMethodDataList, "", "sys_user") {
		t.Error("expected data list on unauthorized table to be denied")
	}
}

// TestHandleHostServiceInvokeRejectsUnsupportedMethod verifies unknown handler
// methods return a not-found response.
func TestHandleHostServiceInvokeRejectsUnsupportedMethod(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilityRuntime: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{protocol.HostServiceMethodRuntimeInfoUUID},
			},
		},
	}
	request := &protocol.HostServiceRequestEnvelope{
		Service: protocol.HostServiceRuntime,
		Method:  "info.unknown",
	}
	response := handleHostServiceInvoke(nil, hcc, protocol.MarshalHostServiceRequestEnvelope(request))
	if response.Status != protocol.HostCallStatusNotFound {
		t.Errorf("expected not_found, got status %d", response.Status)
	}
}

// TestHandleHostServiceInvokeRejectsUnauthorizedMethod verifies declared
// capabilities alone do not bypass host-service method authorization.
func TestHandleHostServiceInvokeRejectsUnauthorizedMethod(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilityRuntime: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{protocol.HostServiceMethodRuntimeInfoUUID},
			},
		},
	}
	request := &protocol.HostServiceRequestEnvelope{
		Service: protocol.HostServiceRuntime,
		Method:  protocol.HostServiceMethodRuntimeInfoNode,
	}
	response := handleHostServiceInvoke(nil, hcc, protocol.MarshalHostServiceRequestEnvelope(request))
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Errorf("expected capability_denied, got status %d", response.Status)
	}
}

// TestHandleHostServiceInvokeRejectsUnauthorizedResourceRef verifies resource
// scoping is enforced before dispatching storage host-service calls.
func TestHandleHostServiceInvokeRejectsUnauthorizedResourceRef(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilityStorage: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceStorage,
				Methods: []string{protocol.HostServiceMethodStorageGet},
				Paths:   []string{"authorized-files/"},
			},
		},
	}
	request := &protocol.HostServiceRequestEnvelope{
		Service:     protocol.HostServiceStorage,
		Method:      protocol.HostServiceMethodStorageGet,
		ResourceRef: "denied-files/demo.txt",
		Payload: protocol.MarshalHostServiceStorageGetRequest(&protocol.HostServiceStorageGetRequest{
			Path: "denied-files/demo.txt",
		}),
	}
	response := handleHostServiceInvoke(nil, hcc, protocol.MarshalHostServiceRequestEnvelope(request))
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Errorf("expected capability_denied, got status %d", response.Status)
	}
}

// TestHandleHostServiceInvokeReturnsRuntimeUUID verifies the runtime UUID
// helper returns a non-empty value when authorized.
func TestHandleHostServiceInvokeReturnsRuntimeUUID(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		capabilities: map[string]struct{}{
			protocol.CapabilityRuntime: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{protocol.HostServiceMethodRuntimeInfoUUID},
			},
		},
	}
	request := &protocol.HostServiceRequestEnvelope{
		Service: protocol.HostServiceRuntime,
		Method:  protocol.HostServiceMethodRuntimeInfoUUID,
	}
	response := handleHostServiceInvoke(nil, hcc, protocol.MarshalHostServiceRequestEnvelope(request))
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected success, got status %d payload=%s", response.Status, string(response.Payload))
	}
	value, err := protocol.UnmarshalHostServiceValueResponse(response.Payload)
	if err != nil {
		t.Fatalf("expected runtime info payload to decode, got error: %v", err)
	}
	if value.Value == "" {
		t.Fatal("expected runtime uuid value to be non-empty")
	}
}
