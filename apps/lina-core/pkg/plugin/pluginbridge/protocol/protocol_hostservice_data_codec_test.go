// This file tests data host service request and response codec round trips.
package protocol

import (
	"testing"
)

// TestHostServiceDataListCodecRoundTrip verifies list requests preserve typed
// plan payloads through the codec.
func TestHostServiceDataListCodecRoundTrip(t *testing.T) {
	original := &HostServiceDataListRequest{
		PlanJSON: []byte(`{"table":"sys_plugin_node_state","action":"list"}`),
	}
	data := MarshalHostServiceDataListRequest(original)
	decoded, err := UnmarshalHostServiceDataListRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if string(decoded.PlanJSON) != string(original.PlanJSON) {
		t.Fatalf("planJson: got %s want %s", string(decoded.PlanJSON), string(original.PlanJSON))
	}
}

// TestHostServiceDataListResponseCodecRoundTrip verifies list responses
// preserve record payloads and totals through the codec.
func TestHostServiceDataListResponseCodecRoundTrip(t *testing.T) {
	original := &HostServiceDataListResponse{
		Records: [][]byte{
			[]byte(`{"id":1,"name":"alpha"}`),
			[]byte(`{"id":2,"name":"beta"}`),
		},
		Total: 2,
	}
	data := MarshalHostServiceDataListResponse(original)
	decoded, err := UnmarshalHostServiceDataListResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Total != original.Total || len(decoded.Records) != 2 {
		t.Fatalf("decoded: got %#v want %#v", decoded, original)
	}
}

// TestHostServiceDataGetCodecRoundTrip verifies get requests and responses
// preserve key and record payloads through the codec.
func TestHostServiceDataGetCodecRoundTrip(t *testing.T) {
	original := &HostServiceDataGetRequest{
		PlanJSON: []byte(`{"table":"sys_plugin_node_state","action":"get","keyJson":"NDI="}`),
	}
	data := MarshalHostServiceDataGetRequest(original)
	decoded, err := UnmarshalHostServiceDataGetRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if string(decoded.PlanJSON) != string(original.PlanJSON) {
		t.Fatalf("planJson: got %s want %s", string(decoded.PlanJSON), string(original.PlanJSON))
	}

	response := &HostServiceDataGetResponse{
		Found:      true,
		RecordJSON: []byte(`{"id":42,"name":"demo"}`),
	}
	responseData := MarshalHostServiceDataGetResponse(response)
	decodedResponse, err := UnmarshalHostServiceDataGetResponse(responseData)
	if err != nil {
		t.Fatalf("response unmarshal failed: %v", err)
	}
	if !decodedResponse.Found || string(decodedResponse.RecordJSON) != string(response.RecordJSON) {
		t.Fatalf("response: got %#v want %#v", decodedResponse, response)
	}
}

// TestHostServiceDataMutationCodecRoundTrip verifies mutation requests and
// responses preserve key and affected-row data through the codec.
func TestHostServiceDataMutationCodecRoundTrip(t *testing.T) {
	original := &HostServiceDataMutationRequest{
		KeyJSON:    []byte(`1`),
		RecordJSON: []byte(`{"status":"done"}`),
	}
	data := MarshalHostServiceDataMutationRequest(original)
	decoded, err := UnmarshalHostServiceDataMutationRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if string(decoded.KeyJSON) != "1" || string(decoded.RecordJSON) != `{"status":"done"}` {
		t.Fatalf("decoded: got %#v", decoded)
	}

	response := &HostServiceDataMutationResponse{
		AffectedRows: 1,
		KeyJSON:      []byte(`1`),
	}
	responseData := MarshalHostServiceDataMutationResponse(response)
	decodedResponse, err := UnmarshalHostServiceDataMutationResponse(responseData)
	if err != nil {
		t.Fatalf("response unmarshal failed: %v", err)
	}
	if decodedResponse.AffectedRows != 1 || string(decodedResponse.KeyJSON) != "1" {
		t.Fatalf("response: got %#v", decodedResponse)
	}
}

// TestHostServiceDataTransactionCodecRoundTrip verifies transaction requests
// and responses preserve ordered operations and per-step results.
func TestHostServiceDataTransactionCodecRoundTrip(t *testing.T) {
	original := &HostServiceDataTransactionRequest{
		Operations: []*HostServiceDataTransactionOperation{
			{
				Method:     HostServiceMethodDataCreate,
				RecordJSON: []byte(`{"name":"alpha"}`),
			},
			{
				Method:  HostServiceMethodDataDelete,
				KeyJSON: []byte(`2`),
			},
		},
	}
	data := MarshalHostServiceDataTransactionRequest(original)
	decoded, err := UnmarshalHostServiceDataTransactionRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(decoded.Operations) != 2 || decoded.Operations[0].Method != HostServiceMethodDataCreate {
		t.Fatalf("operations: got %#v", decoded.Operations)
	}

	response := &HostServiceDataTransactionResponse{
		Results: []*HostServiceDataMutationResponse{
			{AffectedRows: 1, KeyJSON: []byte(`10`)},
			{AffectedRows: 1},
		},
		AffectedRows: 2,
	}
	responseData := MarshalHostServiceDataTransactionResponse(response)
	decodedResponse, err := UnmarshalHostServiceDataTransactionResponse(responseData)
	if err != nil {
		t.Fatalf("response unmarshal failed: %v", err)
	}
	if decodedResponse.AffectedRows != 2 || len(decodedResponse.Results) != 2 {
		t.Fatalf("response: got %#v", decodedResponse)
	}
}
