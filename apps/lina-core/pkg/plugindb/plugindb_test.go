// This file tests plugindb guest-side query-plan and transaction builders.

package plugindb

import (
	"testing"

	"lina-core/pkg/plugindb/shared"
)

func TestTransactionRejectsMultipleTables(t *testing.T) {
	tx := &Tx{}
	_ = tx.Table("sys_plugin_node_state")
	_ = tx.Table("sys_plugin")
	if tx.err == nil {
		t.Fatal("expected transaction to reject multiple tables")
	}
}

func TestQueryBuilderBuildsTypedPlan(t *testing.T) {
	query := Open().
		Table("sys_plugin_node_state").
		Fields("id", "pluginId", "currentState").
		WhereEq("pluginId", "plugin-demo").
		WhereIn("currentState", []string{"pending", "running"}).
		WhereLike("nodeKey", "demo-").
		OrderDesc("id").
		Page(2, 20)
	if query.err != nil {
		t.Fatalf("expected query builder to succeed, got %v", query.err)
	}
	if len(query.plan.Fields) != 3 {
		t.Fatalf("unexpected fields: %#v", query.plan.Fields)
	}
	if len(query.plan.Filters) != 3 {
		t.Fatalf("unexpected filters: %#v", query.plan.Filters)
	}
	if query.plan.Filters[1].Operator != shared.DataFilterOperatorIN {
		t.Fatalf("unexpected filter operator: %#v", query.plan.Filters[1])
	}
	if len(query.plan.Orders) != 1 || query.plan.Orders[0].Direction != shared.DataOrderDirectionDESC {
		t.Fatalf("unexpected orders: %#v", query.plan.Orders)
	}
	if query.plan.Page == nil || query.plan.Page.PageNum != 2 || query.plan.Page.PageSize != 20 {
		t.Fatalf("unexpected page: %#v", query.plan.Page)
	}
}
