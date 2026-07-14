// This file tests runtime host service methods including log/state/info
// dispatch and capability validation.

package wasm

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/manifestresource"
	"lina-core/internal/service/plugin/internal/pluginconfig"
	"lina-core/pkg/dialect"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// createPluginStateTableSQL provisions the plugin runtime state table required
// by runtime host-service state lifecycle tests.
const createPluginStateTableSQL = `
CREATE TABLE IF NOT EXISTS sys_plugin_state (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    plugin_id   VARCHAR(64) NOT NULL DEFAULT '',
    tenant_id   INT NOT NULL DEFAULT 0,
    state_key   VARCHAR(255) NOT NULL DEFAULT '',
    state_value TEXT,
    enabled     BOOL NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_plugin_state_plugin_tenant_key ON sys_plugin_state (plugin_id, tenant_id, state_key);
`

// TestHandleHostServiceInvokeRuntimeStateLifecycle verifies runtime state
// get/set/delete calls persist and remove plugin-scoped values correctly.
func TestHandleHostServiceInvokeRuntimeStateLifecycle(t *testing.T) {
	ctx := context.Background()
	ensureRuntimeStateTable(t, ctx)

	hcc := &hostCallContext{
		pluginID: "test-plugin-runtime-state",
		capabilities: map[string]struct{}{
			protocol.CapabilityRuntime: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{
					protocol.HostServiceMethodRuntimeStateGet,
					protocol.HostServiceMethodRuntimeStateSet,
					protocol.HostServiceMethodRuntimeStateDelete,
				},
			},
		},
	}
	cleanupRuntimeStateKey(t, ctx, hcc.pluginID, "demo")
	t.Cleanup(func() {
		cleanupRuntimeStateKey(t, ctx, hcc.pluginID, "demo")
	})

	setResponse := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateSet,
		protocol.MarshalHostCallStateSetRequest(&protocol.HostCallStateSetRequest{
			Key:   "demo",
			Value: "value-1",
		}),
	)
	if setResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected state.set success, got status=%d payload=%s", setResponse.Status, string(setResponse.Payload))
	}

	getResponse := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateGet,
		protocol.MarshalHostCallStateGetRequest(&protocol.HostCallStateGetRequest{Key: "demo"}),
	)
	if getResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected state.get success, got status=%d payload=%s", getResponse.Status, string(getResponse.Payload))
	}
	getPayload, err := protocol.UnmarshalHostCallStateGetResponse(getResponse.Payload)
	if err != nil {
		t.Fatalf("expected state.get payload decode to succeed, got error: %v", err)
	}
	if !getPayload.Found || getPayload.Value != "value-1" {
		t.Fatalf("expected stored state value to round-trip, got %#v", getPayload)
	}

	updateResponse := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateSet,
		protocol.MarshalHostCallStateSetRequest(&protocol.HostCallStateSetRequest{
			Key:   "demo",
			Value: "value-2",
		}),
	)
	if updateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected second state.set success, got status=%d payload=%s", updateResponse.Status, string(updateResponse.Payload))
	}
	getUpdatedResponse := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateGet,
		protocol.MarshalHostCallStateGetRequest(&protocol.HostCallStateGetRequest{Key: "demo"}),
	)
	if getUpdatedResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected updated state.get success, got status=%d payload=%s", getUpdatedResponse.Status, string(getUpdatedResponse.Payload))
	}
	getUpdatedPayload, err := protocol.UnmarshalHostCallStateGetResponse(getUpdatedResponse.Payload)
	if err != nil {
		t.Fatalf("expected updated state.get payload decode to succeed, got error: %v", err)
	}
	if !getUpdatedPayload.Found || getUpdatedPayload.Value != "value-2" {
		t.Fatalf("expected updated state value to round-trip, got %#v", getUpdatedPayload)
	}

	deleteResponse := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateDelete,
		protocol.MarshalHostCallStateDeleteRequest(&protocol.HostCallStateDeleteRequest{Key: "demo"}),
	)
	if deleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected state.delete success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}
}

