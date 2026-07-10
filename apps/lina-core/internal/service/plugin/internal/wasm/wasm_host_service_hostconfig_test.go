// This file tests the dynamic-plugin hostConfig host service.

package wasm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	hostconfigadapter "lina-core/internal/service/plugin/internal/hostconfig"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// trackingHostConfigService records host config reads.
type trackingHostConfigService struct {
	values      map[string]any
	sysConfig   *trackingSysConfigService
	getCalls    int
	existsCalls int
	lastKey     string
}

// Get records one host config read.
func (s *trackingHostConfigService) Get(_ context.Context, key string, defaultValue any) (*gvar.Var, error) {
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

// Exists records one host config existence read.
func (s *trackingHostConfigService) Exists(_ context.Context, key string) (bool, error) {
	s.existsCalls++
	s.lastKey = key
	_, ok := s.values[key]
	return ok, nil
}

// String reads a deterministic string value.
func (s *trackingHostConfigService) String(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.String(), nil
}

// Bool reads a deterministic bool value.
func (s *trackingHostConfigService) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Bool(), nil
}

// Int reads a deterministic int value.
func (s *trackingHostConfigService) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Int(), nil
}

// Duration returns a deterministic duration value.
func (s *trackingHostConfigService) Duration(context.Context, string, time.Duration) (time.Duration, error) {
	return 15 * time.Second, nil
}

// SysConfig returns tracking sys_config methods for dynamic hostConfig dispatch tests.
func (s *trackingHostConfigService) SysConfig() hostconfigcap.SysConfigService {
	if s.sysConfig != nil {
		return s.sysConfig
	}
	return noopTestSysConfigService{}
}

type trackingSysConfigService struct {
	values       map[hostconfigcap.SysConfigKey]string
	lastGetKey   hostconfigcap.SysConfigKey
	lastSetKey   hostconfigcap.SysConfigKey
	lastSetValue string
	lastResetKey hostconfigcap.SysConfigKey
}

func (s *trackingSysConfigService) Get(_ context.Context, key hostconfigcap.SysConfigKey) (*hostconfigcap.SysConfigInfo, error) {
	s.lastGetKey = key
	if value, ok := s.values[key]; ok {
		return &hostconfigcap.SysConfigInfo{Key: key, Value: value}, nil
	}
	return nil, nil
}

func (s *trackingSysConfigService) BatchGet(ctx context.Context, keys []hostconfigcap.SysConfigKey) (*capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey], error) {
	result := &capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey]{
		Items:      map[hostconfigcap.SysConfigKey]*hostconfigcap.SysConfigInfo{},
		MissingIDs: []hostconfigcap.SysConfigKey{},
	}
	for _, key := range keys {
		item, _ := s.Get(ctx, key)
		if item == nil {
			result.MissingIDs = append(result.MissingIDs, key)
			continue
		}
		result.Items[key] = item
	}
	return result, nil
}

func (s *trackingSysConfigService) List(context.Context, hostconfigcap.ListSysConfigInput) (*capmodel.PageResult[*hostconfigcap.SysConfigInfo], error) {
	return &capmodel.PageResult[*hostconfigcap.SysConfigInfo]{Items: []*hostconfigcap.SysConfigInfo{}}, nil
}

func (s *trackingSysConfigService) SetValue(_ context.Context, key hostconfigcap.SysConfigKey, value string) error {
	s.lastSetKey = key
	s.lastSetValue = value
	return nil
}

func (s *trackingSysConfigService) Reset(_ context.Context, key hostconfigcap.SysConfigKey) error {
	s.lastResetKey = key
	return nil
}

func (s *trackingSysConfigService) EnsureVisible(context.Context, []hostconfigcap.SysConfigKey) error {
	return nil
}

// TestHandleHostServiceInvokeHostConfigReadsAuthorizedKey verifies dynamic
// plugins can read a hostConfig key only when it is authorized.
func TestHandleHostServiceInvokeHostConfigReadsAuthorizedKey(t *testing.T) {
	hostConfigSvc := &trackingHostConfigService{values: map[string]any{
		"workspace.basePath": "/admin",
	}}
	configureTrackingHostConfigService(t, hostConfigSvc)

	response := invokeHostConfigService(t, hostConfigHostCallContext([]string{"workspace.basePath"}), "workspace.basePath")
	payload := decodeConfigResponse(t, response)
	if !payload.Found || payload.Value != `"/admin"` {
		t.Fatalf("expected workspace.basePath JSON value, got %#v", payload)
	}
	if hostConfigSvc.existsCalls != 1 || hostConfigSvc.getCalls != 1 || hostConfigSvc.lastKey != "workspace.basePath" {
		t.Fatalf("expected hostConfig exists/get calls, got exists=%d get=%d key=%q", hostConfigSvc.existsCalls, hostConfigSvc.getCalls, hostConfigSvc.lastKey)
	}
}

