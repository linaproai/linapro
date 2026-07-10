// This file tests the dynamic-plugin manifest host service.

package wasm

import (
	"context"
	"testing"

	"lina-core/internal/service/plugin/internal/manifestresource"
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// trackingManifestFactory records plugin scopes requested by the wasm dispatcher.
type trackingManifestFactory struct {
	service               *trackingManifestService
	lastPluginID          string
	lastArtifactPlugin    string
	lastArtifactResources map[string][]byte
}

// ForPlugin returns the configured tracking manifest service for one plugin scope.
func (f *trackingManifestFactory) ForPlugin(pluginID string) manifestcap.Service {
	f.lastPluginID = pluginID
	return f.service
}

// WithArtifactResources records release-bound resources passed by the execution context.
func (f *trackingManifestFactory) WithArtifactResources(pluginID string, resources map[string][]byte) manifestresource.Factory {
	f.lastArtifactPlugin = pluginID
	f.lastArtifactResources = resources
	return f
}

// trackingManifestService records manifest reads while returning deterministic values.
type trackingManifestService struct {
	resources map[string][]byte
	getCalls  int
	lastPath  string
}

// Get records one manifest resource read.
func (s *trackingManifestService) Get(_ context.Context, path string) ([]byte, error) {
	s.getCalls++
	s.lastPath = path
	if content, ok := s.resources[path]; ok {
		return append([]byte(nil), content...), nil
	}
	return nil, nil
}

// GetMany returns multiple resources using the same opaque missing semantics
// as the real manifest service.
func (s *trackingManifestService) GetMany(ctx context.Context, input manifestcap.GetManyInput) (*manifestcap.GetManyOutput, error) {
	output := &manifestcap.GetManyOutput{Resources: []*manifestcap.ResourceContent{}}
	for _, path := range input.Paths {
		content, err := s.Get(ctx, path)
		if err != nil {
			return nil, err
		}
		if len(content) == 0 {
			output.MissingPaths = append(output.MissingPaths, path)
			continue
		}
		output.Resources = append(output.Resources, &manifestcap.ResourceContent{Path: path, Body: content})
	}
	return output, nil
}

// List returns deterministic resource metadata for paths under the requested prefix.
func (s *trackingManifestService) List(_ context.Context, input manifestcap.ListInput) (*manifestcap.ListOutput, error) {
	resources := []*manifestcap.Resource{}
	limit := input.Limit
	if limit <= 0 {
		limit = manifestcap.DefaultListLimit
	}
	for path, content := range s.resources {
		if input.Prefix != "" && len(path) >= len(input.Prefix) && path[:len(input.Prefix)] != input.Prefix {
			continue
		}
		if input.Prefix != "" && len(path) < len(input.Prefix) {
			continue
		}
		resources = append(resources, &manifestcap.Resource{Path: path, Size: int64(len(content))})
		if len(resources) >= limit {
			break
		}
	}
	return &manifestcap.ListOutput{Resources: resources, Limit: limit}, nil
}

// Exists reports whether one resource exists.
func (s *trackingManifestService) Exists(ctx context.Context, path string) (bool, error) {
	content, err := s.Get(ctx, path)
	return len(content) > 0, err
}

// Scan is unused in wasm dispatcher tests.
func (s *trackingManifestService) Scan(context.Context, string, string, any) error { return nil }

// TestHandleHostServiceInvokeManifestReadsAuthorizedPath verifies manifest.get
// reads authorized plugin-scoped manifest resources.
func TestHandleHostServiceInvokeManifestReadsAuthorizedPath(t *testing.T) {
	manifestSvc := &trackingManifestService{resources: map[string][]byte{
		"metadata.yaml":              []byte("name: demo\n"),
		"config/config.example.yaml": []byte("feature:\n  enabled: false\n"),
		"sql/001-schema.sql":         []byte("CREATE TABLE plugin_demo(id bigint);\n"),
		"i18n/zh-CN/plugin.json":     []byte(`{"plugin.demo":"demo"}`),
	}}
	factory := configureTrackingManifestFactory(t, manifestSvc)

	hcc := manifestHostCallContext([]string{
		"metadata.yaml",
		"config/config.example.yaml",
		"sql/001-schema.sql",
		"i18n/zh-CN/plugin.json",
	})
	for path, expected := range manifestSvc.resources {
		response := invokeManifestHostService(t, hcc, path)
		payload := decodeManifestResponse(t, response)
		if !payload.Found || string(payload.Body) != string(expected) {
			t.Fatalf("expected %s payload %q, got %#v", path, string(expected), payload)
		}
	}
	if factory.lastPluginID != "test-plugin-manifest" {
		t.Fatalf("expected manifest factory to be scoped to plugin, got %q", factory.lastPluginID)
	}
	if manifestSvc.getCalls != len(manifestSvc.resources) {
		t.Fatalf("expected manifest get call, got calls=%d path=%q", manifestSvc.getCalls, manifestSvc.lastPath)
	}
}

// TestHandleHostServiceInvokeManifestRejectsUnauthorizedPath verifies
// resources.paths are enforced before dispatch.
func TestHandleHostServiceInvokeManifestRejectsUnauthorizedPath(t *testing.T) {
	configureTrackingManifestFactory(t, &trackingManifestService{resources: map[string][]byte{
		"metadata.yaml":              []byte("name: demo\n"),
		"config/config.example.yaml": []byte("feature:\n  enabled: false\n"),
	}})

	response := invokeManifestHostService(t, manifestHostCallContext([]string{"metadata.yaml"}), "config/config.example.yaml")
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected unauthorized manifest path to be denied, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeManifestAllowsGlobPath verifies glob paths can
// authorize manifest resources.
func TestHandleHostServiceInvokeManifestAllowsGlobPath(t *testing.T) {
	configureTrackingManifestFactory(t, &trackingManifestService{resources: map[string][]byte{
		"resources/policy.yaml": []byte("enabled: true\n"),
	}})

	response := invokeManifestHostService(t, manifestHostCallContext([]string{"resources/*.yaml"}), "resources/policy.yaml")
	payload := decodeManifestResponse(t, response)
	if !payload.Found || string(payload.Body) != "enabled: true\n" {
		t.Fatalf("expected globbed resource payload, got %#v", payload)
	}
}

// TestHandleHostServiceInvokeManifestBindsArtifactResources verifies active
// release manifest resources are passed to the scoped factory for each execution.
func TestHandleHostServiceInvokeManifestBindsArtifactResources(t *testing.T) {
	manifestSvc := &trackingManifestService{resources: map[string][]byte{
		"config/config.example.yaml": []byte("feature:\n  enabled: false\n"),
	}}
	factory := configureTrackingManifestFactory(t, manifestSvc)
	hcc := manifestHostCallContext([]string{"config/config.example.yaml"})
	hcc.artifactManifestResources = map[string][]byte{
		"config/config.example.yaml": []byte("feature:\n  enabled: true\n"),
	}

	response := invokeManifestHostService(t, hcc, "config/config.example.yaml")
	payload := decodeManifestResponse(t, response)
	if !payload.Found || string(payload.Body) != "feature:\n  enabled: false\n" {
		t.Fatalf("expected manifest payload, got %#v", payload)
	}
	if factory.lastArtifactPlugin != "test-plugin-manifest" || string(factory.lastArtifactResources["config/config.example.yaml"]) != "feature:\n  enabled: true\n" {
		t.Fatalf("expected artifact manifest resources binding, got plugin=%q resources=%#v", factory.lastArtifactPlugin, factory.lastArtifactResources)
	}
}

// TestHandleHostServiceInvokeManifestJSONMethods verifies get_many and list use
// the JSON envelope while enforcing path authorization inside the dispatcher.
func TestHandleHostServiceInvokeManifestJSONMethods(t *testing.T) {
	manifestSvc := &trackingManifestService{resources: map[string][]byte{
		"config/a.yaml":   []byte("a: true\n"),
		"config/b.yaml":   []byte("b: true\n"),
		"private/c.yaml":  []byte("c: true\n"),
		"metadata.yaml":   []byte("name: demo\n"),
		"resources/x.txt": []byte("x"),
	}}
	configureTrackingManifestFactory(t, manifestSvc)
	hcc := manifestHostCallContext([]string{"config/"})

	getManyResponse := invokeManifestJSONHostService(
		t,
		hcc,
		protocol.HostServiceMethodManifestGetMany,
		"config/",
		manifestcap.GetManyInput{Paths: []string{"config/a.yaml", "config/missing.yaml"}},
	)
	if getManyResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("get_many: expected success, got status=%d payload=%s", getManyResponse.Status, string(getManyResponse.Payload))
	}
	var getMany manifestcap.GetManyOutput
	decodeCapabilityJSONResponse(t, getManyResponse.Payload, &getMany)
	if len(getMany.Resources) != 1 || getMany.Resources[0].Path != "config/a.yaml" || string(getMany.Resources[0].Body) != "a: true\n" {
		t.Fatalf("unexpected get_many resources: %#v", getMany.Resources)
	}
	if len(getMany.MissingPaths) != 1 || getMany.MissingPaths[0] != "config/missing.yaml" {
		t.Fatalf("unexpected get_many missing paths: %#v", getMany.MissingPaths)
	}

	listResponse := invokeManifestJSONHostService(
		t,
		hcc,
		protocol.HostServiceMethodManifestList,
		"config/.manifest-list-probe",
		manifestcap.ListInput{Prefix: "config/", Limit: 5},
	)
	if listResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("list: expected success, got status=%d payload=%s", listResponse.Status, string(listResponse.Payload))
	}
	var list manifestcap.ListOutput
	decodeCapabilityJSONResponse(t, listResponse.Payload, &list)
	if len(list.Resources) == 0 {
		t.Fatalf("expected list resources under authorized prefix")
	}
	for _, resource := range list.Resources {
		if resource.Path[:len("config/")] != "config/" {
			t.Fatalf("unexpected resource outside config prefix: %#v", resource)
		}
	}

	deniedResponse := invokeManifestJSONHostService(
		t,
		hcc,
		protocol.HostServiceMethodManifestGetMany,
		"config/a.yaml",
		manifestcap.GetManyInput{Paths: []string{"config/a.yaml", "private/c.yaml"}},
	)
	if deniedResponse.Status != protocol.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request for unauthorized embedded manifest path, got status=%d payload=%s", deniedResponse.Status, string(deniedResponse.Payload))
	}
}