// TestHandleHostServiceInvokeRuntimeInfoNowAndNode verifies runtime info
// methods return non-empty host metadata payloads.
func TestHandleHostServiceInvokeRuntimeInfoNowAndNode(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin-runtime-info",
		capabilities: map[string]struct{}{
			protocol.CapabilityRuntime: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{
					protocol.HostServiceMethodRuntimeInfoNow,
					protocol.HostServiceMethodRuntimeInfoNode,
				},
			},
		},
	}

	var (
		beforeMillis = time.Now().Add(-1 * time.Second).UnixMilli()
		nowResponse  = invokeRuntimeHostService(t, hcc, protocol.HostServiceMethodRuntimeInfoNow, nil)
		afterMillis  = time.Now().Add(1 * time.Second).UnixMilli()
	)
	if nowResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected info.now success, got status=%d payload=%s", nowResponse.Status, string(nowResponse.Payload))
	}
	nowPayload, err := protocol.UnmarshalHostServiceValueResponse(nowResponse.Payload)
	if err != nil {
		t.Fatalf("expected info.now payload decode to succeed, got error: %v", err)
	}
	if strings.TrimSpace(nowPayload.Value) == "" {
		t.Fatal("expected info.now value to be non-empty")
	}
	nowMillis, err := strconv.ParseInt(nowPayload.Value, 10, 64)
	if err != nil {
		t.Fatalf("expected info.now value to be Unix milliseconds, got %q: %v", nowPayload.Value, err)
	}
	if nowMillis < beforeMillis || nowMillis > afterMillis {
		t.Fatalf("expected info.now value within test window, got %d outside [%d,%d]", nowMillis, beforeMillis, afterMillis)
	}

	nodeResponse := invokeRuntimeHostService(t, hcc, protocol.HostServiceMethodRuntimeInfoNode, nil)
	if nodeResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected info.node success, got status=%d payload=%s", nodeResponse.Status, string(nodeResponse.Payload))
	}
	nodePayload, err := protocol.UnmarshalHostServiceValueResponse(nodeResponse.Payload)
	if err != nil {
		t.Fatalf("expected info.node payload decode to succeed, got error: %v", err)
	}
	if strings.TrimSpace(nodePayload.Value) == "" {
		t.Fatal("expected info.node value to be non-empty")
	}
}

// TestHandleHostServiceInvokeRuntimeStateBatchLifecycle verifies runtime state
// batch methods persist, read, and delete plugin-scoped values.
func TestHandleHostServiceInvokeRuntimeStateBatchLifecycle(t *testing.T) {
	ctx := context.Background()
	ensureRuntimeStateTable(t, ctx)

	hcc := &hostCallContext{
		pluginID: "test-plugin-runtime-state-batch",
		capabilities: map[string]struct{}{
			protocol.CapabilityRuntime: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{
					protocol.HostServiceMethodRuntimeStateGetMany,
					protocol.HostServiceMethodRuntimeStateSetMany,
					protocol.HostServiceMethodRuntimeStateDeleteMany,
				},
			},
		},
	}
	for _, key := range []string{"one", "two", "missing"} {
		cleanupRuntimeStateKey(t, ctx, hcc.pluginID, key)
	}
	t.Cleanup(func() {
		for _, key := range []string{"one", "two", "missing"} {
			cleanupRuntimeStateKey(t, ctx, hcc.pluginID, key)
		}
	})

	setResponse := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateSetMany,
		marshalCapabilityJSONRequest(t, runtimeStateSetManyRequest{Values: map[string]string{
			"one": "1",
			"two": "2",
		}}),
	)
	if setResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("state.set_many: expected success, got status=%d payload=%s", setResponse.Status, string(setResponse.Payload))
	}

	getResponse := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateGetMany,
		marshalCapabilityJSONRequest(t, runtimeStateGetManyRequest{Keys: []string{"one", "two", "missing"}}),
	)
	if getResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("state.get_many: expected success, got status=%d payload=%s", getResponse.Status, string(getResponse.Payload))
	}
	var getPayload runtimeStateGetManyResponse
	decodeCapabilityJSONResponse(t, getResponse.Payload, &getPayload)
	if getPayload.Values["one"] != "1" || getPayload.Values["two"] != "2" {
		t.Fatalf("unexpected state.get_many values: %#v", getPayload.Values)
	}
	if len(getPayload.MissingKeys) != 1 || getPayload.MissingKeys[0] != "missing" {
		t.Fatalf("unexpected missing keys: %#v", getPayload.MissingKeys)
	}

	deleteResponse := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateDeleteMany,
		marshalCapabilityJSONRequest(t, runtimeStateDeleteManyRequest{Keys: []string{"one", "missing"}}),
	)
	if deleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("state.delete_many: expected success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}

	getAfterDelete := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateGetMany,
		marshalCapabilityJSONRequest(t, runtimeStateGetManyRequest{Keys: []string{"one", "two"}}),
	)
	if getAfterDelete.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("state.get_many after delete: expected success, got status=%d payload=%s", getAfterDelete.Status, string(getAfterDelete.Payload))
	}
	var afterDelete runtimeStateGetManyResponse
	decodeCapabilityJSONResponse(t, getAfterDelete.Payload, &afterDelete)
	if _, exists := afterDelete.Values["one"]; exists || afterDelete.Values["two"] != "2" {
		t.Fatalf("unexpected values after delete: %#v", afterDelete.Values)
	}
}

