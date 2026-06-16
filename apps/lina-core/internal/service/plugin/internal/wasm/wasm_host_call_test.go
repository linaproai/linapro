// This file tests the shared host call entrypoint dispatch and error
// propagation behavior.

package wasm

import (
	"sync"
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

// TestHostCallContextRequiresExplicitReadServiceMethods verifies read services
// are not authorized from service and resource declarations alone.
func TestHostCallContextRequiresExplicitReadServiceMethods(t *testing.T) {
	cases := []struct {
		name        string
		service     string
		method      string
		resourceRef string
		spec        *protocol.HostServiceSpec
	}{
		{
			name:        "host config",
			service:     protocol.HostServiceHostConfig,
			method:      protocol.HostServiceMethodHostConfigGet,
			resourceRef: "workspace.basePath",
			spec: &protocol.HostServiceSpec{
				Service: protocol.HostServiceHostConfig,
				Keys:    []string{"workspace.basePath"},
			},
		},
		{
			name:        "manifest",
			service:     protocol.HostServiceManifest,
			method:      protocol.HostServiceMethodManifestGet,
			resourceRef: "metadata.yaml",
			spec: &protocol.HostServiceSpec{
				Service: protocol.HostServiceManifest,
				Paths:   []string{"metadata.yaml"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hcc := &hostCallContext{
				pluginID:     "test-plugin",
				hostServices: []*protocol.HostServiceSpec{tc.spec},
			}
			if hcc.hasHostServiceAccess(tc.service, tc.method, tc.resourceRef, "") {
				t.Fatal("expected host service access to require an explicit method grant")
			}
		})
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

// TestHostCallContextReusesRequestAuthorizationSnapshot verifies one request
// builds the host-service authorization index once and keeps it detached from
// later mutations to the caller-owned declaration slice.
func TestHostCallContextReusesRequestAuthorizationSnapshot(t *testing.T) {
	specs := []*protocol.HostServiceSpec{
		{
			Service: protocol.HostServiceRuntime,
			Methods: []string{
				protocol.HostServiceMethodRuntimeLogWrite,
				protocol.HostServiceMethodRuntimeInfoUUID,
			},
		},
	}
	hcc := &hostCallContext{
		pluginID:     "test-plugin",
		hostServices: specs,
	}
	firstSnapshot := hcc.accessSnapshot()
	secondSnapshot := hcc.accessSnapshot()
	if firstSnapshot == nil || firstSnapshot != secondSnapshot {
		t.Fatal("expected host call context to reuse one request-local authorization snapshot")
	}
	specs[0].Methods = []string{protocol.HostServiceMethodRuntimeInfoUUID}
	if !hcc.hasHostServiceAccess(protocol.HostServiceRuntime, protocol.HostServiceMethodRuntimeLogWrite, "", "") {
		t.Fatal("expected current request to keep using its original authorization snapshot")
	}
}

// TestHostCallContextAccessSnapshotConcurrent verifies lazy authorization
// snapshot construction remains safe when one request dispatches concurrent
// host-service calls through a shared context.
func TestHostCallContextAccessSnapshotConcurrent(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{
					protocol.HostServiceMethodRuntimeInfoUUID,
				},
			},
		},
	}

	const (
		workers    = 8
		iterations = 50
	)
	snapshots := make(chan *hostServiceAccessSnapshot, workers*iterations)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				snapshots <- hcc.accessSnapshot()
			}
		}()
	}
	wg.Wait()
	close(snapshots)

	var first *hostServiceAccessSnapshot
	for snapshot := range snapshots {
		if snapshot == nil {
			t.Fatal("expected host-service authorization snapshot")
		}
		if first == nil {
			first = snapshot
			continue
		}
		if snapshot != first {
			t.Fatal("expected concurrent access to reuse one request-local authorization snapshot")
		}
	}
}

// TestHostCallContextAuthorizationShrinkUsesNewSnapshot verifies the next
// request observes a shrunken active-release host-service snapshot.
func TestHostCallContextAuthorizationShrinkUsesNewSnapshot(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{
					protocol.HostServiceMethodRuntimeInfoUUID,
				},
			},
		},
	}
	if hcc.hasHostServiceAccess(protocol.HostServiceRuntime, protocol.HostServiceMethodRuntimeLogWrite, "", "") {
		t.Fatal("expected shrunken request snapshot to reject removed runtime log.write method")
	}
	if !hcc.hasHostServiceAccess(protocol.HostServiceRuntime, protocol.HostServiceMethodRuntimeInfoUUID, "", "") {
		t.Fatal("expected remaining runtime info.uuid method to stay authorized")
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
	response := handleHostServiceInvoke(nil, withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(request))
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
	response := handleHostServiceInvoke(nil, withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(request))
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
	response := handleHostServiceInvoke(nil, withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(request))
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
	response := handleHostServiceInvoke(nil, withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(request))
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
