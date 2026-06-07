// This file verifies user capability query assembly inside the usercap component.

package usercap

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
)

// TestSearchUsersCountsWithoutProjectionFields verifies the search count query
// does not reuse projection columns, which keeps PostgreSQL count SQL valid.
func TestSearchUsersCountsWithoutProjectionFields(t *testing.T) {
	ctx := context.Background()

	sqls, err := gdb.CatchSQL(ctx, func(sqlCtx context.Context) error {
		_, searchErr := New(nil).SearchUsers(sqlCtx, capmodel.CapabilityContext{}, capabilityusercap.SearchInput{
			Page: capmodel.PageRequest{PageNum: 1, PageSize: 10},
		})
		return searchErr
	})
	if err != nil {
		t.Fatalf("render search users SQL: %v", err)
	}

	combinedSQL := strings.Join(sqls, "\n")
	normalizedSQL := strings.ReplaceAll(combinedSQL, "`", `"`)
	if strings.Contains(normalizedSQL, `COUNT("id","tenant_id","username","nickname","avatar","status")`) {
		t.Fatalf("count query reused projection fields: %s", combinedSQL)
	}
	if !strings.Contains(strings.ToUpper(normalizedSQL), "COUNT") {
		t.Fatalf("expected rendered SQL to include count query, got: %s", combinedSQL)
	}
}
