// This file tests generic structured host-service envelope codecs.

package hostservice

import "testing"

// TestHostServiceRequestEnvelopeRoundTrip verifies host service request
// envelopes preserve service, method, table, and payload data.
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

// TestHostServiceValueResponseRoundTrip verifies simple string-valued host
// service responses round-trip through the codec.
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
