// This file tests the dynamic-plugin hostConfig host service.

package wasm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// trackingHostConfigService records host config reads.
type trackingHostConfigService struct {
	values      map[string]any
	getCalls    int
	existsCalls int
	lastKey     string
}

// Get records one host config read.
func (s *trackingHostConfigService) Get(_ context.Context, key string) (*gvar.Var, error) {
	s.getCalls++
	s.lastKey = key
	if value, ok := s.values[key]; ok {
		return gvar.New(value), nil
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
	value, err := s.Get(ctx, key)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.String(), nil
}

// Bool reads a deterministic bool value.
func (s *trackingHostConfigService) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Bool(), nil
}

// Int reads a deterministic int value.
func (s *trackingHostConfigService) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Int(), nil
}

// Duration returns a deterministic duration value.
func (s *trackingHostConfigService) Duration(context.Context, string, time.Duration) (time.Duration, error) {
	return 15 * time.Second, nil
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

// TestHandleHostServiceInvokeHostConfigReadsAuthorizedCustomSysConfig verifies
// dynamic plugins can read custom sys_config keys only after key authorization.
func TestHandleHostServiceInvokeHostConfigReadsAuthorizedCustomSysConfig(t *testing.T) {
	ctx := context.Background()
	key := fmt.Sprintf("custom.dynamic.limit.%d", time.Now().UnixNano())
	insertDynamicHostConfigSysConfig(t, ctx, key, "64")
	if err := hostconfig.New().MarkRuntimeParamsChanged(ctx); err != nil {
		t.Fatalf("mark runtime params changed: %v", err)
	}
	bindTestHostServiceRuntime(t, withTestHostConfigService(hostconfigcap.New(hostconfig.New().(hostconfigcap.RawConfigReader))))

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

// TestConfigureHostConfigServiceRejectsNil verifies nil hostConfig injection fails explicitly.
func TestConfigureHostConfigServiceRejectsNil(t *testing.T) {
	if _, err := NewRuntime(
		&capabilityHostServiceTestServices{},
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
			Methods: []string{protocol.HostServiceMethodHostConfigGet},
			Keys:    append([]string(nil), keys...),
		}},
	}
}

// invokeHostConfigService dispatches one hostConfig.get request.
func invokeHostConfigService(t *testing.T, hcc *hostCallContext, key string) *protocol.HostCallResponseEnvelope {
	t.Helper()

	request := &protocol.HostServiceRequestEnvelope{
		Service:     protocol.HostServiceHostConfig,
		Method:      protocol.HostServiceMethodHostConfigGet,
		ResourceRef: key,
		Payload: protocol.MarshalHostServiceConfigKeyRequest(&protocol.HostServiceConfigKeyRequest{
			Key: key,
		}),
	}
	return handleHostServiceInvoke(context.Background(), withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(request))
}

// configureTrackingHostConfigService swaps the process hostConfig adapter for one test case.
func configureTrackingHostConfigService(t *testing.T, service *trackingHostConfigService) {
	t.Helper()

	bindTestHostServiceRuntime(t, withTestHostConfigService(service))
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