// TestHandleHostServiceInvokeManifestRootListRequiresBroadGrant verifies root
// listing only succeeds with an explicit broad path grant.
func TestHandleHostServiceInvokeManifestRootListRequiresBroadGrant(t *testing.T) {
	configureTrackingManifestFactory(t, &trackingManifestService{resources: map[string][]byte{
		"metadata.yaml": []byte("name: demo\n"),
	}})

	denied := invokeManifestJSONHostService(
		t,
		manifestHostCallContext([]string{"config/"}),
		protocol.HostServiceMethodManifestList,
		".manifest-list-probe",
		manifestcap.ListInput{Limit: 5},
	)
	if denied.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected root list without broad grant to be denied, got status=%d payload=%s", denied.Status, string(denied.Payload))
	}

	allowed := invokeManifestJSONHostService(
		t,
		manifestHostCallContext([]string{"*"}),
		protocol.HostServiceMethodManifestList,
		".manifest-list-probe",
		manifestcap.ListInput{Limit: 5},
	)
	if allowed.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected root list with broad grant to succeed, got status=%d payload=%s", allowed.Status, string(allowed.Payload))
	}
}

// TestConfigureManifestHostServiceRejectsNil verifies nil manifest injection fails explicitly.
func TestConfigureManifestHostServiceRejectsNil(t *testing.T) {
	if _, err := NewRuntime(
		&capabilityHostServiceTestServices{},
		capregistry.NewRegistry(),
		noopTestConfigFactory{},
		noopTestHostConfigService{},
		nil,
	); err == nil {
		t.Fatal("expected nil manifest host service factory to return an error")
	}
}

