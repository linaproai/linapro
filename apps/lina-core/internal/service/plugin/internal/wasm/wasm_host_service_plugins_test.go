// This file tests plugin-governance host-service dispatch, including
// plugins.config.get for dynamic plugin config reads.

package wasm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/pluginconfig"
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// trackingConfigFactory records plugin scopes requested by the wasm dispatcher.
type trackingConfigFactory struct {
	service             *trackingConfigService
	lastPluginID        string
	lastArtifactPlugin  string
	lastArtifactContent string
}

// ForPlugin returns the configured tracking service for one plugin scope.
func (f *trackingConfigFactory) ForPlugin(pluginID string) plugincap.ConfigService {
	f.lastPluginID = pluginID
	return f.service
}

// WithArtifactConfig records release-bound default config passed by the execution context.
func (f *trackingConfigFactory) WithArtifactConfig(pluginID string, content []byte) pluginconfig.Factory {
	f.lastArtifactPlugin = pluginID
	f.lastArtifactContent = string(content)
	return f
}

// trackingConfigService records config reads while returning deterministic values.
type trackingConfigService struct {
	values      map[string]any
	getCalls    int
	existsCalls int
	lastKey     string
}

// Get records one raw config read.
func (s *trackingConfigService) Get(_ context.Context, key string, defaultValue any) (*gvar.Var, error) {
	if strings.TrimSpace(key) == "" || strings.TrimSpace(key) == "." {
		return nil, gerror.New("plugin config key cannot be empty or root")
	}
	s.getCalls++
	s.lastKey = key
	if value, ok := s.values[key]; ok {
		return gvar.New(value), nil
	}
	if defaultValue != nil {
		return gvar.New(defaultValue), nil
	}
	return nil, nil
}

// Exists records one config existence read.
func (s *trackingConfigService) Exists(_ context.Context, key string) (bool, error) {
	if strings.TrimSpace(key) == "" || strings.TrimSpace(key) == "." {
		return false, gerror.New("plugin config key cannot be empty or root")
	}
	s.existsCalls++
	s.lastKey = key
	_, ok := s.values[key]
	return ok, nil
}

// Scan records no behavior for the config fake.
func (s *trackingConfigService) Scan(context.Context, string, any) error { return nil }

// String reads a deterministic string value.
func (s *trackingConfigService) String(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.String(), nil
}

// Bool reads a deterministic bool value.
func (s *trackingConfigService) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Bool(), nil
}

// Int reads a deterministic int value.
func (s *trackingConfigService) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Int(), nil
}

// Duration returns a deterministic duration value.
func (s *trackingConfigService) Duration(context.Context, string, time.Duration) (time.Duration, error) {
	return 15 * time.Second, nil
}

// TestHandleHostServiceInvokeConfigReadsValues verifies dynamic plugins read
// plugin-scoped config values through plugins.config.get only.
func TestHandleHostServiceInvokeConfigReadsValues(t *testing.T) {
	configSvc := &trackingConfigService{values: map[string]any{
		"monitor.interval": "45s",
		"feature.enabled":  true,
		"feature.retries":  3,
	}}
	factory := configureTrackingConfigFactory(t, configSvc)

	getResponse := invokeConfigHostService(
		t,
		configHostCallContext(),
		protocol.HostServiceMethodPluginsConfigGet,
		"monitor.interval",
	)
	getPayload := decodeConfigResponse(t, getResponse)
	if !getPayload.Found || getPayload.Value != `"45s"` {
		t.Fatalf("expected monitor.interval JSON value, got %#v", getPayload)
	}
	if factory.lastPluginID != "test-plugin-config" {
		t.Fatalf("expected config factory to be scoped to plugin, got %q", factory.lastPluginID)
	}

	boolResponse := invokeConfigHostService(
		t,
		configHostCallContext(),
		protocol.HostServiceMethodPluginsConfigGet,
		"feature.enabled",
	)
	boolPayload := decodeConfigResponse(t, boolResponse)
	if !boolPayload.Found || boolPayload.Value != `true` {
		t.Fatalf("expected feature.enabled JSON bool value, got %#v", boolPayload)
	}

	intResponse := invokeConfigHostService(
		t,
		configHostCallContext(),
		protocol.HostServiceMethodPluginsConfigGet,
		"feature.retries",
	)
	intPayload := decodeConfigResponse(t, intResponse)
	if !intPayload.Found || intPayload.Value != `3` {
		t.Fatalf("expected feature.retries JSON int value, got %#v", intPayload)
	}
	if configSvc.existsCalls != 3 || configSvc.getCalls != 3 {
		t.Fatalf("expected exists/get per config read, got exists=%d get=%d", configSvc.existsCalls, configSvc.getCalls)
	}
}