// invokeRuntimeHostService dispatches one runtime host-service request and
// returns the raw response envelope for assertions.
func invokeRuntimeHostService(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *protocol.HostCallResponseEnvelope {
	t.Helper()

	request := &protocol.HostServiceRequestEnvelope{
		Service: protocol.HostServiceRuntime,
		Method:  method,
		Payload: payload,
	}
	return handleHostServiceInvoke(context.Background(), withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(request))
}

func ensureRuntimeStateTable(t *testing.T, ctx context.Context) {
	t.Helper()
	for _, statement := range dialect.SplitSQLStatements(createPluginStateTableSQL) {
		if _, err := g.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("expected plugin state table to be created, got error: %v\nSQL:\n%s", err, statement)
		}
	}
}

// cleanupRuntimeStateKey deletes one plugin runtime state row so lifecycle
// tests can run repeatedly without leftover state.
func cleanupRuntimeStateKey(t *testing.T, ctx context.Context, pluginID string, key string) {
	t.Helper()
	if _, err := dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{PluginId: pluginID, StateKey: key}).
		Delete(); err != nil {
		t.Fatalf("failed to cleanup runtime state key %s/%s: %v", pluginID, key, err)
	}
}

var testHostServiceRuntimes sync.Map

type testHostServiceRuntimeOption func(*hostServiceRuntime)

func withTestDomainServices(services capability.Services) testHostServiceRuntimeOption {
	return func(runtime *hostServiceRuntime) {
		runtime.domainServices = services
	}
}

func withTestOwnerCapabilities(registry *capregistry.Registry) testHostServiceRuntimeOption {
	return func(runtime *hostServiceRuntime) {
		runtime.ownerCapabilities = registry
	}
}

func withTestConfigFactory(factory pluginconfig.Factory) testHostServiceRuntimeOption {
	return func(runtime *hostServiceRuntime) {
		runtime.pluginConfigFactory = factory
	}
}

func withTestHostConfigService(service hostconfigcap.Service) testHostServiceRuntimeOption {
	return func(runtime *hostServiceRuntime) {
		runtime.hostConfigService = service
	}
}

func withTestManifestFactory(factory manifestresource.Factory) testHostServiceRuntimeOption {
	return func(runtime *hostServiceRuntime) {
		runtime.manifestFactory = factory
	}
}

func bindTestHostServiceRuntime(t *testing.T, opts ...testHostServiceRuntimeOption) *hostServiceRuntime {
	t.Helper()
	runtime := newTestHostServiceRuntime(opts...)
	testHostServiceRuntimes.Store(t.Name(), runtime)
	t.Cleanup(func() {
		testHostServiceRuntimes.Delete(t.Name())
	})
	return runtime
}

func withTestHostCallRuntime(t *testing.T, hcc *hostCallContext) *hostCallContext {
	t.Helper()
	if hcc == nil || hcc.runtime != nil {
		return hcc
	}
	if runtime, ok := testHostServiceRuntimes.Load(t.Name()); ok {
		hcc.runtime = runtime.(*hostServiceRuntime)
		return hcc
	}
	hcc.runtime = bindTestHostServiceRuntime(t)
	return hcc
}

func newTestHostServiceRuntime(opts ...testHostServiceRuntimeOption) *hostServiceRuntime {
	runtime := &hostServiceRuntime{
		domainServices:      &capabilityHostServiceTestServices{},
		ownerCapabilities:   capregistry.NewRegistry(),
		pluginConfigFactory: noopTestConfigFactory{},
		hostConfigService:   noopTestHostConfigService{},
		manifestFactory:     noopTestManifestFactory{},
		storageUploads:      newStorageUploadSessions(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(runtime)
		}
	}
	return runtime
}

type noopTestConfigFactory struct{}