// TestHandleHostServiceInvokeHostConfigRejectsUnauthorizedKey verifies
// resources.keys are enforced before dispatch.
func TestHandleHostServiceInvokeHostConfigRejectsUnauthorizedKey(t *testing.T) {
	configureTrackingHostConfigService(t, &trackingHostConfigService{values: map[string]any{
		"workspace.basePath": "/admin",
	}})

	response := invokeHostConfigService(t, hostConfigHostCallContext([]string{"workspace.basePath"}), "database.default.link")
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected unauthorized hostConfig key to be denied, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeHostConfigRejectsMismatchedPayloadKey verifies
// guests cannot authorize one hostConfig key and read a different payload key.
func TestHandleHostServiceInvokeHostConfigRejectsMismatchedPayloadKey(t *testing.T) {
	hostConfigSvc := &trackingHostConfigService{values: map[string]any{
		"workspace.basePath":    "/admin",
		"database.default.link": "hidden",
	}}
	configureTrackingHostConfigService(t, hostConfigSvc)

	response := invokeHostConfigServiceWithPayloadKey(
		t,
		hostConfigHostCallContext([]string{"workspace.basePath"}),
		"workspace.basePath",
		"database.default.link",
	)
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected mismatched hostConfig key to be denied, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if hostConfigSvc.existsCalls != 0 || hostConfigSvc.getCalls != 0 {
		t.Fatalf("expected mismatched hostConfig key to be rejected before service calls, got exists=%d get=%d", hostConfigSvc.existsCalls, hostConfigSvc.getCalls)
	}
}

// TestHandleHostServiceInvokeHostConfigReadsAuthorizedCustomSysConfig verifies
// dynamic plugins can read custom sys_config keys only after key authorization.
func TestHandleHostServiceInvokeHostConfigReadsAuthorizedCustomSysConfig(t *testing.T) {
	ctx := context.Background()
	key := fmt.Sprintf("custom.dynamic.limit.%d", time.Now().UnixNano())
	insertDynamicHostConfigSysConfig(t, ctx, key, "64")
	if err := hostconfig.New().MarkRuntimeParamsChanged(ctx); err != nil {
		t.Fatalf("mark runtime params changed: %v", err)
	}
	bindTestHostServiceRuntime(t, withTestHostConfigService(
		hostconfigadapter.NewStaticCapabilityAdapter(hostconfig.New()),
	))

	response := invokeHostConfigService(t, hostConfigHostCallContext([]string{key}), key)
	payload := decodeConfigResponse(t, response)
	if !payload.Found || payload.Value != `"64"` {
		t.Fatalf("expected custom sys_config JSON value, got %#v", payload)
	}

	denied := invokeHostConfigService(t, hostConfigHostCallContext([]string{"workspace.basePath"}), key)
	if denied.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected unauthorized custom sys_config key to be denied, got status=%d", denied.Status)
	}
}

// TestHandleHostServiceInvokeHostConfigAuthorizedKeyUsesUnifiedPriority
// verifies authorized dynamic hostConfig reads use the host GetRaw pipeline.
func TestHandleHostServiceInvokeHostConfigAuthorizedKeyUsesUnifiedPriority(t *testing.T) {
	ctx := context.Background()
	suffix := fmt.Sprintf("priority%d", time.Now().UnixNano())
	key := "custom.dynamic." + suffix
	setWasmHostConfigAdapter(t, fmt.Sprintf(`
custom:
  dynamic:
    %s: "/static-dynamic"
`, suffix))
	insertDynamicHostConfigSysConfig(t, ctx, key, "/runtime-dynamic")
	if err := hostconfig.New().MarkRuntimeParamsChanged(ctx); err != nil {
		t.Fatalf("mark runtime params changed: %v", err)
	}
	bindTestHostServiceRuntime(t, withTestHostConfigService(
		hostconfigadapter.NewStaticCapabilityAdapter(hostconfig.New()),
	))

	response := invokeHostConfigService(t, hostConfigHostCallContext([]string{key}), key)
	payload := decodeConfigResponse(t, response)
	if !payload.Found || payload.Value != `"/runtime-dynamic"` {
		t.Fatalf("expected runtime sys_config value to win, got %#v", payload)
	}
}

