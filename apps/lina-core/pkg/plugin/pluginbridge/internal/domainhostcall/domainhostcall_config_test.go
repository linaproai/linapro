// This file verifies plugin and host configuration host-service clients.

package domainhostcall

import (
	"encoding/json"
	"testing"

	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

type configHostCallRecord struct {
	service     string
	method      string
	resourceRef string
	requestKey  string
}

type configHostCallRecorder struct {
	record configHostCallRecord
	value  string
	found  bool
}

func (r *configHostCallRecorder) invoke(service string, method string, resourceRef string, _ string, request []byte) ([]byte, error) {
	record := configHostCallRecord{
		service:     service,
		method:      method,
		resourceRef: resourceRef,
	}
	if method == protocol.HostServiceMethodHostConfigGet {
		decoded, err := protocol.UnmarshalHostServiceConfigKeyRequest(request)
		if err != nil {
			return nil, err
		}
		record.requestKey = decoded.Key
	}
	r.record = record
	if method == protocol.HostServiceMethodHostConfigGet {
		return protocol.MarshalHostServiceConfigValueResponse(&protocol.HostServiceConfigValueResponse{
			Value: r.value,
			Found: r.found,
		}), nil
	}
	if method == protocol.HostServiceMethodHostConfigSysConfigGet {
		content, err := json.Marshal(&hostconfigcap.SysConfigInfo{Key: hostconfigcap.SysConfigKey(resourceRef), Value: r.value})
		if err != nil {
			return nil, err
		}
		return protocol.MarshalHostServiceJSONResponse(&protocol.HostServiceJSONResponse{Value: content}), nil
	}
	return protocol.MarshalHostServiceJSONResponse(&protocol.HostServiceJSONResponse{Value: []byte(`true`)}), nil
}

// TestHostConfigCapabilityGetUsesDefaultForMissingOrNilValues verifies the
// bridge-backed HostConfig capability mirrors source-plugin default semantics.
func TestHostConfigCapabilityGetUsesDefaultForMissingOrNilValues(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		found      bool
		defaultVal any
		wantInt    int
	}{
		{
			name:       "missing key uses default",
			found:      false,
			defaultVal: 10,
			wantInt:    10,
		},
		{
			name:       "json null uses default",
			value:      "null",
			found:      true,
			defaultVal: 20,
			wantInt:    20,
		},
		{
			name:       "existing value wins",
			value:      "30",
			found:      true,
			defaultVal: 40,
			wantInt:    30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := &configHostCallRecorder{value: tt.value, found: tt.found}
			service := HostConfigCapability(recorder.invoke)

			value, err := service.Get(t.Context(), "custom.feature.limit", tt.defaultVal)
			if err != nil {
				t.Fatalf("get host config value: %v", err)
			}
			if value == nil || value.Int() != tt.wantInt {
				t.Fatalf("host config value: got %#v want %d", value, tt.wantInt)
			}
			if recorder.record.service != protocol.HostServiceHostConfig ||
				recorder.record.method != protocol.HostServiceMethodHostConfigGet ||
				recorder.record.resourceRef != "custom.feature.limit" ||
				recorder.record.requestKey != "custom.feature.limit" {
				t.Fatalf("unexpected host config call record: %#v", recorder.record)
			}
		})
	}
}

// TestHostConfigCapabilityGetPreservesNilWhenNoDefault verifies nil defaults
// keep absent-key and nil-value reads distinguishable from concrete values.
func TestHostConfigCapabilityGetPreservesNilWhenNoDefault(t *testing.T) {
	tests := []struct {
		name  string
		value string
		found bool
	}{
		{name: "missing", found: false},
		{name: "json null", value: "null", found: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := &configHostCallRecorder{value: tt.value, found: tt.found}
			service := HostConfigCapability(recorder.invoke)

			value, err := service.Get(t.Context(), "custom.feature.limit", nil)
			if err != nil {
				t.Fatalf("get host config value: %v", err)
			}
			if value != nil {
				t.Fatalf("expected nil host config value, got %#v", value)
			}
		})
	}
}

// TestHostConfigSysConfigUsesKeyResourceRef verifies dynamic sys_config calls
// bind single-key methods to hostconfig resources.keys authorization.
func TestHostConfigSysConfigUsesKeyResourceRef(t *testing.T) {
	recorder := &configHostCallRecorder{value: "64", found: true}
	service := HostConfigCapability(recorder.invoke).SysConfig()

	value, err := service.Get(t.Context(), "custom.feature.limit")
	if err != nil {
		t.Fatalf("get sys_config value: %v", err)
	}
	if value == nil || value.Value != "64" {
		t.Fatalf("unexpected sys_config value: %#v", value)
	}
	if recorder.record.service != protocol.HostServiceHostConfig ||
		recorder.record.method != protocol.HostServiceMethodHostConfigSysConfigGet ||
		recorder.record.resourceRef != "custom.feature.limit" {
		t.Fatalf("unexpected sys_config get record: %#v", recorder.record)
	}

	if err = service.SetValue(t.Context(), "custom.feature.limit", "128"); err != nil {
		t.Fatalf("set sys_config value: %v", err)
	}
	if recorder.record.method != protocol.HostServiceMethodHostConfigSysConfigSetValue ||
		recorder.record.resourceRef != "custom.feature.limit" {
		t.Fatalf("unexpected sys_config set record: %#v", recorder.record)
	}

	if err = service.Reset(t.Context(), "custom.feature.limit"); err != nil {
		t.Fatalf("reset sys_config value: %v", err)
	}
	if recorder.record.method != protocol.HostServiceMethodHostConfigSysConfigReset ||
		recorder.record.resourceRef != "custom.feature.limit" {
		t.Fatalf("unexpected sys_config reset record: %#v", recorder.record)
	}
}
