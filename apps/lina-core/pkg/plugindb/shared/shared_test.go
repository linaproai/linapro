// This file tests typed plugindb enum validation and plan helper behavior.

package shared

import (
	"encoding/json"
	"testing"
)

// TestParseDataEnums verifies all typed plugindb enum parsers accept valid values.
func TestParseDataEnums(t *testing.T) {
	if _, err := ParseDataPlanAction("list"); err != nil {
		t.Fatalf("expected valid action, got %v", err)
	}
	if _, err := ParseDataFilterOperator("eq"); err != nil {
		t.Fatalf("expected valid operator, got %v", err)
	}
	if _, err := ParseDataOrderDirection("asc"); err != nil {
		t.Fatalf("expected valid direction, got %v", err)
	}
	if _, err := ParseDataMutationAction("create"); err != nil {
		t.Fatalf("expected valid mutation action, got %v", err)
	}
	if _, err := ParseDataAccessMode("both"); err != nil {
		t.Fatalf("expected valid access mode, got %v", err)
	}
}

// TestMarshalValuesJSONRejectsNonSlice verifies only slice inputs are accepted
// by the JSON value-list encoder.
func TestMarshalValuesJSONRejectsNonSlice(t *testing.T) {
	if _, err := MarshalValuesJSON("not-a-slice"); err == nil {
		t.Fatal("expected MarshalValuesJSON to reject non-slice input")
	}
}

// TestUnmarshalValueJSONPreservesNumbers verifies numeric values stay as
// json.Number so guest-side governed data keys avoid float64 precision loss.
func TestUnmarshalValueJSONPreservesNumbers(t *testing.T) {
	value, err := UnmarshalValueJSON([]byte(`1776132000000`))
	if err != nil {
		t.Fatalf("UnmarshalValueJSON failed: %v", err)
	}
	number, ok := value.(json.Number)
	if !ok {
		t.Fatalf("expected json.Number for numeric value, got %T", value)
	}
	if got := number.String(); got != "1776132000000" {
		t.Fatalf("expected original numeric literal, got %q", got)
	}
}

// TestMarshalValueJSONNormalizesMapRecords verifies directly constructed
// mutation maps are encoded deterministically while preserving numeric values.
func TestMarshalValueJSONNormalizesMapRecords(t *testing.T) {
	data, err := MarshalValueJSON(map[string]any{
		"title":      "demo",
		"generation": 1,
	})
	if err != nil {
		t.Fatalf("MarshalValueJSON failed: %v", err)
	}
	if string(data) != `{"generation":1,"title":"demo"}` {
		t.Fatalf("unexpected map JSON: %s", string(data))
	}

	value, err := UnmarshalValueJSON(data)
	if err != nil {
		t.Fatalf("UnmarshalValueJSON failed: %v", err)
	}
	record, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected decoded map, got %T", value)
	}
	if _, ok = record["generation"].(json.Number); !ok {
		t.Fatalf("expected generation to decode as json.Number, got %T", record["generation"])
	}
}

// TestQueryPlanJSONRoundTrip verifies query plans keep their typed fields after
// JSON marshal and unmarshal.
func TestQueryPlanJSONRoundTrip(t *testing.T) {
	plan := &DataQueryPlan{
		Table:  "sys_plugin_node_state",
		Action: DataPlanActionList,
		Filters: []*DataFilter{{
			Field:     "pluginId",
			Operator:  DataFilterOperatorEQ,
			ValueJSON: []byte(`"plugin-demo"`),
		}},
		Orders: []*DataOrder{{
			Field:     "id",
			Direction: DataOrderDirectionDESC,
		}},
		Page: &DataPagination{PageNum: 1, PageSize: 10},
	}
	data, err := MarshalQueryPlanJSON(plan)
	if err != nil {
		t.Fatalf("MarshalQueryPlanJSON failed: %v", err)
	}
	decoded, err := UnmarshalQueryPlanJSON(data)
	if err != nil {
		t.Fatalf("UnmarshalQueryPlanJSON failed: %v", err)
	}
	if decoded.Table != plan.Table || decoded.Action != plan.Action {
		t.Fatalf("unexpected plan round trip: %#v", decoded)
	}
}
