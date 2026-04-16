// This file tests data host service request and response codec round trips.
package pluginbridge

import (
	"testing"

	"lina-core/pkg/plugindb/shared"
)

func TestHostServiceDataListCodecRoundTrip(t *testing.T) {
	original := &HostServiceDataListRequest{
		Filters: map[string]string{
			"pluginId": "plugin-demo",
			"status":   "enabled",
		},
		PageNum:  2,
		PageSize: 20,
		PlanJSON: []byte(`{"table":"sys_plugin_node_state","action":"list"}`),
	}
	data := MarshalHostServiceDataListRequest(original)
	decoded, err := UnmarshalHostServiceDataListRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.PageNum != original.PageNum || decoded.PageSize != original.PageSize {
		t.Fatalf("page: got %#v want %#v", decoded, original)
	}
	if len(decoded.Filters) != 2 || decoded.Filters["pluginId"] != "plugin-demo" {
		t.Fatalf("filters: got %#v", decoded.Filters)
	}
	if string(decoded.PlanJSON) != string(original.PlanJSON) {
		t.Fatalf("planJson: got %s want %s", string(decoded.PlanJSON), string(original.PlanJSON))
	}
}

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

func TestHostServiceDataGetCodecRoundTrip(t *testing.T) {
	original := &HostServiceDataGetRequest{
		KeyJSON:  []byte(`42`),
		PlanJSON: []byte(`{"table":"sys_plugin_node_state","action":"get"}`),
	}
	data := MarshalHostServiceDataGetRequest(original)
	decoded, err := UnmarshalHostServiceDataGetRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if string(decoded.KeyJSON) != "42" {
		t.Fatalf("keyJson: got %s", string(decoded.KeyJSON))
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

func TestDecodeHostServiceDataPlanHelpers(t *testing.T) {
	planJSON, err := shared.MarshalQueryPlanJSON(&shared.DataQueryPlan{
		Table:  "sys_plugin_node_state",
		Action: shared.DataPlanActionList,
		Page:   &shared.DataPagination{PageNum: 1, PageSize: 10},
	})
	if err != nil {
		t.Fatalf("marshal query plan failed: %v", err)
	}

	listPlan, err := DecodeHostServiceDataListPlan(&HostServiceDataListRequest{PlanJSON: planJSON})
	if err != nil {
		t.Fatalf("decode list plan failed: %v", err)
	}
	if listPlan == nil || listPlan.Action != shared.DataPlanActionList {
		t.Fatalf("unexpected list plan: %#v", listPlan)
	}

	getPlan, err := DecodeHostServiceDataGetPlan(&HostServiceDataGetRequest{PlanJSON: planJSON})
	if err != nil {
		t.Fatalf("decode get plan failed: %v", err)
	}
	if getPlan == nil || getPlan.Table != "sys_plugin_node_state" {
		t.Fatalf("unexpected get plan: %#v", getPlan)
	}
}

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
