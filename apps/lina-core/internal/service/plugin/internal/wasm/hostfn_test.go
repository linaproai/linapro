// This file tests the shared host call entrypoint dispatch and error
// propagation behavior.

package wasm

import (
	"testing"

	"lina-core/pkg/pluginbridge"
)

// TestValidateCapabilitiesAcceptsValid verifies known capabilities pass schema
// validation.
func TestValidateCapabilitiesAcceptsValid(t *testing.T) {
	err := pluginbridge.ValidateCapabilities([]string{
		pluginbridge.CapabilityRuntime,
		pluginbridge.CapabilityDataRead,
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestValidateCapabilitiesRejectsUnknown verifies unknown capability names are
// rejected by validation.
func TestValidateCapabilitiesRejectsUnknown(t *testing.T) {
	err := pluginbridge.ValidateCapabilities([]string{pluginbridge.CapabilityRuntime, "host:unknown"})
	if err == nil {
		t.Error("expected error for unknown capability")
	}
}

// TestValidateCapabilitiesRejectsEmpty verifies empty capability entries are
// rejected during validation.
func TestValidateCapabilitiesRejectsEmpty(t *testing.T) {
	err := pluginbridge.ValidateCapabilities([]string{""})
	if err == nil {
		t.Error("expected error for empty capability")
	}
}

// TestCapabilitiesFromHostServicesDerivesRuntimeCapability verifies runtime
// host services imply the runtime capability grant.
func TestCapabilitiesFromHostServicesDerivesRuntimeCapability(t *testing.T) {
	capabilities := pluginbridge.CapabilitiesFromHostServices(
		[]*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceRuntime,
				Methods: []string{
					pluginbridge.HostServiceMethodRuntimeLogWrite,
					pluginbridge.HostServiceMethodRuntimeInfoUUID,
				},
			},
		},
	)
	if len(capabilities) != 1 || capabilities[0] != pluginbridge.CapabilityRuntime {
		t.Fatalf("expected derived runtime capability, got %#v", capabilities)
	}
}

// TestHostCallContextHasCapability verifies direct capability lookups against
// the precomputed capability set.
func TestHostCallContextHasCapability(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityRuntime: {},
		},
	}
	if !hcc.hasCapability(pluginbridge.CapabilityRuntime) {
		t.Error("expected host:runtime to be granted")
	}
	if hcc.hasCapability(pluginbridge.CapabilityStorage) {
		t.Error("expected host:storage to not be granted")
	}
}

// TestHostCallContextHasHostServiceAccess verifies host-service method
// authorization honors the declared method allowlist.
func TestHostCallContextHasHostServiceAccess(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		hostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceRuntime,
				Methods: []string{
					pluginbridge.HostServiceMethodRuntimeLogWrite,
					pluginbridge.HostServiceMethodRuntimeInfoUUID,
				},
			},
		},
	}
	if !hcc.hasHostServiceAccess(pluginbridge.HostServiceRuntime, pluginbridge.HostServiceMethodRuntimeLogWrite, "", "") {
		t.Error("expected runtime log.write to be authorized")
	}
	if hcc.hasHostServiceAccess(pluginbridge.HostServiceRuntime, pluginbridge.HostServiceMethodRuntimeStateGet, "", "") {
		t.Error("expected runtime state.get to be denied")
	}
}

// TestHostCallContextHasDataTableAccess verifies data-table authorization is
// limited to explicitly granted tables.
func TestHostCallContextHasDataTableAccess(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		hostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceData,
				Methods: []string{pluginbridge.HostServiceMethodDataList},
				Tables:  []string{"sys_plugin_node_state"},
			},
		},
	}
	if !hcc.hasHostServiceAccess(pluginbridge.HostServiceData, pluginbridge.HostServiceMethodDataList, "", "sys_plugin_node_state") {
		t.Error("expected data list on authorized table to be allowed")
	}
	if hcc.hasHostServiceAccess(pluginbridge.HostServiceData, pluginbridge.HostServiceMethodDataList, "", "sys_user") {
		t.Error("expected data list on unauthorized table to be denied")
	}
}