// manifestHostCallContext builds an authorized manifest host service context.
func manifestHostCallContext(paths []string) *hostCallContext {
	return &hostCallContext{
		pluginID: "test-plugin-manifest",
		capabilities: map[string]struct{}{
			protocol.CapabilityManifest: {},
		},
		hostServices: []*protocol.HostServiceSpec{{
			Service: protocol.HostServiceManifest,
			Methods: []string{
				protocol.HostServiceMethodManifestGet,
				protocol.HostServiceMethodManifestGetMany,
				protocol.HostServiceMethodManifestList,
			},
			Paths: append([]string(nil), paths...),
		}},
	}
}

// invokeManifestHostService dispatches one manifest.get request.
func invokeManifestHostService(t *testing.T, hcc *hostCallContext, path string) *protocol.HostCallResponseEnvelope {
	t.Helper()

	request := &protocol.HostServiceRequestEnvelope{
		Service:     protocol.HostServiceManifest,
		Method:      protocol.HostServiceMethodManifestGet,
		ResourceRef: path,
		Payload: protocol.MarshalHostServiceManifestGetRequest(&protocol.HostServiceManifestGetRequest{
			Path: path,
		}),
	}
	return handleHostServiceInvoke(context.Background(), withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(request))
}

// invokeManifestJSONHostService dispatches one manifest JSON-envelope request.
func invokeManifestJSONHostService(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	resourceRef string,
	input any,
) *protocol.HostCallResponseEnvelope {
	t.Helper()

	request := &protocol.HostServiceRequestEnvelope{
		Service:     protocol.HostServiceManifest,
		Method:      method,
		ResourceRef: resourceRef,
		Payload:     marshalCapabilityJSONRequest(t, input),
	}
	return handleHostServiceInvoke(context.Background(), withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(request))
}

// decodeManifestResponse verifies success and decodes one manifest response.
func decodeManifestResponse(
	t *testing.T,
	response *protocol.HostCallResponseEnvelope,
) *protocol.HostServiceManifestGetResponse {
	t.Helper()

	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected manifest host service success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	payload, err := protocol.UnmarshalHostServiceManifestGetResponse(response.Payload)
	if err != nil {
		t.Fatalf("expected manifest response decode to succeed, got error: %v", err)
	}
	return payload
}

// configureTrackingManifestFactory swaps the process manifest factory for one test case.
func configureTrackingManifestFactory(t *testing.T, service *trackingManifestService) *trackingManifestFactory {
	t.Helper()

	factory := &trackingManifestFactory{service: service}
	bindTestHostServiceRuntime(t, withTestManifestFactory(factory))
	return factory
}