// TestHandleHostServiceInvokeHostConfigSysConfigSingleKeyMethods verifies
// sys_config dynamic methods stay bound to resources.keys through resourceRef.
func TestHandleHostServiceInvokeHostConfigSysConfigSingleKeyMethods(t *testing.T) {
	sysConfigSvc := &trackingSysConfigService{
		values: map[hostconfigcap.SysConfigKey]string{
			"custom.dynamic.limit": "64",
		},
	}
	domainServices := &capabilityHostServiceTestServices{
		hostConfig: &trackingHostConfigService{sysConfig: sysConfigSvc},
	}
	bindTestHostServiceRuntime(t, withTestDomainServices(domainServices))

	hcc := hostConfigHostCallContext([]string{"custom.dynamic.limit"})
	getResponse := invokeCapabilityHostService(
		t,
		hcc,
		protocol.HostServiceHostConfig,
		protocol.HostServiceMethodHostConfigSysConfigGet,
		marshalCapabilityJSONRequest(t, hostConfigSysConfigKeyRequest{Key: "custom.dynamic.limit"}),
	)
	if getResponse.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected missing resourceRef to be denied, got status=%d payload=%s", getResponse.Status, string(getResponse.Payload))
	}

	getResponse = invokeHostConfigSysConfigService(t, hcc, protocol.HostServiceMethodHostConfigSysConfigGet, "custom.dynamic.limit", hostConfigSysConfigKeyRequest{Key: "custom.dynamic.limit"})
	if getResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected sys_config get success, got status=%d payload=%s", getResponse.Status, string(getResponse.Payload))
	}
	var info *hostconfigcap.SysConfigInfo
	decodeCapabilityJSONResponse(t, getResponse.Payload, &info)
	if info == nil || info.Value != "64" || sysConfigSvc.lastGetKey != "custom.dynamic.limit" {
		t.Fatalf("unexpected sys_config get info=%#v lastKey=%q", info, sysConfigSvc.lastGetKey)
	}

	setResponse := invokeHostConfigSysConfigService(
		t,
		hcc,
		protocol.HostServiceMethodHostConfigSysConfigSetValue,
		"custom.dynamic.limit",
		hostConfigSysConfigSetValueRequest{Key: "custom.dynamic.limit", Value: "128"},
	)
	if setResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected sys_config set success, got status=%d payload=%s", setResponse.Status, string(setResponse.Payload))
	}
	if sysConfigSvc.lastSetKey != "custom.dynamic.limit" || sysConfigSvc.lastSetValue != "128" {
		t.Fatalf("unexpected sys_config set key=%q value=%q", sysConfigSvc.lastSetKey, sysConfigSvc.lastSetValue)
	}

	resetResponse := invokeHostConfigSysConfigService(
		t,
		hcc,
		protocol.HostServiceMethodHostConfigSysConfigReset,
		"custom.dynamic.limit",
		hostConfigSysConfigKeyRequest{Key: "custom.dynamic.limit"},
	)
	if resetResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected sys_config reset success, got status=%d payload=%s", resetResponse.Status, string(resetResponse.Payload))
	}
	if sysConfigSvc.lastResetKey != "custom.dynamic.limit" {
		t.Fatalf("unexpected sys_config reset key=%q", sysConfigSvc.lastResetKey)
	}

	denied := invokeHostConfigSysConfigService(
		t,
		hcc,
		protocol.HostServiceMethodHostConfigSysConfigGet,
		"custom.dynamic.denied",
		hostConfigSysConfigKeyRequest{Key: "custom.dynamic.denied"},
	)
	if denied.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected unauthorized sys_config key to be denied, got status=%d payload=%s", denied.Status, string(denied.Payload))
	}

	mismatched := invokeHostConfigSysConfigService(
		t,
		hcc,
		protocol.HostServiceMethodHostConfigSysConfigSetValue,
		"custom.dynamic.limit",
		hostConfigSysConfigSetValueRequest{Key: "custom.dynamic.denied", Value: "256"},
	)
	if mismatched.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected mismatched sys_config key to be denied, got status=%d payload=%s", mismatched.Status, string(mismatched.Payload))
	}
	if sysConfigSvc.lastSetValue == "256" {
		t.Fatal("expected mismatched sys_config set to be rejected before owner service call")
	}
}

