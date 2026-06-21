// This file verifies plugin and host configuration host-service clients.

package domainhostcall

import (
	"testing"

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
	decoded, err := protocol.UnmarshalHostServiceConfigKeyRequest(request)
	if err != nil {
		return nil, err
	}
	r.record = configHostCallRecord{
		service:     service,
		method:      method,
		resourceRef: resourceRef,
		requestKey:  decoded.Key,
	}
	return protocol.MarshalHostServiceConfigValueResponse(&protocol.HostServiceConfigValueResponse{
		Value: r.value,
		Found: r.found,
	}), nil
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
