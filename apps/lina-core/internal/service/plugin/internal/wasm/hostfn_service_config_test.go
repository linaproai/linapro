// This file tests the dynamic-plugin read-only config host service.

package wasm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gerror"

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
func (f *trackingConfigFactory) WithArtifactConfig(pluginID string, content []byte) plugincap.ConfigServiceFactory {
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
func (s *trackingConfigService) Get(_ context.Context, key string) (*gvar.Var, error) {
	if strings.TrimSpace(key) == "" || strings.TrimSpace(key) == "." {
		return nil, gerror.New("plugin config key cannot be empty or root")
	}
	s.getCalls++
	s.lastKey = key
	if value, ok := s.values[key]; ok {
		return gvar.New(value), nil
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
	value, err := s.Get(ctx, key)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.String(), nil
}

// Bool reads a deterministic bool value.
func (s *trackingConfigService) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Bool(), nil
}

// Int reads a deterministic int value.
func (s *trackingConfigService) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key)
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
// plugin-scoped config values through config.get only.
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
		protocol.HostServiceMethodConfigGet,
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
		protocol.HostServiceMethodConfigGet,
		"feature.enabled",
	)
	boolPayload := decodeConfigResponse(t, boolResponse)
	if !boolPayload.Found || boolPayload.Value != `true` {
		t.Fatalf("expected feature.enabled JSON bool value, got %#v", boolPayload)
	}

	intResponse := invokeConfigHostService(
		t,
		configHostCallContext(),
		protocol.HostServiceMethodConfigGet,
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
		protocol.HostServiceMethodConfigGet,
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
		protocol.HostServiceMethodConfigGet,
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
		protocol.HostServiceMethodConfigGet,
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

// TestHandleHostServiceInvokeConfigRejectsTypedMethod verifies dynamic config
// calls reject SDK helper names at authorization time.
func TestHandleHostServiceInvokeConfigRejectsTypedMethod(t *testing.T) {
	configureTrackingConfigFactory(t, &trackingConfigService{values: map[string]any{
		"feature.name": "demo",
	}})

	response := invokeConfigHostService(
		t,
		configHostCallContext(),
		protocol.HostServiceMethodConfigString,
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
// config host service declarations and calls remain limited to get.
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

// TestConfigureConfigHostServiceRejectsNil verifies missing runtime config
// factory injection returns an error instead of silently constructing an isolated adapter.
func TestConfigureConfigHostServiceRejectsNil(t *testing.T) {
	if err := ConfigureConfigHostService(nil); err == nil {
		t.Fatal("expected nil config host service factory to return an error")
	}
}

// configHostCallContext builds an authorized config host service context.
func configHostCallContext() *hostCallContext {
	return &hostCallContext{
		pluginID: "test-plugin-config",
		capabilities: map[string]struct{}{
			protocol.CapabilityConfig: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceConfig,
			},
		},
	}
}

// invokeConfigHostService dispatches one config host-service request.
func invokeConfigHostService(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	key string,
) *protocol.HostCallResponseEnvelope {
	t.Helper()

	request := &protocol.HostServiceRequestEnvelope{
		Service: protocol.HostServiceConfig,
		Method:  method,
		Payload: protocol.MarshalHostServiceConfigKeyRequest(&protocol.HostServiceConfigKeyRequest{
			Key: key,
		}),
	}
	return handleHostServiceInvoke(context.Background(), hcc, protocol.MarshalHostServiceRequestEnvelope(request))
}

// decodeConfigResponse verifies success and decodes one config host service response.
func decodeConfigResponse(
	t *testing.T,
	response *protocol.HostCallResponseEnvelope,
) *protocol.HostServiceConfigValueResponse {
	t.Helper()

	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected config host service success, got status=%d payload=%s", response.Status, string(response.Payload))
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
	previousFactory := configHostServiceFactory
	if err := ConfigureConfigHostService(factory); err != nil {
		t.Fatalf("configure config host service failed: %v", err)
	}
	t.Cleanup(func() {
		configHostServiceFactory = previousFactory
	})
	return factory
}