// TestHandleHostServiceInvokeRejectsUnsupportedMethod verifies unknown handler
// methods return a not-found response.
func TestHandleHostServiceInvokeRejectsUnsupportedMethod(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityRuntime: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceRuntime,
				Methods: []string{pluginbridge.HostServiceMethodRuntimeInfoUUID},
			},
		},
	}
	request := &pluginbridge.HostServiceRequestEnvelope{
		Service: pluginbridge.HostServiceRuntime,
		Method:  "info.unknown",
	}
	response := handleHostServiceInvoke(nil, hcc, pluginbridge.MarshalHostServiceRequestEnvelope(request))
	if response.Status != pluginbridge.HostCallStatusNotFound {
		t.Errorf("expected not_found, got status %d", response.Status)
	}
}

// TestHandleHostServiceInvokeRejectsUnauthorizedMethod verifies declared
// capabilities alone do not bypass host-service method authorization.
func TestHandleHostServiceInvokeRejectsUnauthorizedMethod(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityRuntime: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceRuntime,
				Methods: []string{pluginbridge.HostServiceMethodRuntimeInfoUUID},
			},
		},
	}
	request := &pluginbridge.HostServiceRequestEnvelope{
		Service: pluginbridge.HostServiceRuntime,
		Method:  pluginbridge.HostServiceMethodRuntimeInfoNode,
	}
	response := handleHostServiceInvoke(nil, hcc, pluginbridge.MarshalHostServiceRequestEnvelope(request))
	if response.Status != pluginbridge.HostCallStatusCapabilityDenied {
		t.Errorf("expected capability_denied, got status %d", response.Status)
	}
}

// TestHandleHostServiceInvokeRejectsUnauthorizedResourceRef verifies resource
// scoping is enforced before dispatching storage host-service calls.
func TestHandleHostServiceInvokeRejectsUnauthorizedResourceRef(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityStorage: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceStorage,
				Methods: []string{pluginbridge.HostServiceMethodStorageGet},
				Paths:   []string{"authorized-files/"},
			},
		},
	}
	request := &pluginbridge.HostServiceRequestEnvelope{
		Service:     pluginbridge.HostServiceStorage,
		Method:      pluginbridge.HostServiceMethodStorageGet,
		ResourceRef: "denied-files/demo.txt",
		Payload: pluginbridge.MarshalHostServiceStorageGetRequest(&pluginbridge.HostServiceStorageGetRequest{
			Path: "denied-files/demo.txt",
		}),
	}
	response := handleHostServiceInvoke(nil, hcc, pluginbridge.MarshalHostServiceRequestEnvelope(request))
	if response.Status != pluginbridge.HostCallStatusCapabilityDenied {
		t.Errorf("expected capability_denied, got status %d", response.Status)
	}
}

// TestHandleHostServiceInvokeReturnsRuntimeUUID verifies the runtime UUID
// helper returns a non-empty value when authorized.
func TestHandleHostServiceInvokeReturnsRuntimeUUID(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityRuntime: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceRuntime,
				Methods: []string{pluginbridge.HostServiceMethodRuntimeInfoUUID},
			},
		},
	}
	request := &pluginbridge.HostServiceRequestEnvelope{
		Service: pluginbridge.HostServiceRuntime,
		Method:  pluginbridge.HostServiceMethodRuntimeInfoUUID,
	}
	response := handleHostServiceInvoke(nil, hcc, pluginbridge.MarshalHostServiceRequestEnvelope(request))
	if response.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("expected success, got status %d payload=%s", response.Status, string(response.Payload))
	}
	value, err := pluginbridge.UnmarshalHostServiceValueResponse(response.Payload)
	if err != nil {
		t.Fatalf("expected runtime info payload to decode, got error: %v", err)
	}
	if value.Value == "" {
		t.Fatal("expected runtime uuid value to be non-empty")
	}
}