// TestHandleHostServiceInvokeConfigRejectsRootRead verifies empty key cannot
// read a full config snapshot.
func TestHandleHostServiceInvokeConfigRejectsRootRead(t *testing.T) {
	configureTrackingConfigFactory(t, &trackingConfigService{values: map[string]any{
		"custom.name": "demo",
	}})

	response := invokeConfigHostService(
		t,
		configHostCallContext(),
		protocol.HostServiceMethodPluginsConfigGet,
		"",
	)
	if response.Status != protocol.HostCallStatusInternalError {
		t.Fatalf("expected root config read to be rejected, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeConfigMissingKey verifies missing keys return found=false.
func TestHandleHostServiceInvokeConfigMissingKey(t *testing.T) {
	configureTrackingConfigFactory(t, &trackingConfigService{values: map[string]any{
		"custom.name": "demo",
	}})

	response := invokeConfigHostService(
		t,
		configHostCallContext(),
		protocol.HostServiceMethodPluginsConfigGet,
		"custom.missing",
	)
	payload := decodeConfigResponse(t, response)
	if payload.Found {
		t.Fatalf("expected missing key to return found=false, got %#v", payload)
	}
}

// TestHandleHostServiceInvokeConfigBindsArtifactDefaultConfig verifies active
// release default config is passed to the scoped factory for each execution.
func TestHandleHostServiceInvokeConfigBindsArtifactDefaultConfig(t *testing.T) {
	configSvc := &trackingConfigService{values: map[string]any{
		"feature.name": "demo",
	}}
	factory := configureTrackingConfigFactory(t, configSvc)
	hcc := configHostCallContext()
	hcc.artifactDefaultConfig = []byte("feature:\n  name: artifact\n")

	response := invokeConfigHostService(
		t,
		hcc,
		protocol.HostServiceMethodPluginsConfigGet,
		"feature.name",
	)
	payload := decodeConfigResponse(t, response)
	if !payload.Found || payload.Value != `"demo"` {
		t.Fatalf("expected config response, got %#v", payload)
	}
	if factory.lastArtifactPlugin != "test-plugin-config" || factory.lastArtifactContent != "feature:\n  name: artifact\n" {
		t.Fatalf("expected artifact config binding, got plugin=%q content=%q", factory.lastArtifactPlugin, factory.lastArtifactContent)
	}
}

// TestHandleHostServiceInvokeConfigPrefersHostStaticConfig verifies dynamic
// plugins.config.get uses plugin.<plugin-id> from host static config before
// release-bound artifact defaults.
func TestHandleHostServiceInvokeConfigPrefersHostStaticConfig(t *testing.T) {
	factory := pluginconfig.NewFactoryWithHostStaticConfig(
		t.TempDir(),
		t.TempDir(),
		wasmTestHostStaticConfigReader{values: map[string]any{
			"plugin.test-plugin-config": map[string]any{
				"feature": map[string]any{
					"name": "host",
				},
			},
		}},
	)
	bindTestHostServiceRuntime(t, withTestConfigFactory(factory))
	hcc := configHostCallContext()
	hcc.artifactDefaultConfig = []byte("feature:\n  name: artifact\n")

	response := invokeConfigHostService(
		t,
		hcc,
		protocol.HostServiceMethodPluginsConfigGet,
		"feature.name",
	)
	payload := decodeConfigResponse(t, response)
	if !payload.Found || payload.Value != `"host"` {
		t.Fatalf("expected host static config response, got %#v", payload)
	}
}

// TestHandleHostServiceInvokeConfigUsesArtifactDefaultConfig verifies
// release-bound artifact default config remains scoped to the host-call context.
func TestHandleHostServiceInvokeConfigUsesArtifactDefaultConfig(t *testing.T) {
	factory := pluginconfig.NewFactoryWithHostStaticConfig(
		t.TempDir(),
		t.TempDir(),
		wasmTestHostStaticConfigReader{},
	)
	bindTestHostServiceRuntime(t, withTestConfigFactory(factory))
	hcc := configHostCallContext()
	hcc.artifactDefaultConfig = []byte("feature:\n  name: artifact\n")

	response := invokeConfigHostService(
		t,
		hcc,
		protocol.HostServiceMethodPluginsConfigGet,
		"feature.name",
	)
	payload := decodeConfigResponse(t, response)
	if !payload.Found || payload.Value != `"artifact"` {
		t.Fatalf("expected artifact default config response, got %#v", payload)
	}
}

// TestHandleHostServiceInvokePluginStateGovernanceMethods verifies dynamic
// plugin governance reads dispatch through Plugins().State().
func TestHandleHostServiceInvokePluginStateGovernanceMethods(t *testing.T) {
	stateSvc := &trackingPluginStateService{enabled: true}
	services := &capabilityHostServiceTestServices{
		plugins: &capabilityHostServicePluginsService{state: stateSvc},
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	response := invokeCapabilityHostService(
		t,
		pluginGovernanceHostCallContext(protocol.HostServiceMethodPluginsStateIsEnabled),
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsStateIsEnabled,
		marshalCapabilityJSONRequest(t, pluginIDRequest{PluginID: "linapro-tenant-core"}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected plugin capability success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	var enabled bool
	decodeCapabilityJSONResponse(t, response.Payload, &enabled)
	if !enabled || stateSvc.lastPluginID != "linapro-tenant-core" {
		t.Fatalf("unexpected plugin capability result enabled=%v plugin=%q", enabled, stateSvc.lastPluginID)
	}
}

// TestHandleHostServiceInvokePluginLifecycleGovernanceMethods verifies dynamic
// lifecycle governance dispatches through Plugins().Lifecycle().
func TestHandleHostServiceInvokePluginLifecycleGovernanceMethods(t *testing.T) {
	lifecycleSvc := &trackingPluginLifecycleService{}
	services := &capabilityHostServiceTestServices{
		plugins: &capabilityHostServicePluginsService{lifecycle: lifecycleSvc},
	}
	configureDomainHostServicesForCapabilityTest(t, services)

	ensureResponse := invokeCapabilityHostService(
		t,
		pluginGovernanceHostCallContext(protocol.HostServiceMethodPluginsLifecycleEnsureTenantPluginDisableAllowed),
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleEnsureTenantPluginDisableAllowed,
		marshalCapabilityJSONRequest(t, tenantPluginLifecycleRequest{PluginID: "linapro-demo-dynamic", TenantID: 9}),
	)
	if ensureResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected plugin lifecycle ensure success, got status=%d payload=%s", ensureResponse.Status, string(ensureResponse.Payload))
	}
	if lifecycleSvc.disablePluginID != "linapro-demo-dynamic" || lifecycleSvc.disableTenantID != 9 {
		t.Fatalf("unexpected disable lifecycle call plugin=%q tenant=%d", lifecycleSvc.disablePluginID, lifecycleSvc.disableTenantID)
	}

	notifyResponse := invokeCapabilityHostService(
		t,
		pluginGovernanceHostCallContext(protocol.HostServiceMethodPluginsLifecycleNotifyTenantDeleted),
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleNotifyTenantDeleted,
		marshalCapabilityJSONRequest(t, tenantIDRequest{TenantID: 9}),
	)
	if notifyResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected plugin lifecycle notify success, got status=%d payload=%s", notifyResponse.Status, string(notifyResponse.Payload))
	}
	if lifecycleSvc.deletedTenantID != 9 {
		t.Fatalf("unexpected tenant deleted lifecycle call tenant=%d", lifecycleSvc.deletedTenantID)
	}
}

// TestHandleHostServiceInvokeConfigRejectsTypedMethod verifies dynamic config
// calls reject SDK helper names at authorization time.
func TestHandleHostServiceInvokeConfigRejectsTypedMethod(t *testing.T) {
	configureTrackingConfigFactory(t, &trackingConfigService{values: map[string]any{
		"feature.name": "demo",
	}})

	response := invokeConfigHostService(
		t,
		configHostCallContext(),
		"string",
		"feature.name",
	)
	if response.Status != protocol.HostCallStatusNotFound {
		t.Fatalf(
			"expected typed config method to be rejected, got status=%d payload=%s",
			response.Status,
			string(response.Payload),
		)
	}
}

// TestHandleHostServiceInvokeConfigRejectsUnsupportedMethod verifies dynamic
// plugin config declarations and calls remain limited to get.
func TestHandleHostServiceInvokeConfigRejectsUnsupportedMethod(t *testing.T) {
	configureTrackingConfigFactory(t, &trackingConfigService{values: map[string]any{
		"custom.name": "demo",
	}})

	response := invokeConfigHostService(
		t,
		configHostCallContext(),
		"set",
		"custom.name",
	)
	if response.Status != protocol.HostCallStatusNotFound {
		t.Fatalf(
			"expected unsupported config method to be rejected, got status=%d payload=%s",
			response.Status,
			string(response.Payload),
		)
	}
}

// TestConfigurePluginConfigFactoryRejectsNil verifies missing runtime config
// factory injection returns an error instead of silently constructing an isolated adapter.
func TestConfigurePluginConfigFactoryRejectsNil(t *testing.T) {
	if _, err := NewRuntime(
		&capabilityHostServiceTestServices{},
		capregistry.NewRegistry(),
		nil,
		noopTestHostConfigService{},
		noopTestManifestFactory{},
	); err == nil {
		t.Fatal("expected nil plugin config service factory to return an error")
	}
}

// configHostCallContext builds an authorized plugins.config.get context.
func configHostCallContext() *hostCallContext {
	return &hostCallContext{
		pluginID: "test-plugin-config",
		capabilities: map[string]struct{}{
			protocol.CapabilityPlugins: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServicePlugins,
				Methods: []string{
					protocol.HostServiceMethodPluginsConfigGet,
				},
			},
		},
	}
}

// pluginGovernanceHostCallContext builds an authorized plugins host-service context.
func pluginGovernanceHostCallContext(methods ...string) *hostCallContext {
	return &hostCallContext{
		pluginID: "test-plugin-governance",
		capabilities: map[string]struct{}{
			protocol.CapabilityPlugins: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServicePlugins,
				Methods: methods,
			},
		},
		requestID: "trace-plugin-governance",
	}
}

// invokeConfigHostService dispatches one plugins.config.get request.
func invokeConfigHostService(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	key string,
) *protocol.HostCallResponseEnvelope {
	t.Helper()

	request := &protocol.HostServiceRequestEnvelope{
		Service: protocol.HostServicePlugins,
		Method:  method,
		Payload: protocol.MarshalHostServiceConfigKeyRequest(&protocol.HostServiceConfigKeyRequest{
			Key: key,
		}),
	}
	return handleHostServiceInvoke(context.Background(), withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(request))
}

// decodeConfigResponse verifies success and decodes one plugin config response.
func decodeConfigResponse(
	t *testing.T,
	response *protocol.HostCallResponseEnvelope,
) *protocol.HostServiceConfigValueResponse {
	t.Helper()

	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected plugin config success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	payload, err := protocol.UnmarshalHostServiceConfigValueResponse(response.Payload)
	if err != nil {
		t.Fatalf("expected config response decode to succeed, got error: %v", err)
	}
	return payload
}

// configureTrackingConfigFactory swaps the process config factory for one test case.
func configureTrackingConfigFactory(t *testing.T, service *trackingConfigService) *trackingConfigFactory {
	t.Helper()

	factory := &trackingConfigFactory{service: service}
	bindTestHostServiceRuntime(t, withTestConfigFactory(factory))
	return factory
}

// trackingPluginStateService records plugin enablement lookups.
type trackingPluginStateService struct {
	enabled       bool
	provider      bool
	authoritative bool
	lastPluginID  plugincap.PluginID
}

// IsEnabled records a regular plugin enablement lookup.
func (s *trackingPluginStateService) IsEnabled(_ context.Context, pluginID plugincap.PluginID) (bool, error) {
	s.lastPluginID = pluginID
	return s.enabled, nil
}

// IsProviderEnabled records a provider-level plugin enablement lookup.
func (s *trackingPluginStateService) IsProviderEnabled(_ context.Context, pluginID plugincap.PluginID) (bool, error) {
	s.lastPluginID = pluginID
	return s.provider, nil
}

// IsEnabledAuthoritative records an authoritative plugin enablement lookup.
func (s *trackingPluginStateService) IsEnabledAuthoritative(_ context.Context, pluginID plugincap.PluginID) (bool, error) {
	s.lastPluginID = pluginID
	return s.authoritative, nil
}

// trackingPluginLifecycleService records plugin lifecycle governance calls.
type trackingPluginLifecycleService struct {
	disablePluginID string
	disableTenantID int
	deletedTenantID int
}

// EnsureTenantPluginDisableAllowed records one tenant-plugin disable precondition call.
func (s *trackingPluginLifecycleService) EnsureTenantPluginDisableAllowed(
	_ context.Context,
	pluginID string,
	tenantID int,
) error {
	s.disablePluginID = pluginID
	s.disableTenantID = tenantID
	return nil
}

// NotifyTenantPluginDisabled records no behavior for this test fake.
func (*trackingPluginLifecycleService) NotifyTenantPluginDisabled(context.Context, string, int) {}

// EnsureTenantDeleteAllowed records no behavior for this test fake.
func (*trackingPluginLifecycleService) EnsureTenantDeleteAllowed(context.Context, int) error {
	return nil
}

// NotifyTenantDeleted records one tenant deletion notification.
func (s *trackingPluginLifecycleService) NotifyTenantDeleted(_ context.Context, tenantID int) {
	s.deletedTenantID = tenantID
}

// wasmTestHostStaticConfigReader returns deterministic host static sections
// for dynamic config host-service tests.
type wasmTestHostStaticConfigReader struct {
	values map[string]any
}

// GetRaw returns one configured host static test value.
func (r wasmTestHostStaticConfigReader) GetRaw(_ context.Context, key string) (*gvar.Var, error) {
	if r.values == nil {
		return nil, nil
	}
	value, ok := r.values[key]
	if !ok {
		return nil, nil
	}
	return gvar.New(value), nil
}
