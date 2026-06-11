// This file tests historical capability JSON envelope aliases.

package protocol

import (
	"reflect"
	"testing"
)

// TestHostServiceCapabilityCodecsRoundTrip verifies JSON request and response
// aliases continue to use the generic host-service JSON envelope.
func TestHostServiceCapabilityCodecsRoundTrip(t *testing.T) {
	request := &HostServiceCapabilityJSONRequest{Value: []byte(`{"userIds":["7","8"]}`)}
	decodedRequest, err := UnmarshalHostServiceCapabilityJSONRequest(
		MarshalHostServiceCapabilityJSONRequest(request),
	)
	if err != nil {
		t.Fatalf("decode JSON request failed: %v", err)
	}
	if !reflect.DeepEqual(decodedRequest, request) {
		t.Fatalf("unexpected JSON request: %#v", decodedRequest)
	}

	response := &HostServiceCapabilityJSONResponse{Value: []byte(`{"ok":true}`)}
	decodedResponse, err := UnmarshalHostServiceCapabilityJSONResponse(
		MarshalHostServiceCapabilityJSONResponse(response),
	)
	if err != nil {
		t.Fatalf("decode JSON response failed: %v", err)
	}
	if !reflect.DeepEqual(decodedResponse, response) {
		t.Fatalf("unexpected JSON response: %#v", decodedResponse)
	}
}
