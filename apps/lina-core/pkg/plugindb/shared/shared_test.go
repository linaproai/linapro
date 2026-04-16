// This file tests typed plugindb enum validation and plan helper behavior.

package shared

import "testing"

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

func TestMarshalValuesJSONRejectsNonSlice(t *testing.T) {
	if _, err := MarshalValuesJSON("not-a-slice"); err == nil {
		t.Fatal("expected MarshalValuesJSON to reject non-slice input")
	}
}

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
