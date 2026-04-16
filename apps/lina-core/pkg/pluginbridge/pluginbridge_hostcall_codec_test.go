// This file tests host call request and response codec round trips.

package pluginbridge

import (
	"testing"
)

func TestHostCallResponseEnvelopeRoundTrip(t *testing.T) {
	original := &HostCallResponseEnvelope{
		Status:  HostCallStatusCapabilityDenied,
		Payload: []byte("missing host:runtime capability"),
	}
	data := MarshalHostCallResponse(original)
	decoded, err := UnmarshalHostCallResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Status != original.Status {
		t.Errorf("status: got %d, want %d", decoded.Status, original.Status)
	}
	if string(decoded.Payload) != string(original.Payload) {
		t.Errorf("payload: got %q, want %q", decoded.Payload, original.Payload)
	}
}

func TestHostCallSuccessResponseRoundTrip(t *testing.T) {
	original := NewHostCallEmptySuccessResponse()
	data := MarshalHostCallResponse(original)
	decoded, err := UnmarshalHostCallResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Status != HostCallStatusSuccess {
		t.Errorf("status: got %d, want %d", decoded.Status, HostCallStatusSuccess)
	}
}

func TestHostCallLogRequestRoundTrip(t *testing.T) {
	original := &HostCallLogRequest{
		Level:   LogLevelWarning,
		Message: "test warning message",
		Fields:  map[string]string{"key1": "val1", "key2": "val2"},
	}
	data := MarshalHostCallLogRequest(original)
	decoded, err := UnmarshalHostCallLogRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Level != original.Level {
		t.Errorf("level: got %d, want %d", decoded.Level, original.Level)
	}
	if decoded.Message != original.Message {
		t.Errorf("message: got %q, want %q", decoded.Message, original.Message)
	}
	if len(decoded.Fields) != 2 || decoded.Fields["key1"] != "val1" {
		t.Errorf("fields: got %v, want %v", decoded.Fields, original.Fields)
	}
}

func TestHostServiceRequestEnvelopeRoundTrip(t *testing.T) {
	original := &HostServiceRequestEnvelope{
		Service: HostServiceData,
		Method:  HostServiceMethodDataGet,
		Table:   "sys_plugin_node_state",
		Payload: MarshalHostServiceDataGetRequest(&HostServiceDataGetRequest{
			KeyJSON: []byte("1"),
		}),
	}
	data := MarshalHostServiceRequestEnvelope(original)
	decoded, err := UnmarshalHostServiceRequestEnvelope(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Service != original.Service {
		t.Errorf("service: got %q, want %q", decoded.Service, original.Service)
	}
	if decoded.Method != original.Method {
		t.Errorf("method: got %q, want %q", decoded.Method, original.Method)
	}
	if decoded.Table != original.Table {
		t.Errorf("table: got %q, want %q", decoded.Table, original.Table)
	}
	if string(decoded.Payload) != string(original.Payload) {
		t.Errorf("payload: got %v, want %v", decoded.Payload, original.Payload)
	}
}

func TestHostServiceValueResponseRoundTrip(t *testing.T) {
	original := &HostServiceValueResponse{Value: "node-a"}
	data := MarshalHostServiceValueResponse(original)
	decoded, err := UnmarshalHostServiceValueResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Value != original.Value {
		t.Errorf("value: got %q, want %q", decoded.Value, original.Value)
	}
}

func TestHostCallStateGetRequestRoundTrip(t *testing.T) {
	original := &HostCallStateGetRequest{Key: "counter"}
	data := MarshalHostCallStateGetRequest(original)
	decoded, err := UnmarshalHostCallStateGetRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Key != original.Key {
		t.Errorf("key: got %q, want %q", decoded.Key, original.Key)
	}
}

func TestHostCallStateGetResponseRoundTrip(t *testing.T) {
	original := &HostCallStateGetResponse{Value: "42", Found: true}
	data := MarshalHostCallStateGetResponse(original)
	decoded, err := UnmarshalHostCallStateGetResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Value != original.Value {
		t.Errorf("value: got %q, want %q", decoded.Value, original.Value)
	}
	if decoded.Found != original.Found {
		t.Errorf("found: got %v, want %v", decoded.Found, original.Found)
	}
}

func TestHostCallStateSetRequestRoundTrip(t *testing.T) {
	original := &HostCallStateSetRequest{Key: "counter", Value: "43"}
	data := MarshalHostCallStateSetRequest(original)
	decoded, err := UnmarshalHostCallStateSetRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Key != original.Key {
		t.Errorf("key: got %q, want %q", decoded.Key, original.Key)
	}
	if decoded.Value != original.Value {
		t.Errorf("value: got %q, want %q", decoded.Value, original.Value)
	}
}

func TestHostCallStateDeleteRequestRoundTrip(t *testing.T) {
	original := &HostCallStateDeleteRequest{Key: "counter"}
	data := MarshalHostCallStateDeleteRequest(original)
	decoded, err := UnmarshalHostCallStateDeleteRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Key != original.Key {
		t.Errorf("key: got %q, want %q", decoded.Key, original.Key)
	}
}
