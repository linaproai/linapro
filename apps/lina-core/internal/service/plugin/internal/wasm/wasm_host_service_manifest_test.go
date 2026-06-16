// This file tests the dynamic-plugin manifest host service.

package wasm

import (
	"context"
	"testing"

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
func (f *trackingManifestFactory) WithArtifactResources(pluginID string, resources map[string][]byte) manifestcap.ServiceFactory {
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

// TestConfigureManifestHostServiceRejectsNil verifies nil manifest injection fails explicitly.
func TestConfigureManifestHostServiceRejectsNil(t *testing.T) {
	if _, err := NewRuntime(
		&capabilityHostServiceTestServices{},
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
			Methods: []string{protocol.HostServiceMethodManifestGet},
			Paths:   append([]string(nil), paths...),
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
