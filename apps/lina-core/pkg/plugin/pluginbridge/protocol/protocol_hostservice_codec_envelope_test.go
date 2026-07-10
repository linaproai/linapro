// This file tests generic structured host-service envelope codecs.

package protocol

import "testing"

// TestHostServiceRequestEnvelopeRoundTrip verifies host service request
// envelopes preserve owner, service, version, method, table, and payload data.
func TestHostServiceRequestEnvelopeRoundTrip(t *testing.T) {
	original := &HostServiceRequestEnvelope{
		Owner:   "linapro-ai-core",
		Service: HostServiceData,
		Version: "v1",
		Method:  HostServiceMethodDataGet,
		Table:   "sys_plugin_node_state",
		Payload: MarshalHostServiceDataGetRequest(&HostServiceDataGetRequest{
			PlanJSON: []byte(`{"table":"sys_plugin_node_state","action":"get","keyJson":"MQ=="}`),
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
	if decoded.Owner != original.Owner {
		t.Errorf("owner: got %q, want %q", decoded.Owner, original.Owner)
	}
	if decoded.Version != original.Version {
		t.Errorf("version: got %q, want %q", decoded.Version, original.Version)
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

// TestHostServiceJSONEnvelopeRoundTrip verifies ordinary domain JSON payloads
// can use one transport-only request and response envelope.
func TestHostServiceJSONEnvelopeRoundTrip(t *testing.T) {
	request := &HostServiceJSONRequest{Value: []byte(`{"keyword":"admin"}`)}
	decodedRequest, err := UnmarshalHostServiceJSONRequest(MarshalHostServiceJSONRequest(request))
	if err != nil {
		t.Fatalf("decode JSON request failed: %v", err)
	}
	if string(decodedRequest.Value) != string(request.Value) {
		t.Fatalf("unexpected JSON request value: %s", decodedRequest.Value)
	}

	response := &HostServiceJSONResponse{Value: []byte(`{"items":[{"id":"1"}]}`)}
	decodedResponse, err := UnmarshalHostServiceJSONResponse(MarshalHostServiceJSONResponse(response))
	if err != nil {
		t.Fatalf("decode JSON response failed: %v", err)
	}
	if string(decodedResponse.Value) != string(response.Value) {
		t.Fatalf("unexpected JSON response value: %s", decodedResponse.Value)
	}
}
