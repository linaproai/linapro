// This file tests the dynamic-plugin read-only config host service.

package wasm

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/pkg/pluginbridge"
)

// TestHandleHostServiceInvokeConfigReadsValues verifies dynamic plugins can
// read arbitrary host configuration values through the config host service.
func TestHandleHostServiceInvokeConfigReadsValues(t *testing.T) {
	setWasmConfigAdapter(t, `
monitor:
  interval: 45s
  retentionMultiplier: 8
feature:
  enabled: true
  retries: 3
`)

	hcc := configHostCallContext()

	getResponse := invokeConfigHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodConfigGet,
		"monitor.interval",
	)
	getPayload := decodeConfigResponse(t, getResponse)
	if !getPayload.Found || getPayload.Value != `"45s"` {
		t.Fatalf("expected monitor.interval JSON value, got %#v", getPayload)
	}

	stringResponse := invokeConfigHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodConfigString,
		"monitor.interval",
	)
	stringPayload := decodeConfigResponse(t, stringResponse)
	if !stringPayload.Found || stringPayload.Value != "45s" {
		t.Fatalf("expected monitor.interval string value, got %#v", stringPayload)
	}

	boolResponse := invokeConfigHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodConfigBool,
		"feature.enabled",
	)
	boolPayload := decodeConfigResponse(t, boolResponse)
	if !boolPayload.Found || boolPayload.Value != "true" {
		t.Fatalf("expected feature.enabled bool value, got %#v", boolPayload)
	}

	intResponse := invokeConfigHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodConfigInt,
		"feature.retries",
	)
	intPayload := decodeConfigResponse(t, intResponse)
	if !intPayload.Found || intPayload.Value != "3" {
		t.Fatalf("expected feature.retries int value, got %#v", intPayload)
	}

	durationResponse := invokeConfigHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodConfigDuration,
		"monitor.interval",
	)
	durationPayload := decodeConfigResponse(t, durationResponse)
	if !durationPayload.Found || durationPayload.Value != "45s" {
		t.Fatalf("expected monitor.interval duration value, got %#v", durationPayload)
	}

	existsResponse := invokeConfigHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodConfigExists,
		"feature.retries",
	)
	existsPayload := decodeConfigResponse(t, existsResponse)
	if !existsPayload.Found {
		t.Fatalf("expected feature.retries exists response to be found, got %#v", existsPayload)
	}
}

// TestHandleHostServiceInvokeConfigReadsFullSnapshot verifies empty key reads
// the full GoFrame configuration snapshot.
func TestHandleHostServiceInvokeConfigReadsFullSnapshot(t *testing.T) {
	setWasmConfigAdapter(t, `
custom:
  name: demo
`)

	response := invokeConfigHostService(
		t,
		configHostCallContext(),
		pluginbridge.HostServiceMethodConfigGet,
		"",
	)
	payload := decodeConfigResponse(t, response)
	if !payload.Found {
		t.Fatal("expected full config snapshot to be found")
	}
	if !strings.Contains(payload.Value, `"custom"`) || !strings.Contains(payload.Value, `"demo"`) {
		t.Fatalf("expected full config snapshot JSON to include custom value, got %s", payload.Value)
	}
}

// TestHandleHostServiceInvokeConfigMissingKey verifies missing keys return found=false.
func TestHandleHostServiceInvokeConfigMissingKey(t *testing.T) {
	setWasmConfigAdapter(t, `
custom:
  name: demo
`)

	response := invokeConfigHostService(
		t,
		configHostCallContext(),
		pluginbridge.HostServiceMethodConfigGet,
		"custom.missing",
	)
	payload := decodeConfigResponse(t, response)
	if payload.Found {
		t.Fatalf("expected missing key to return found=false, got %#v", payload)
	}
}

// TestHandleHostServiceInvokeConfigRejectsUnsupportedMethod verifies dynamic
// config host service declarations and calls remain limited to read-only methods.
func TestHandleHostServiceInvokeConfigRejectsUnsupportedMethod(t *testing.T) {
	setWasmConfigAdapter(t, `
custom:
  name: demo
`)

	response := invokeConfigHostService(
		t,
		configHostCallContext(),
		"set",
		"custom.name",
	)
	if response.Status != pluginbridge.HostCallStatusNotFound {
		t.Fatalf(
			"expected unsupported config method to be rejected, got status=%d payload=%s",
			response.Status,
			string(response.Payload),
		)
	}
}

// configHostCallContext builds an authorized config host service context.
func configHostCallContext() *hostCallContext {
	return &hostCallContext{
		pluginID: "test-plugin-config",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityConfig: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceConfig,
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
) *pluginbridge.HostCallResponseEnvelope {
	t.Helper()

	request := &pluginbridge.HostServiceRequestEnvelope{
		Service: pluginbridge.HostServiceConfig,
		Method:  method,
		Payload: pluginbridge.MarshalHostServiceConfigKeyRequest(&pluginbridge.HostServiceConfigKeyRequest{
			Key: key,
		}),
	}
	return handleHostServiceInvoke(context.Background(), hcc, pluginbridge.MarshalHostServiceRequestEnvelope(request))
}

// decodeConfigResponse verifies success and decodes one config host service response.
func decodeConfigResponse(
	t *testing.T,
	response *pluginbridge.HostCallResponseEnvelope,
) *pluginbridge.HostServiceConfigValueResponse {
	t.Helper()

	if response.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("expected config host service success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	payload, err := pluginbridge.UnmarshalHostServiceConfigValueResponse(response.Payload)
	if err != nil {
		t.Fatalf("expected config response decode to succeed, got error: %v", err)
	}
	return payload
}

// setWasmConfigAdapter swaps the process config adapter for one test case.
func setWasmConfigAdapter(t *testing.T, content string) {
	t.Helper()

	adapter, err := gcfg.NewAdapterContent(content)
	if err != nil {
		t.Fatalf("create content adapter: %v", err)
	}

	originalAdapter := g.Cfg().GetAdapter()
	g.Cfg().SetAdapter(adapter)

	t.Cleanup(func() {
		g.Cfg().SetAdapter(originalAdapter)
	})
}