// TestConfigureHostConfigServiceRejectsNil verifies nil hostConfig injection fails explicitly.
func TestConfigureHostConfigServiceRejectsNil(t *testing.T) {
	if _, err := NewRuntime(
		&capabilityHostServiceTestServices{},
		capregistry.NewRegistry(),
		noopTestConfigFactory{},
		nil,
		noopTestManifestFactory{},
	); err == nil {
		t.Fatal("expected nil host config service to return an error")
	}
}

// hostConfigHostCallContext builds an authorized hostConfig host service context.
func hostConfigHostCallContext(keys []string) *hostCallContext {
	return &hostCallContext{
		pluginID: "test-plugin-runtime",
		capabilities: map[string]struct{}{
			protocol.CapabilityHostConfig: {},
		},
		hostServices: []*protocol.HostServiceSpec{{
			Service: protocol.HostServiceHostConfig,
			Methods: []string{
				protocol.HostServiceMethodHostConfigGet,
				protocol.HostServiceMethodHostConfigSysConfigGet,
				protocol.HostServiceMethodHostConfigSysConfigSetValue,
				protocol.HostServiceMethodHostConfigSysConfigReset,
			},
			Keys: append([]string(nil), keys...),
		}},
	}
}

// invokeHostConfigService dispatches one hostConfig.get request.
func invokeHostConfigService(t *testing.T, hcc *hostCallContext, key string) *protocol.HostCallResponseEnvelope {
	t.Helper()
	return invokeHostConfigServiceWithPayloadKey(t, hcc, key, key)
}

// invokeHostConfigServiceWithPayloadKey dispatches one hostConfig.get request
// and allows tests to verify resourceRef/payload key mismatch handling.
func invokeHostConfigServiceWithPayloadKey(t *testing.T, hcc *hostCallContext, resourceRef string, payloadKey string) *protocol.HostCallResponseEnvelope {
	t.Helper()

	request := &protocol.HostServiceRequestEnvelope{
		Service:     protocol.HostServiceHostConfig,
		Method:      protocol.HostServiceMethodHostConfigGet,
		ResourceRef: resourceRef,
		Payload: protocol.MarshalHostServiceConfigKeyRequest(&protocol.HostServiceConfigKeyRequest{
			Key: payloadKey,
		}),
	}
	return handleHostServiceInvoke(context.Background(), withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(request))
}

// invokeHostConfigSysConfigService dispatches one JSON sys_config request.
func invokeHostConfigSysConfigService(t *testing.T, hcc *hostCallContext, method string, resourceRef string, input any) *protocol.HostCallResponseEnvelope {
	t.Helper()
	request := &protocol.HostServiceRequestEnvelope{
		Service:     protocol.HostServiceHostConfig,
		Method:      method,
		ResourceRef: resourceRef,
		Payload:     marshalCapabilityJSONRequest(t, input),
	}
	return handleHostServiceInvoke(context.Background(), withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(request))
}

// configureTrackingHostConfigService swaps the process hostConfig adapter for one test case.
func configureTrackingHostConfigService(t *testing.T, service *trackingHostConfigService) {
	t.Helper()

	bindTestHostServiceRuntime(t, withTestHostConfigService(service))
}

// setWasmHostConfigAdapter swaps the GoFrame config adapter for dynamic
// hostConfig priority tests.
func setWasmHostConfigAdapter(t *testing.T, content string) {
	t.Helper()

	adapter, err := gcfg.NewAdapterContent(content)
	if err != nil {
		t.Fatalf("create wasm host config adapter: %v", err)
	}
	originalAdapter := g.Cfg().GetAdapter()
	g.Cfg().SetAdapter(adapter)
	t.Cleanup(func() {
		g.Cfg().SetAdapter(originalAdapter)
	})
}

// insertDynamicHostConfigSysConfig inserts one platform sys_config row for
// dynamic hostConfig dispatch tests.
func insertDynamicHostConfigSysConfig(t *testing.T, ctx context.Context, key string, value string) {
	t.Helper()
	id, err := dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
		TenantId: datascope.PlatformTenantID,
		Name:     key,
		Key:      key,
		Value:    value,
		Remark:   "dynamic hostConfig test",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert dynamic hostConfig sys_config %s: %v", key, err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Id: id}).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup dynamic hostConfig sys_config %s: %v", key, cleanupErr)
		}
		if markErr := hostconfig.New().MarkRuntimeParamsChanged(ctx); markErr != nil {
			t.Fatalf("mark runtime params changed after dynamic hostConfig cleanup: %v", markErr)
		}
	})
}
