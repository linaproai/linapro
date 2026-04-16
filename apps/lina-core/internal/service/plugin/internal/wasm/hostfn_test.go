// This file tests the shared host call entrypoint dispatch and error
// propagation behavior.
package wasm

import (
	"testing"

	"lina-core/pkg/pluginbridge"
)

func TestValidateCapabilitiesAcceptsValid(t *testing.T) {
	err := pluginbridge.ValidateCapabilities([]string{
		pluginbridge.CapabilityRuntime,
		pluginbridge.CapabilityDataRead,
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateCapabilitiesRejectsUnknown(t *testing.T) {
	err := pluginbridge.ValidateCapabilities([]string{pluginbridge.CapabilityRuntime, "host:unknown"})
	if err == nil {
		t.Error("expected error for unknown capability")
	}
}

func TestValidateCapabilitiesRejectsEmpty(t *testing.T) {
	err := pluginbridge.ValidateCapabilities([]string{""})
	if err == nil {
		t.Error("expected error for empty capability")
	}
}

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