func (noopTestConfigFactory) ForPlugin(string) plugincap.ConfigService {
	return noopTestConfigService{}
}

func (f noopTestConfigFactory) WithArtifactConfig(string, []byte) pluginconfig.Factory {
	return f
}

type noopTestConfigService struct{}

func (noopTestConfigService) Get(_ context.Context, _ string, defaultValue any) (*gvar.Var, error) {
	if defaultValue != nil {
		return gvar.New(defaultValue), nil
	}
	return nil, nil
}
func (noopTestConfigService) Exists(context.Context, string) (bool, error) { return false, nil }
func (noopTestConfigService) Scan(context.Context, string, any) error      { return nil }
func (noopTestConfigService) String(context.Context, string, string) (string, error) {
	return "", nil
}
func (noopTestConfigService) Bool(context.Context, string, bool) (bool, error) { return false, nil }
func (noopTestConfigService) Int(context.Context, string, int) (int, error)    { return 0, nil }
func (noopTestConfigService) Duration(context.Context, string, time.Duration) (time.Duration, error) {
	return 0, nil
}

type noopTestHostConfigService struct{}

func (noopTestHostConfigService) Get(context.Context, string, any) (*gvar.Var, error) {
	return nil, nil
}
func (noopTestHostConfigService) Exists(context.Context, string) (bool, error) { return false, nil }
func (noopTestHostConfigService) String(context.Context, string, string) (string, error) {
	return "", nil
}
func (noopTestHostConfigService) Bool(context.Context, string, bool) (bool, error) {
	return false, nil
}
func (noopTestHostConfigService) Int(context.Context, string, int) (int, error) { return 0, nil }
func (noopTestHostConfigService) Duration(context.Context, string, time.Duration) (time.Duration, error) {
	return 0, nil
}
func (noopTestHostConfigService) SysConfig() hostconfigcap.SysConfigService {
	return noopTestSysConfigService{}
}

type noopTestSysConfigService struct{}

func (s noopTestSysConfigService) Get(ctx context.Context, key hostconfigcap.SysConfigKey) (*hostconfigcap.SysConfigInfo, error) {
	result, err := s.BatchGet(ctx, []hostconfigcap.SysConfigKey{key})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[key], nil
}

func (noopTestSysConfigService) BatchGet(context.Context, []hostconfigcap.SysConfigKey) (*capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey], error) {
	return &capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey]{Items: map[hostconfigcap.SysConfigKey]*hostconfigcap.SysConfigInfo{}}, nil
}

func (noopTestSysConfigService) List(context.Context, hostconfigcap.ListSysConfigInput) (*capmodel.PageResult[*hostconfigcap.SysConfigInfo], error) {
	return &capmodel.PageResult[*hostconfigcap.SysConfigInfo]{Items: []*hostconfigcap.SysConfigInfo{}}, nil
}

func (noopTestSysConfigService) SetValue(context.Context, hostconfigcap.SysConfigKey, string, *hostconfigcap.SetSysConfigValueOptions) error {
	return nil
}

func (noopTestSysConfigService) BatchSetValue(context.Context, []hostconfigcap.SetSysConfigValueItem, *hostconfigcap.SetSysConfigValueOptions) error {
	return nil
}

func (noopTestSysConfigService) Reset(context.Context, hostconfigcap.SysConfigKey) error {
	return nil
}

func (noopTestSysConfigService) EnsureVisible(context.Context, []hostconfigcap.SysConfigKey) error {
	return nil
}

type noopTestManifestFactory struct{}

func (noopTestManifestFactory) ForPlugin(string) manifestcap.Service {
	return noopTestManifestService{}
}

func (f noopTestManifestFactory) WithArtifactResources(string, map[string][]byte) manifestresource.Factory {
	return f
}

type noopTestManifestService struct{}

func (noopTestManifestService) Get(context.Context, string) ([]byte, error) { return nil, nil }
func (noopTestManifestService) GetMany(context.Context, manifestcap.GetManyInput) (*manifestcap.GetManyOutput, error) {
	return &manifestcap.GetManyOutput{}, nil
}
func (noopTestManifestService) List(context.Context, manifestcap.ListInput) (*manifestcap.ListOutput, error) {
	return &manifestcap.ListOutput{}, nil
}
func (noopTestManifestService) Exists(context.Context, string) (bool, error) {
	return false, nil
}
func (noopTestManifestService) Scan(context.Context, string, string, any) error { return nil }
